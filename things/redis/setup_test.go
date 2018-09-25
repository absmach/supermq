package redis_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-redis/redis"
	dockertest "gopkg.in/ory-am/dockertest.v3"
)

const (
	wrongID    = 0
	wrongValue = "wrong-value"
)

var (
	cacheClient *redis.Client
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("redis", "4.0.9-alpine", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		cacheClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
			Password: "",
			DB:       0,
		})

		return cacheClient.Ping().Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()
	// When you're done, kill and remove the container
	// err = pool.Purge(resource)
	defer pool.Purge(resource)

	os.Exit(code)
}
