// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/absmach/magistrala"
	"google.golang.org/grpc"
)

func TestNewClient(t *testing.T) {
	type args struct {
		conn    *grpc.ClientConn
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args
		want magistrala.ThingsServiceClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.conn, tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grpcClient_Authorize(t *testing.T) {
	type args struct {
		ctx context.Context
		req *magistrala.ThingsAuthzReq
		in2 []grpc.CallOption
	}
	tests := []struct {
		name    string
		client  grpcClient
		args    args
		wantR   *magistrala.ThingsAuthzRes
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, err := tt.client.Authorize(tt.args.ctx, tt.args.req, tt.args.in2...)
			if (err != nil) != tt.wantErr {
				t.Errorf("grpcClient.Authorize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("grpcClient.Authorize() = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}

func Test_decodeAuthorizeResponse(t *testing.T) {
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
			got, err := decodeAuthorizeResponse(tt.args.in0, tt.args.grpcRes)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeAuthorizeResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeAuthorizeResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encodeAuthorizeRequest(t *testing.T) {
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
			got, err := encodeAuthorizeRequest(tt.args.in0, tt.args.grpcReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeAuthorizeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeAuthorizeRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeError(t *testing.T) {
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
			if err := decodeError(tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("decodeError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
