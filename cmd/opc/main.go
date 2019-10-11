package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	r "github.com/go-redis/redis"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/opc"
	"github.com/mainflux/mainflux/opc/api"
	pub "github.com/mainflux/mainflux/opc/nats"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux/opc/redis"
	nats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	defHTTPPort     = "8180"
	defOPCServerURI = "opc.tcp://opcua.rocks:4840"
	defOPCNodeID    = "ns=0;i=2256"
	defNatsURL      = nats.DefaultURL
	defLogLevel     = "debug"
	defESURL        = "localhost:6379"
	defESPass       = ""
	defESDB         = "0"
	defInstanceName = "opc"
	defRouteMapURL  = "localhost:6379"
	defRouteMapPass = ""
	defRouteMapDB   = "0"

	envHTTPPort     = "MF_OPC_ADAPTER_HTTP_PORT"
	envOPCServerURI = "MF_OPC_ADAPTER_SERVER_URI"
	envOPCNodeID    = "MF_OPC_ADAPTER_NODE_ID"
	envNatsURL      = "MF_NATS_URL"
	envLogLevel     = "MF_LORA_ADAPTER_LOG_LEVEL"
	envESURL        = "MF_THINGS_ES_URL"
	envESPass       = "MF_THINGS_ES_PASS"
	envESDB         = "MF_THINGS_ES_DB"
	envInstanceName = "MF_OPC_ADAPTER_INSTANCE_NAME"
	envRouteMapURL  = "MF_OPC_ADAPTER_ROUTEMAP_URL"
	envRouteMapPass = "MF_OPC_ADAPTER_ROUTEMAP_PASS"
	envRouteMapDB   = "MF_OPC_ADAPTER_ROUTEMAP_DB"

	loraServerTopic = "application/+/device/+/rx"

	thingsRMPrefix   = "thing"
	channelsRMPrefix = "channel"
)

type config struct {
	httpPort     string
	OPCServerURI string
	OPCNodeID    string
	natsURL      string
	logLevel     string
	esURL        string
	esPass       string
	esDB         string
	instanceName string
	routeMapURL  string
	routeMapPass string
	routeMapDB   string
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	natsConn := connectToNATS(cfg.natsURL, logger)
	defer natsConn.Close()

	rmConn := connectToRedis(cfg.routeMapURL, cfg.routeMapPass, cfg.routeMapDB, logger)
	defer rmConn.Close()

	esConn := connectToRedis(cfg.esURL, cfg.esPass, cfg.esDB, logger)
	defer esConn.Close()

	publisher := pub.NewMessagePublisher(natsConn)

	thingRM := newRouteMapRepositoy(rmConn, thingsRMPrefix, logger)
	chanRM := newRouteMapRepositoy(rmConn, channelsRMPrefix, logger)

	svc := opc.New(publisher, thingRM, chanRM)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "lora_adapter",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "lora_adapter",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	go subscribeToOpcServer(svc, cfg.OPCServerURI, cfg.OPCNodeID, logger)
	go subscribeToThingsES(svc, esConn, cfg.instanceName, logger)

	errs := make(chan error, 2)

	go startHTTPServer(cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("LoRa adapter terminated: %s", err))
}

func loadConfig() config {
	return config{
		httpPort:     mainflux.Env(envHTTPPort, defHTTPPort),
		OPCServerURI: mainflux.Env(envOPCServerURI, defOPCServerURI),
		OPCNodeID:    mainflux.Env(envOPCNodeID, defOPCNodeID),
		natsURL:      mainflux.Env(envNatsURL, defNatsURL),
		logLevel:     mainflux.Env(envLogLevel, defLogLevel),
		esURL:        mainflux.Env(envESURL, defESURL),
		esPass:       mainflux.Env(envESPass, defESPass),
		esDB:         mainflux.Env(envESDB, defESDB),
		instanceName: mainflux.Env(envInstanceName, defInstanceName),
		routeMapURL:  mainflux.Env(envRouteMapURL, defRouteMapURL),
		routeMapPass: mainflux.Env(envRouteMapPass, defRouteMapPass),
		routeMapDB:   mainflux.Env(envRouteMapDB, defRouteMapDB),
	}
}

func connectToNATS(url string, logger logger.Logger) *nats.Conn {
	conn, err := nats.Connect(url)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}

	logger.Info("Connected to NATS")
	return conn
}

func connectToRedis(redisURL, redisPass, redisDB string, logger logger.Logger) *r.Client {
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to redis: %s", err))
		os.Exit(1)
	}

	return r.NewClient(&r.Options{
		Addr:     redisURL,
		Password: redisPass,
		DB:       db,
	})
}

func subscribeToOpcServer(svc opc.Service, uri, nid string, logger logger.Logger) {
	var (
		policy   = "" // Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto
		mode     = "" // Modes: None, Sign, SignAndEncrypt. Default: auto
		certFile = ""
		keyFile  = ""
	)

	ctx := context.Background()

	endpoints, err := opcua.GetEndpoints(uri)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to fetch OPC server endpoints: %s", err.Error()))
	}

	ep := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	if ep == nil {
		logger.Error("Failed to find suitable endpoint")
	}

	logger.Info(fmt.Sprintf("OPC-UA server URI: %s", ep.SecurityPolicyURI))

	opts := []opcua.Option{
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
		opcua.CertificateFile(certFile),
		opcua.PrivateKeyFile(keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(ep.EndpointURL, opts...)
	if errC := c.Connect(ctx); err != nil {
		logger.Error(errC.Error())
	}
	defer c.Close()

	sub, err := c.Subscribe(&opcua.SubscriptionParameters{
		Interval: 2000 * time.Millisecond,
	})
	if err != nil {
		logger.Error(err.Error())
	}
	defer sub.Cancel()
	logger.Info(fmt.Sprintf("Created subscription with id %v", sub.SubscriptionID))

	nodeID, err := ua.ParseNodeID(nid)
	if err != nil {
		logger.Error(err.Error())
	}

	// arbitrary client handle for the monitoring item
	handle := uint32(42)
	miCreateRequest := opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, handle)
	res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		logger.Error(err.Error())
	}

	go sub.Run(ctx)

	// read from subscription's notification channel until ctx is cancelled
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-sub.Notifs:
			if res.Error != nil {
				logger.Error(res.Error.Error())
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					data := item.Value.Value.Value()

					mData, err := json.Marshal(data)
					if err != nil {
						logger.Error("Failed to Marshal JSON data")
						continue
					}

					log.Printf("MonitoredItem with client handle %v = %s", item.ClientHandle, string(mData))

					// Publish on Mainflux NATS broker
					msg := opc.Message{
						Namespace: "0",
						ID:        "2256",
						Data:      mData,
					}
					svc.Publish(ctx, "", msg)
				}

			default:
				logger.Info(fmt.Sprintf("what's this publish result? %T", res.Value))
			}
		}
	}
}

func subscribeToThingsES(svc opc.Service, client *r.Client, consumer string, logger logger.Logger) {
	eventStore := redis.NewEventStore(svc, client, consumer, logger)
	if err := eventStore.Subscribe("mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Failed to subscribe to Redis event sourcing: %s", err))
	}
}

func newRouteMapRepositoy(client *r.Client, prefix string, logger logger.Logger) opc.RouteMapRepository {
	logger.Info("Connected to Redis Route map")
	return redis.NewRouteMapRepository(client, prefix)
}

func startHTTPServer(cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	logger.Info(fmt.Sprintf("Lora-adapter service started, exposed port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, api.MakeHandler())
}
