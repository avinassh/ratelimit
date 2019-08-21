package ratelimit

import (
	"fmt"
	"time"
)

var timeNow = time.Now

type Bucket struct {
	identifier string
	rate       int
	windowSize int
	store      Store
}

type Status struct {
	Allowed     bool
	Remaining   int
	NextRefresh time.Duration
}

type StoreResponse struct {
	Allowed    bool
	Counter    int
	LastRefill time.Time
}

// The Redis algorithm
// given a key like `user-post:avinassh`, rate
// if the doesn't exist then allow, refill and allow
// if key exists
// 		- check if the current value less than limit, if less allow and increment
//		- if the limit has already exceeded, then see if it can be refilled

// NewTokenBucket returns a new rate limiter which uses the token bucket algorithm.
func NewTokenBucket(identifier string, rate, windowSize int, store Store) Bucket {
	return Bucket{identifier, rate, windowSize, store}
}

func (b *Bucket) Allow(key string) (bool, error) {
	s, err := b.AllowWithStatus(key)
	if err != nil {
		return false, err
	}
	return s.Allowed, nil
}

func (b *Bucket) AllowWithStatus(key string) (Status, error) {
	userKey := fmt.Sprintf("%s:%s", b.identifier, key)
	res, err := b.store.Inc(userKey, b.rate, b.windowSize, int(timeNow().UnixMilli()))
	if err != nil {
		return Status{}, err
	}
	timeElapsed := timeNow().Sub(res.LastRefill)
	s := Status{
		Allowed:     res.Allowed,
		Remaining:   b.rate - res.Counter,
		NextRefresh: timeElapsed + (time.Duration(b.windowSize) * time.Millisecond),
	}
	return s, nil
}
