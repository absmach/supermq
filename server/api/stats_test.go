// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	mqttbroker "github.com/absmach/fluxmq/mqtt/broker"
	"github.com/absmach/fluxmq/storage/memory"
)

func TestStatsHappyPath(t *testing.T) {
	store := memory.New()
	b := mqttbroker.NewBroker(store, nil, mqttbroker.WithLogger(slog.Default()))
	srv := New(Config{}, b, nil, nil, nil, nil, slog.Default())

	b.Stats().IncrementConnections()
	b.Stats().IncrementConnections()
	b.Stats().DecrementConnections()
	b.Stats().IncrementPublishReceived()
	b.Stats().AddBytesReceived(1024)
	b.Stats().IncrementProtocolErrors()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp statsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.UptimeSeconds <= 0 {
		t.Fatal("expected positive uptime")
	}
	if resp.Connections.Current != 1 {
		t.Fatalf("expected current connections 1, got %d", resp.Connections.Current)
	}
	if resp.Connections.Total != 2 {
		t.Fatalf("expected total connections 2, got %d", resp.Connections.Total)
	}
	if resp.Connections.Disconnections != 1 {
		t.Fatalf("expected disconnections 1, got %d", resp.Connections.Disconnections)
	}
	if resp.Messages.PublishReceived != 1 {
		t.Fatalf("expected publish_received 1, got %d", resp.Messages.PublishReceived)
	}
	if resp.Messages.Received != 1 {
		t.Fatalf("expected messages received 1, got %d", resp.Messages.Received)
	}
	if resp.Bytes.Received != 1024 {
		t.Fatalf("expected bytes received 1024, got %d", resp.Bytes.Received)
	}
	if resp.Errors.Protocol != 1 {
		t.Fatalf("expected protocol errors 1, got %d", resp.Errors.Protocol)
	}
}

func TestStatsNilBrokerReturns503(t *testing.T) {
	srv := New(Config{}, nil, nil, nil, nil, nil, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestStatsRejectsPost(t *testing.T) {
	store := memory.New()
	b := mqttbroker.NewBroker(store, nil, mqttbroker.WithLogger(slog.Default()))
	srv := New(Config{}, b, nil, nil, nil, nil, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stats", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
