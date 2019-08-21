package ratelimit

import (
	"testing"
	"time"
)

func TestBucket_AllowWithStatus(t *testing.T) {
	pool := newRedisPool("localhost:6379")
	store := NewRedigoStore(pool)
	rate := 5
	key := "test"

	n := time.Now()
	timeNow = func() time.Time {
		return n
	}

	b := &Bucket{
		identifier: "marketing_campaigns",
		rate:       rate,
		windowSize: 1000,
		store:      &store,
	}
	// nextRefresh := time.Duration(b.windowSize) * time.Millisecond
	for i := 1; i <= rate; i++ {
		got, err := b.AllowWithStatus(key)
		if err != nil {
			t.Error("AllowWithStatus call failed = ", err)
		}
		if !got.Allowed {
			t.Errorf("limit breached")
		}
		//if got.NextRefresh != nextRefresh {
		//	t.Errorf("invalid next refresh duration, got = %d, want = %d\n", got.NextRefresh, nextRefresh)
		//}
	}
	got, _ := b.AllowWithStatus(key)
	if got.Allowed {
		t.Error("expected to hit the limit")
	}
}
