// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"time"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	ClientTLS bool          `env:"CLIENT_TLS"    default:""`
	CACerts   string        `env:"CA_CERTS"      default:""`
	URL       string        `env:"GRPC_URL"      default:""`
	Timeout   time.Duration `env:"GRPC_TIMEOUT"  default:"1s"`
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
