// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/things"
)

func TestNewServer(t *testing.T) {
	type args struct {
		svc things.Service
	}
	tests := []struct {
		name string
		args args
		want magistrala.ThingsServiceServer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.svc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grpcServer_Authorize(t *testing.T) {
	type args struct {
		ctx context.Context
		req *magistrala.ThingsAuthzReq
	}
	tests := []struct {
		name    string
		s       *grpcServer
		args    args
		want    *magistrala.ThingsAuthzRes
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Authorize(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("grpcServer.Authorize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grpcServer.Authorize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeAuthorizeRequest(t *testing.T) {
	type args struct {
		in0     context.Context
		grpcReq interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeAuthorizeRequest(tt.args.in0, tt.args.grpcReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeAuthorizeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeAuthorizeRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encodeAuthorizeResponse(t *testing.T) {
	type args struct {
		in0     context.Context
		grpcRes interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeAuthorizeResponse(tt.args.in0, tt.args.grpcRes)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeAuthorizeResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeAuthorizeResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encodeError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := encodeError(tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("encodeError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
