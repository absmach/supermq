// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import "net/http"

type connectionStats struct {
	Current       uint64 `json:"current"`
	Total         uint64 `json:"total"`
	Disconnections uint64 `json:"disconnections"`
}

type messageStats struct {
	Received        uint64 `json:"received"`
	Sent            uint64 `json:"sent"`
	PublishReceived uint64 `json:"publish_received"`
	PublishSent     uint64 `json:"publish_sent"`
}

type byteStats struct {
	Received uint64 `json:"received"`
	Sent     uint64 `json:"sent"`
}

type subscriptionStats struct {
	Active           uint64 `json:"active"`
	RetainedMessages uint64 `json:"retained_messages"`
}

type errorStats struct {
	Protocol uint64 `json:"protocol"`
	Auth     uint64 `json:"auth"`
	Authz    uint64 `json:"authz"`
	Packet   uint64 `json:"packet"`
}

type statsResponse struct {
	UptimeSeconds float64           `json:"uptime_seconds"`
	Connections   connectionStats   `json:"connections"`
	Messages      messageStats      `json:"messages"`
	Bytes         byteStats         `json:"bytes"`
	Subscriptions subscriptionStats `json:"subscriptions"`
	Errors        errorStats        `json:"errors"`
}

func (s *Server) buildStatsResponse() statsResponse {
	st := s.broker.Stats()
	return statsResponse{
		UptimeSeconds: st.GetUptime().Seconds(),
		Connections: connectionStats{
			Current:        st.GetCurrentConnections(),
			Total:          st.GetTotalConnections(),
			Disconnections: st.GetDisconnections(),
		},
		Messages: messageStats{
			Received:        st.GetMessagesReceived(),
			Sent:            st.GetMessagesSent(),
			PublishReceived: st.GetPublishReceived(),
			PublishSent:     st.GetPublishSent(),
		},
		Bytes: byteStats{
			Received: st.GetBytesReceived(),
			Sent:     st.GetBytesSent(),
		},
		Subscriptions: subscriptionStats{
			Active:           st.GetSubscriptions(),
			RetainedMessages: st.GetRetainedMessages(),
		},
		Errors: errorStats{
			Protocol: st.GetProtocolErrors(),
			Auth:     st.GetAuthErrors(),
			Authz:    st.GetAuthzErrors(),
			Packet:   st.GetPacketErrors(),
		},
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.broker == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "broker not available")
		return
	}

	writeJSON(w, http.StatusOK, s.buildStatsResponse())
}
