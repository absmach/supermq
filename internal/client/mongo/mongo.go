// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"fmt"
	"os"

	"github.com/mainflux/mainflux/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect create connection to MongoDB
func Connect(host, port, name string, logger logger.Logger) *mongo.Database {
	addr := fmt.Sprintf("mongodb://%s:%s", host, port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}

	return client.Database(name)
}
