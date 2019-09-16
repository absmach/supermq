// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package bench

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mat "gonum.org/v1/gonum/mat"
	stat "gonum.org/v1/gonum/stat"
)

// Client - represents mqtt client
type Client struct {
	ID         string
	BrokerURL  string
	BrokerUser string
	BrokerPass string
	MsgTopic   string
	MsgSize    int
	MsgCount   int
	MsgQoS     byte
	Quiet      bool
	timeout    int
	mqttClient *mqtt.Client
	MTLS       bool
	SkipTLSVer bool
	Retain     bool
	CA         []byte
	ClientCert tls.Certificate
	ClientKey  *rsa.PrivateKey
	SendMsg    handler
}

type message struct {
	ID        string
	Topic     string
	QoS       byte
	Payload   []byte
	Sent      time.Time
	Delivered time.Time
	Error     bool
}

type handler func(*message) ([]byte, error)

func (c *Client) RcvMsg(res *chan *map[string](*[]float64), total int, donePub *chan bool) mqtt.MessageHandler {
	subsResults := make(map[string](*[]float64), 1)
	i := 1
	a := []float64{}

	return func(cl mqtt.Client, msg mqtt.Message) {
		go func() {
			for {
				<-*donePub
				if cl.IsConnected() {
					cl.Disconnect(100)
				}
				time.Sleep(2 * time.Second)
				subsResults[c.MsgTopic] = &a
				*res <- &subsResults
			}
		}()

		var m message
		if err := json.Unmarshal(msg.Payload(), &m); err != nil {
			return
		}
		arrival := float64(time.Now().UnixNano())
		timeSent := float64(m.Sent.UnixNano())

		a = append(a, (arrival - timeSent))
		i++
		if i == total {
			cl.Disconnect(0)
			subsResults[c.MsgTopic] = &a
			*res <- &subsResults
		}
	}
}

func (c *Client) subscribe(wg *sync.WaitGroup, tot int, donePub *chan bool, res chan *map[string](*[]float64)) {
	c.ID = fmt.Sprintf("sub-%v-%v", time.Now().Format(time.RFC3339Nano), c.ID)
	if c.connect() != nil {
		wg.Done()
		log.Printf("Client %v failed connecting to the broker\n", c.ID)
	}

	token := (*c.mqttClient).Subscribe(c.MsgTopic, c.MsgQoS, c.RcvMsg(&res, tot, donePub))

	token.Wait()
}

func (c *Client) publish(r chan *runResults) {
	res := &runResults{}
	times := make([]*float64, c.MsgCount)

	start := time.Now()
	if c.connect() != nil {
		flushMessages := make([]message, c.MsgCount)
		for i, m := range flushMessages {
			m.Error = true
			times[i] = calcMsgRes(&m, res)
		}
		r <- calcRes(res, start, arr(times))
		return
	}
	if !c.Quiet {
		log.Printf("Client %v is connected to the broker %v\n", c.ID, c.BrokerURL)
	}
	// Use a single message, since it's going to be sent sequentially.
	m := message{
		Topic: c.MsgTopic,
		QoS:   c.MsgQoS,
		ID:    c.ID,
	}

	for i := 0; i < c.MsgCount; i++ {
		payload, err := c.SendMsg(&m)
		if err != nil {
			log.Fatalf("Failed to marshal payload - %s", err.Error())
		}
		token := (*c.mqttClient).Publish(m.Topic, m.QoS, c.Retain, payload)
		if !token.WaitTimeout(time.Second*time.Duration(c.timeout)) || token.Error() != nil {
			m.Error = true
			times[i] = calcMsgRes(&m, res)
			continue
		}

		m.Delivered = time.Now()
		m.Error = false
		times[i] = calcMsgRes(&m, res)

		if !c.Quiet && i > 0 && i%100 == 0 {
			log.Printf("Client %v published %v messages and keeps publishing...\n", c.ID, i)
		}
	}

	r <- calcRes(res, start, arr(times))
}

func (c *Client) connect() error {
	opts := mqtt.NewClientOptions().
		AddBroker(c.BrokerURL).
		SetClientID(c.ID).
		SetCleanSession(false).
		SetAutoReconnect(false).
		SetOnConnectHandler(c.subConnected).
		SetConnectionLostHandler(c.subLost)

	if c.BrokerUser != "" && c.BrokerPass != "" {
		opts.SetUsername(c.BrokerUser)
		opts.SetPassword(c.BrokerPass)
	}

	if c.MTLS {
		cfg := &tls.Config{
			InsecureSkipVerify: c.SkipTLSVer,
		}

		if c.CA != nil {
			cfg.RootCAs = x509.NewCertPool()
			cfg.RootCAs.AppendCertsFromPEM(c.CA)
		}
		if c.ClientCert.Certificate != nil {
			cfg.Certificates = []tls.Certificate{c.ClientCert}
		}

		cfg.BuildNameToCertificate()
		opts.SetTLSConfig(cfg)
		opts.SetProtocolVersion(4)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()

	c.mqttClient = &client

	if token.Error() != nil {
		log.Printf("Client %v had error connecting to the broker: %s\n", c.ID, token.Error().Error())
		return token.Error()
	}

	return nil
}

func checkConnection(broker string, timeoutSecs int) {
	s := strings.Split(broker, ":")
	if len(s) != 3 {
		log.Fatalf("Wrong host address format")
	}

	network := s[0]
	host := strings.Trim(s[1], "/")
	port := s[2]

	log.Println("Testing connection...")
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), time.Duration(timeoutSecs)*time.Second)
	conClose := func() {
		if conn != nil {
			log.Println("Closing testing connection...")
			conn.Close()
		}
	}

	defer conClose()
	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		log.Fatalf("Timeout error: %s\n", err.Error())
	}

	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}

	log.Printf("Connection to %s://%s:%s looks OK\n", network, host, port)
}

func calcMsgRes(m *message, res *runResults) *float64 {
	if m.Error {
		res.Failures++
		return nil
	}
	res.Successes++
	diff := float64(m.Delivered.Sub(m.Sent).Nanoseconds() / 1000) // in microseconds
	return &diff
}

func calcRes(r *runResults, start time.Time, times []float64) *runResults {
	duration := time.Now().Sub(start)
	timeMatrix := mat.NewDense(1, len(times), times)
	r.MsgTimeMin = mat.Min(timeMatrix)
	r.MsgTimeMax = mat.Max(timeMatrix)
	r.MsgTimeMean = stat.Mean(times, nil)
	r.MsgTimeStd = stat.StdDev(times, nil)
	r.RunTime = duration.Seconds()
	r.MsgsPerSec = float64(r.Successes) / duration.Seconds()
	return r
}

func arr(a []*float64) []float64 {
	ret := []float64{}
	for _, v := range a {
		if v != nil {
			ret = append(ret, *v)
		}
	}
	if len(ret) == 0 {
		ret = append(ret, 0)
	}
	return ret
}

func (c *Client) subConnected(client mqtt.Client) {
	if !c.Quiet {
		log.Printf("Client %v is connected to the broker %v\n", c.ID, c.BrokerURL)
	}
}

func (c *Client) subLost(client mqtt.Client, reason error) {
	log.Printf("Client %v had lost connection to the broker: %s\n", c.ID, reason.Error())
}

func (c *Client) pubConnected(client mqtt.Client) {

}
