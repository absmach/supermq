package auth

import (
	"io"
	"io/ioutil"

	"github.com/mainflux/mainflux/pkg/errors"
	gogrpc "google.golang.org/grpc"

	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"

	grpcClient "github.com/mainflux/mainflux/internal/client/grpc"

	"github.com/mainflux/mainflux/internal/env"
)

const envAuthGrpcPrefix = "MF_AUTH_GRPC_"

var (
	errGrpcConfig = errors.New("failed to load grpc configuration")
)

func Setup(envPrefix, jaegerURL string) (mainflux.AuthServiceClient, *gogrpc.ClientConn, io.Closer, string, error) {
	config := grpcClient.Config{}
	if err := env.Parse(&config, env.Options{Prefix: envAuthGrpcPrefix, AltPrefix: envPrefix}); err != nil {
		return nil, nil, ioutil.NopCloser(nil), "", errors.Wrap(errGrpcConfig, err)
	}

	grpcClient, tracer, tracerCloser, secure, err := grpcClient.Setup(config, "auth", jaegerURL)
	if err != nil {
		return nil, nil, ioutil.NopCloser(nil), "", err
	}

	message := "without TLS"
	if secure {
		message = "with TLS"
	}
	return authapi.NewClient(tracer, grpcClient, config.Timeout), grpcClient, tracerCloser, message, nil
}
