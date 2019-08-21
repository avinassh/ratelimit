package ratelimit

import (
	"fmt"
	"time"
)

type SlidingWindowApprox struct {
	identifier string
	rate       int
	windowSize int
	store      Store
}

// The sliding window with approximation algorithm
// https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
//
// given a key like `user-post:avinassh`, rate
// if the key doesn't exist, then save the time, increase the counter and allow
// if key exists
// 		- check the last fill time, if it more than window time, then fill can be started
// 		- if we have the last window time, then we will calculate the approximation:
//
//

// NewSlidingWindowWithApproximation returns a new rate limiter which uses the token bucket algorithm.
func NewSlidingWindowWithApproximation(identifier string, rate, windowSize int, store Store) SlidingWindow {
	return SlidingWindow{identifier, rate, windowSize, store}
}

func (sw *SlidingWindowApprox) Allow(key string) (bool, error) {
	s, err := sw.AllowWithStatus(key)
	if err != nil {
		return false, err
	}
	return s.Allowed, nil
}

func (sw *SlidingWindowApprox) AllowWithStatus(key string) (Status, error) {
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
