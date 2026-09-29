package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	Id   uuid.UUID
	Name string
}
type TestDomainService struct{}

func (ds *TestDomainService) Find(id uuid.UUID, name string) (*TestModel, error) {
	if name == "error" {
		return nil, fmt.Errorf("test error")
	}
	return &TestModel{Id: id, Name: name}, nil
}

func (ds *TestDomainService) Delete(id uuid.UUID, name string) error {
	if name == "error" {
		return fmt.Errorf("test error")
	}
	return nil
}

type TestDomainServiceCache struct {
	TestDomainService
	Cache Port
}

func (ds *TestDomainServiceCache) Find(id uuid.UUID, name string) (*TestModel, error) {
	return Cache[*TestModel](context.Background(), ds.Cache, func() (*TestModel, error) {
		return ds.TestDomainService.Find(id, name)
	}, "test_model", id)
}

func (ds *TestDomainServiceCache) Delete(id uuid.UUID, name string) error {
	return Evict(context.Background(), ds.Cache, func() error {
		return ds.TestDomainService.Delete(id, name)
	}, "test_model", id)
}

func TestCoreCacheHelpers_CacheOk(t *testing.T) {
	cacheDecorator := TestDomainServiceCache{Cache: NewGoCache(map[string]time.Duration{})}
	id := uuid.New()
	name := "testName"
	testModel, err := cacheDecorator.Find(id, name)
	assert.Nil(t, err)
	assert.Equal(t, testModel.Id, id)
	assert.Equal(t, testModel.Name, name)

	// Calling the service with same key parameters, should return cached value
	testModel, err = cacheDecorator.Find(id, "testName2")
	assert.Nil(t, err)
	assert.Equal(t, testModel.Id, id)
	assert.Equal(t, testModel.Name, name)
}

func TestCoreCacheHelpers_CacheError(t *testing.T) {
	cacheDecorator := TestDomainServiceCache{Cache: NewGoCache(map[string]time.Duration{})}
	testModel, err := cacheDecorator.Find(uuid.New(), "error")
	assert.NotNil(t, err)
	assert.Nil(t, testModel)
}

func TestCoreCacheHelpers_EvictOk(t *testing.T) {
	cacheDecorator := TestDomainServiceCache{Cache: NewGoCache(map[string]time.Duration{})}
	id := uuid.New()
	name := "testName"
	testModel, err := cacheDecorator.Find(id, name)
	assert.Nil(t, err)
	assert.Equal(t, testModel.Id, id)
	assert.Equal(t, testModel.Name, name)

	err = cacheDecorator.Delete(id, name)
	assert.Nil(t, err)

	// After cache evict, no element from cache is returned
	name2 := "testName2"
	testModel, err = cacheDecorator.Find(id, name2)
	assert.Nil(t, err)
	assert.Equal(t, testModel.Id, id)
	assert.Equal(t, testModel.Name, name2)
}

func TestCoreCacheHelpers_EvictError(t *testing.T) {
	cacheDecorator := TestDomainServiceCache{Cache: NewGoCache(map[string]time.Duration{})}
	err := cacheDecorator.Delete(uuid.New(), "error")
	assert.NotNil(t, err)
}
