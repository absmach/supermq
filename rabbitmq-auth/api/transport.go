// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/absmach/supermq"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(logger *slog.Logger, instanceID string, clients grpcClientsV1.ClientsServiceClient) http.Handler {
	r := chi.NewRouter()
	r.Route("/auth", func(r chi.Router) {
		r.HandleFunc("/user", authTokenHandler(clients))
		r.Handle("/vhost", authTokenHandler(clients))
		r.Handle("/resource", authTokenHandler(clients))
		r.Handle("/topic", authTokenHandler(clients))
	})
	r.Get("/health", supermq.Health("http", instanceID))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func authTokenHandler(clients grpcClientsV1.ClientsServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var password string
		// var err error
		// switch r.Method {
		// case http.MethodPost:
		// 	var req map[string]interface{}
		// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 		http.Error(w, err.Error(), http.StatusBadRequest)
		// 		return
		// 	}
		// 	password = req["password"].(string)
		// case http.MethodGet:
		// 	password, err = apiutil.ReadStringQuery(r, "password", "")
		// 	if err != nil {
		// 		http.Error(w, err.Error(), http.StatusBadRequest)
		// 		return
		// 	}
		// default:
		// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// }
		// fmt.Println(password)
		// res, err := clients.Authenticate(r.Context(), &grpcClientsV1.AuthnReq{ClientSecret: password})
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusUnauthorized)
		// 	return
		// }
		// if !res.GetAuthenticated() {
		// 	http.Error(w, svcerr.ErrAuthentication.Error(), http.StatusUnauthorized)
		// 	return
		// }
		fmt.Println(r.URL.Query())
		if _, err := w.Write([]byte("allow")); err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
