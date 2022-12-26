package things

import (
	"io"
	"io/ioutil"

	"github.com/mainflux/mainflux/pkg/errors"
	gogrpc "google.golang.org/grpc"

	"github.com/mainflux/mainflux"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"

	grpcClient "github.com/mainflux/mainflux/internal/client/grpc"

	"github.com/mainflux/mainflux/internal/env"
)

const envThingsAuthGrpcPrefix = "MF_THINGS_AUTH_GRPC_"

var (
	errGrpcConfig = errors.New("failed to load grpc configuration")
)

func Setup(envPrefix, jaegerURL string) (mainflux.ThingsServiceClient, *gogrpc.ClientConn, io.Closer, string, error) {
	config := grpcClient.Config{}
	if err := env.Parse(&config, env.Options{Prefix: envThingsAuthGrpcPrefix, AltPrefix: envPrefix}); err != nil {
		return nil, nil, ioutil.NopCloser(nil), "", errors.Wrap(errGrpcConfig, err)
	}

	grpcClient, tracer, tracerCloser, secure, err := grpcClient.Setup(config, "things", jaegerURL)
	if err != nil {
		return nil, nil, ioutil.NopCloser(nil), "", err
	}

	message := "without TLS"
	if secure {
		message = "with TLS"
	}
	return thingsapi.NewClient(grpcClient, tracer, config.Timeout), grpcClient, tracerCloser, message, nil
}
