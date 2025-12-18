// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/pkg/oauth2/store"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	deviceCodePrefix = "oauth:device:code:"
	userCodePrefix   = "oauth:device:user:"
)

const (
	testRedisAddr = "localhost:6379"
)

// setupRedisTest creates a test Redis client and clears test keys.
func setupRedisTest(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
		DB:   1, // Use DB 1 for tests to avoid conflicts
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis tests")
	}

	// Clear any existing test keys
	pattern := "oauth:device:*"
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		client.Del(ctx, iter.Val())
	}

	cleanup := func() {
		// Clear test keys after test
		iter := client.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			client.Del(ctx, iter.Val())
		}
		client.Close()
	}

	return client, cleanup
}

func TestRedisDeviceCodeStore_Save(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	code := oauth2.DeviceCode{
		DeviceCode:      "test-device-code",
		UserCode:        "ABCD-EFGH",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       600,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:ABCD-EFGH",
	}

	err := deviceStore.Save(code)
	require.NoError(t, err)

	// Verify device code was saved
	deviceKey := deviceCodePrefix + code.DeviceCode
	exists, err := client.Exists(ctx, deviceKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Verify user code mapping was saved
	userKey := userCodePrefix + code.UserCode
	exists, err = client.Exists(ctx, userKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Verify TTL was set
	ttl, err := client.TTL(ctx, deviceKey).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, oauth2.DeviceCodeExpiry)
}

func TestRedisDeviceCodeStore_Get(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	originalCode := oauth2.DeviceCode{
		DeviceCode:      "test-device-code-2",
		UserCode:        "XXXX-YYYY",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       600,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:XXXX-YYYY",
		AccessToken:     "test-access-token",
		Approved:        false,
	}

	err := deviceStore.Save(originalCode)
	require.NoError(t, err)

	// Retrieve the code
	retrievedCode, err := deviceStore.Get(originalCode.DeviceCode)
	require.NoError(t, err)

	assert.Equal(t, originalCode.DeviceCode, retrievedCode.DeviceCode)
	assert.Equal(t, originalCode.UserCode, retrievedCode.UserCode)
	assert.Equal(t, originalCode.Provider, retrievedCode.Provider)
	assert.Equal(t, originalCode.Approved, retrievedCode.Approved)
	assert.Equal(t, originalCode.AccessToken, retrievedCode.AccessToken)
}

func TestRedisDeviceCodeStore_GetNotFound(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	_, err := deviceStore.Get("non-existent-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "device code not found")
}

func TestRedisDeviceCodeStore_GetByUserCode(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	originalCode := oauth2.DeviceCode{
		DeviceCode:      "test-device-code-3",
		UserCode:        "ZZZZ-AAAA",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       600,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:ZZZZ-AAAA",
	}

	err := deviceStore.Save(originalCode)
	require.NoError(t, err)

	// Retrieve by user code
	retrievedCode, err := deviceStore.GetByUserCode(originalCode.UserCode)
	require.NoError(t, err)

	assert.Equal(t, originalCode.DeviceCode, retrievedCode.DeviceCode)
	assert.Equal(t, originalCode.UserCode, retrievedCode.UserCode)
	assert.Equal(t, originalCode.Provider, retrievedCode.Provider)
}

func TestRedisDeviceCodeStore_GetByUserCodeNotFound(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	_, err := deviceStore.GetByUserCode("non-existent-user-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user code not found")
}

func TestRedisDeviceCodeStore_Update(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	originalCode := oauth2.DeviceCode{
		DeviceCode:      "test-device-code-4",
		UserCode:        "BBBB-CCCC",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       600,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:BBBB-CCCC",
		Approved:        false,
		AccessToken:     "",
	}

	err := deviceStore.Save(originalCode)
	require.NoError(t, err)

	// Update the code
	originalCode.Approved = true
	originalCode.AccessToken = "new-access-token"

	err = deviceStore.Update(originalCode)
	require.NoError(t, err)

	// Retrieve and verify update
	retrievedCode, err := deviceStore.Get(originalCode.DeviceCode)
	require.NoError(t, err)

	assert.True(t, retrievedCode.Approved)
	assert.Equal(t, "new-access-token", retrievedCode.AccessToken)
}

func TestRedisDeviceCodeStore_UpdateNotFound(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	code := oauth2.DeviceCode{
		DeviceCode: "non-existent-code",
	}

	err := deviceStore.Update(code)
	require.Error(t, err)
}

func TestRedisDeviceCodeStore_Delete(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	code := oauth2.DeviceCode{
		DeviceCode:      "test-device-code-5",
		UserCode:        "DDDD-EEEE",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       600,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:DDDD-EEEE",
	}

	err := deviceStore.Save(code)
	require.NoError(t, err)

	// Verify it exists
	_, err = deviceStore.Get(code.DeviceCode)
	require.NoError(t, err)

	// Delete it
	err = deviceStore.Delete(code.DeviceCode)
	require.NoError(t, err)

	// Verify it's gone
	_, err = deviceStore.Get(code.DeviceCode)
	require.Error(t, err)

	// Verify user code mapping is also gone
	deviceKey := deviceCodePrefix + code.DeviceCode
	exists, err := client.Exists(ctx, deviceKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	userKey := userCodePrefix + code.UserCode
	exists, err = client.Exists(ctx, userKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestRedisDeviceCodeStore_DeleteNotFound(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	err := deviceStore.Delete("non-existent-code")
	require.Error(t, err)
}

func TestRedisDeviceCodeStore_Expiry(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	code := oauth2.DeviceCode{
		DeviceCode:      "test-device-code-6",
		UserCode:        "FFFF-GGGG",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       1,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:FFFF-GGGG",
	}

	err := deviceStore.Save(code)
	require.NoError(t, err)

	// Manually set a very short TTL for testing
	deviceKey := deviceCodePrefix + code.DeviceCode
	userKey := userCodePrefix + code.UserCode
	err = client.Expire(ctx, deviceKey, 1*time.Second).Err()
	require.NoError(t, err)
	err = client.Expire(ctx, userKey, 1*time.Second).Err()
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Verify it's expired
	_, err = deviceStore.Get(code.DeviceCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "device code not found")

	_, err = deviceStore.GetByUserCode(code.UserCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user code not found")
}

func TestRedisDeviceCodeStore_MultipleInstances(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create two store instances (simulating two service instances)
	deviceStore1 := store.NewRedisDeviceCodeStore(ctx, client)
	deviceStore2 := store.NewRedisDeviceCodeStore(ctx, client)

	code := oauth2.DeviceCode{
		DeviceCode:      "test-device-code-7",
		UserCode:        "HHHH-IIII",
		VerificationURI: "http://localhost/verify",
		ExpiresIn:       600,
		Interval:        5,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "device:HHHH-IIII",
		Approved:        false,
	}

	// Save from instance 1
	err := deviceStore1.Save(code)
	require.NoError(t, err)

	// Retrieve from instance 2
	retrievedCode, err := deviceStore2.Get(code.DeviceCode)
	require.NoError(t, err)
	assert.Equal(t, code.DeviceCode, retrievedCode.DeviceCode)
	assert.False(t, retrievedCode.Approved)

	// Update from instance 2
	retrievedCode.Approved = true
	retrievedCode.AccessToken = "shared-token"
	err = deviceStore2.Update(retrievedCode)
	require.NoError(t, err)

	// Verify update from instance 1
	verifiedCode, err := deviceStore1.Get(code.DeviceCode)
	require.NoError(t, err)
	assert.True(t, verifiedCode.Approved)
	assert.Equal(t, "shared-token", verifiedCode.AccessToken)
}

func TestRedisDeviceCodeStore_ConcurrentAccess(t *testing.T) {
	client, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()
	deviceStore := store.NewRedisDeviceCodeStore(ctx, client)

	// Save multiple codes concurrently
	numCodes := 10
	errChan := make(chan error, numCodes)

	for i := 0; i < numCodes; i++ {
		go func(idx int) {
			code := oauth2.DeviceCode{
				DeviceCode:      fmt.Sprintf("concurrent-code-%d", idx),
				UserCode:        fmt.Sprintf("CODE-%04d", idx),
				VerificationURI: "http://localhost/verify",
				ExpiresIn:       600,
				Interval:        5,
				Provider:        "google",
				CreatedAt:       time.Now(),
				State:           fmt.Sprintf("device:CODE-%04d", idx),
			}
			errChan <- deviceStore.Save(code)
		}(i)
	}

	// Wait for all saves
	for i := 0; i < numCodes; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	// Verify all codes can be retrieved
	for i := 0; i < numCodes; i++ {
		code, err := deviceStore.Get(fmt.Sprintf("concurrent-code-%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("CODE-%04d", i), code.UserCode)
	}
}
