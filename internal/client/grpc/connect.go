// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"io"
	"io/ioutil"
	"time"

	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/opentracing/opentracing-go"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	errGrpcConnect = errors.New("failed to connect to grpc server")
	errJaeger      = errors.New("failed to initialize jaeger ")
)

type Config struct {
	ClientTLS bool          `env:"CLIENT_TLS"    envDefault:"false"`
	CACerts   string        `env:"CA_CERTS"      envDefault:""`
	URL       string        `env:"URL"           envDefault:""`
	Timeout   time.Duration `env:"TIMEOUT"       envDefault:"1s"`
}

func Connect(cfg Config) (*gogrpc.ClientConn, bool, error) {
	var opts []gogrpc.DialOption
	secure := false
	tc := insecure.NewCredentials()

	if cfg.ClientTLS && cfg.CACerts != "" {
		var err error
		tc, err = credentials.NewClientTLSFromFile(cfg.CACerts, "")
		if err != nil {
			return nil, secure, err
		}
		secure = true
	}

	opts = append(opts, gogrpc.WithTransportCredentials(tc))

	conn, err := gogrpc.Dial(cfg.URL, opts...)
	if err != nil {
		return nil, secure, err
	}
	return conn, secure, nil
}

func Setup(config Config, svcName, jaegerURL string) (*gogrpc.ClientConn, opentracing.Tracer, io.Closer, bool, error) {
	secure := false

	// connect to auth grpc server
	grpcClient, secure, err := Connect(config)
	if err != nil {
		return nil, nil, ioutil.NopCloser(nil), false, errors.Wrap(errGrpcConnect, err)
	}

	// initialize auth tracer for auth grpc client
	tracer, tracerCloser, err := jaegerClient.NewTracer(svcName, jaegerURL)
	if err != nil {
		grpcClient.Close()
		return nil, nil, ioutil.NopCloser(nil), false, errors.Wrap(errJaeger, err)
	}

	return grpcClient, tracer, tracerCloser, secure, nil
}
