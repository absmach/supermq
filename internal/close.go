package internal

import (
	grpcClient "github.com/mainflux/mainflux/internal/client/grpc"
	logger "github.com/mainflux/mainflux/logger"
)
func Close(log logger.Logger, clientHandler grpcClient.ClientHandler) {
	if err := clientHandler.Close(); err != nil {
		log.Warn(err.Error())
	}
}