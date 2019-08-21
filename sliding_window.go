package ratelimit

import (
	"fmt"
	"time"
)

type SlidingWindow struct {
	identifier string
	rate       int
	windowSize int
	store      Store
}

// The sliding window algorithm
// given a key like `user-post:avinassh`, rate
// if the doesn't exist then allow, refill and allow
// if key exists
// 		- check if the current value less than limit, if less allow and increment
//		- if the limit has already exceeded, then see if it can be refilled

// NewSlidingWindow returns a new rate limiter which uses the token bucket algorithm.
func NewSlidingWindow(identifier string, rate, windowSize int, store Store) SlidingWindow {
	return SlidingWindow{identifier, rate, windowSize, store}
}

func (sw *SlidingWindow) Allow(key string) (bool, error) {
	s, err := sw.AllowWithStatus(key)
	if err != nil {
		return false, err
	}
	return s.Allowed, nil
}

func (sw *SlidingWindow) AllowWithStatus(key string) (Status, error) {
	userKey := fmt.Sprintf("%s:%s", sw.identifier, key)
	res, err := sw.store.Inc(userKey, sw.rate, sw.windowSize, int(TimeMillis(timeNow())))
	if err != nil {
		return Status{}, err
	}
	timeElapsed := timeNow().Sub(res.LastRefill)
	s := Status{
		Allowed:     res.Allowed,
		Remaining:   sw.rate - res.Counter,
		NextRefresh: timeElapsed + (time.Duration(sw.windowSize) * time.Millisecond),
	}
	return s, nil
}
