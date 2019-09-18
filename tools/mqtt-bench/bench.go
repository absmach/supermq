// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package bench

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cisco/senml"
)

const fixedSize = 491

// Benchmark - main benchmarking function
func Benchmark(cfg Config) {
	checkConnection(cfg.MQTT.Broker.URL, 1)

	var wg sync.WaitGroup
	var err error
	subsResults := map[string](*[]float64){}
	var caByte []byte
	if cfg.MQTT.TLS.MTLS {
		caFile, err := os.Open(cfg.MQTT.TLS.CA)
		defer caFile.Close()
		if err != nil {
			fmt.Println(err)
		}
		caByte, _ = ioutil.ReadAll(caFile)
	}

	mf := mainflux{}
	if _, err := toml.DecodeFile(cfg.Mf.ConnFile, &mf); err != nil {
		log.Fatalf("Cannot load Mainflux connections config %s \nuse tools/provision to create file", cfg.Mf.ConnFile)
	}

	resCh := make(chan *runResults)
	donePub := make(chan bool)
	finishedPub := make(chan bool)
	finishedSub := make(chan bool)

	resR := make(chan *map[string](*[]float64))
	startStamp := time.Now()

	n := len(mf.Channels)
	var cert tls.Certificate

	handler := getBytePayload(cfg.MQTT.Message.Size)

	// Subscribers
	for i := 0; i < cfg.Test.Subs; i++ {
		mfChan := mf.Channels[i%n]
		mfThing := mf.Things[i%n]

		if cfg.MQTT.TLS.MTLS {
			cert, err = tls.X509KeyPair([]byte(mfThing.MTLSCert), []byte(mfThing.MTLSKey))
			if err != nil {
				log.Fatal(err)
			}
		}
		c := makeClient(i, cfg, mfChan, mfThing, startStamp, caByte, cert, handler)

		go c.subscribe(&wg, cfg.Test.Count*cfg.Test.Pubs, &donePub, resR)
	}

	wg.Wait()

	start := time.Now()

	// Publishers
	for i := 0; i < cfg.Test.Pubs; i++ {
		mfChan := mf.Channels[i%n]
		mfThing := mf.Things[i%n]

		if cfg.MQTT.TLS.MTLS {
			cert, err = tls.X509KeyPair([]byte(mfThing.MTLSCert), []byte(mfThing.MTLSKey))
			if err != nil {
				log.Fatal(err)
			}
		}
		c := makeClient(i, cfg, mfChan, mfThing, startStamp, caByte, cert, handler)

		go c.publish(resCh)
	}

	// Collect the results
	var results []*runResults
	if cfg.Test.Pubs > 0 {
		results = make([]*runResults, cfg.Test.Pubs)
	}

	// Wait for publishers to finish
	go func() {
		for i := 0; i < cfg.Test.Pubs; i++ {
			results[i] = <-resCh
		}
		finishedPub <- true
	}()

	go func() {
		for i := 0; i < cfg.Test.Subs; i++ {
			r := <-resR
			for k, v := range *r {
				subsResults[k] = v
			}
		}
		finishedSub <- true
	}()

	<-finishedPub
	// Send signal to subscribers that all the publishers are done
	for i := 0; i < cfg.Test.Subs; i++ {
		donePub <- true
	}

	<-finishedSub

	totalTime := time.Now().Sub(start)
	totals := calculateTotalResults(results, totalTime, subsResults)
	if totals == nil {
		return
	}

	// Print sats
	printResults(results, totals, cfg.MQTT.Message.Format, cfg.Log.Quiet)
}

func getSenMLTimeStamp() senml.SenMLRecord {
	t := (float64)(time.Now().UnixNano())
	timeStamp := senml.SenMLRecord{
		BaseName: "pub-2019-08-31T12:38:25.139715762+02:00-57",
		Value:    &t,
	}
	return timeStamp
}

func buildSenML(size int, payload string) senml.SenML {
	timeStamp := getSenMLTimeStamp()

	tsByte, err := json.Marshal(timeStamp)
	if err != nil || len(payload) == 0 {
		log.Fatalf("Failed to create test message")
	}

	sml := senml.SenMLRecord{}
	err = json.Unmarshal([]byte(payload), &sml)
	if err != nil {
		log.Fatalf("Cannot unmarshal payload")
	}

	msgByte, err := json.Marshal(sml)
	if err != nil {
		log.Fatalf("Failed to create test message")
	}

	// How many records to make message long sz bytes
	n := (size-len(tsByte))/len(msgByte) + 1
	if size < len(tsByte) {
		n = 1
	}

	records := make([]senml.SenMLRecord, n)
	records[0] = timeStamp
	t := time.Now().UnixNano()
	for i := 1; i < n; i++ {
		// Timestamp for each record when saving to db
		sml.Time = float64(t + int64(i))
		records[i] = sml
	}

	s := senml.SenML{
		Records: records,
	}

	return s
}

func getBytePayload(size int) handler {
	s := size - fixedSize
	if s <= 0 {
		s = 0
	}
	var b = make([]byte, s)
	rand.Read(b)
	return func(m *message) ([]byte, error) {
		m.Payload = b
		m.Sent = time.Now()
		return json.Marshal(m)
	}
}

func getSenMLPayload(cid string, time float64, getSenML func() *senml.SenML) ([]byte, error) {
	s := *getSenML()
	s.Records[0].Value = &time
	s.Records[0].BaseName = cid
	payload, err := senml.Encode(s, senml.JSON, senml.OutputOptions{})
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func makeClient(i int, cfg Config, mfChan mfChannel, mfThing mfThing, start time.Time, caCert []byte, clientCert tls.Certificate, h handler) *Client {
	return &Client{
		ID:         strconv.Itoa(i),
		BrokerURL:  cfg.MQTT.Broker.URL,
		BrokerUser: mfThing.ThingID,
		BrokerPass: mfThing.ThingKey,
		MsgTopic:   fmt.Sprintf("channels/%s/messages/%d/test", mfChan.ChannelID, start.UnixNano()),
		MsgSize:    cfg.MQTT.Message.Size,
		MsgCount:   cfg.Test.Count,
		MsgQoS:     byte(cfg.MQTT.Message.QoS),
		Quiet:      cfg.Log.Quiet,
		MTLS:       cfg.MQTT.TLS.MTLS,
		SkipTLSVer: cfg.MQTT.TLS.SkipTLSVer,
		CA:         caCert,
		timeout:    cfg.MQTT.Timeout,
		ClientCert: clientCert,
		Retain:     cfg.MQTT.Message.Retain,
		SendMsg:    h,
	}
}
