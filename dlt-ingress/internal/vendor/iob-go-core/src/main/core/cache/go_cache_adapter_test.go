package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testKey = "testKey::test"
const testValue = "testValue"

var ttls = map[string]time.Duration{
	"testKey": 1 * time.Second,
}

func TestCoreGoCacheAdapter_GetNotFound(t *testing.T) {
	c := NewGoCache(ttls)
	val, err := c.Get(context.Background(), testKey)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "key testKey not found in the cache")
}

func TestCoreGoCacheAdapter_GetOk(t *testing.T) {
	c := NewGoCache(ttls)
	_ = c.Set(context.Background(), testKey, testValue)
	val, err := c.Get(context.Background(), testKey)
	assert.Equal(t, val, testValue)
	assert.Nil(t, err)
}

func TestCoreGoCacheAdapter_GetAfterTTL(t *testing.T) {
	c := NewGoCache(ttls)
	_ = c.Set(context.Background(), testKey, testValue)
	time.Sleep(1 * time.Second)
	val, err := c.Get(context.Background(), testKey)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "key testKey not found in the cache")
}

func TestCoreGoCacheAdapter_Delete(t *testing.T) {
	c := NewGoCache(ttls)
	_ = c.Set(context.Background(), testKey, testValue)
	err := c.Delete(context.Background(), testKey)
	assert.Nil(t, err)

	val, err := c.Get(context.Background(), testKey)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "key testKey not found in the cache")
}
