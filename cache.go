package sample1

import (
	"fmt"
	"sync"
	"time"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCachePrice is a special price that stores the value and the time it was assigned
type TransparentCachePrice struct {
	value float64
	lastAssigment time.Time
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             map[string]TransparentCachePrice
	pricesMutex        sync.Mutex
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		prices:             map[string]TransparentCachePrice{},
		pricesMutex:        sync.Mutex{},
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
// Is safe for concurrent calls.
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	c.pricesMutex.Lock()
	cachePrice, ok := c.prices[itemCode]
	c.pricesMutex.Unlock()
	if ok {
		if time.Since(cachePrice.lastAssigment) < c.maxAge {
			return cachePrice.value, nil
		}
		delete(c.prices, itemCode)
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	c.pricesMutex.Lock()
	c.prices[itemCode] = TransparentCachePrice{
		value:         price,
		lastAssigment: time.Now(),
	}
	c.pricesMutex.Unlock()
	return price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	results := []float64{}
	resultsStream, errStream := make(chan float64), make(chan error)
	for _, itemCode := range itemCodes {
		go func(internalItemCode string) {
			price, err := c.GetPriceFor(internalItemCode)
			if err != nil {
				errStream <- err
			} else {
				resultsStream <- price
			}
		}(itemCode)
	}

	for range itemCodes {
		select {
		case result := <- resultsStream:
			results = append(results, result)
		case err := <- errStream:
			return []float64{}, err
		}
	}
	return results, nil
}