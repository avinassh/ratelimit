package ratelimit

import (
	"testing"
)

func TestRedigoStore_inc(t *testing.T) {
	pool := newRedisPool("localhost:6379")
	store := NewRedigoStore(pool)

	// let's start with a key which doesn't exist yet
	key := "test"
	now := 1000
	limit := 5
	windowSize := 1000 // i.e. 1s in millis
	success := 1
	fail := 0

	// clean up
	defer func() {
		pool.Get().Do("DEL", key)
	}()

	for i := 1; i <= limit; i++ {
		res, err := store.inc(key, limit, windowSize, now)
		if err != nil {
			t.Error("inc call failed = ", err)
		}
		if res["ts"] != now {
			t.Errorf("invalid ts got = %d, want = %d", res["ts"], now)
		}
		if res["c"] != i {
			t.Errorf("invalid counter got = %d, want = %d", res["c"], i)
		}
		if res["s"] != success {
			t.Errorf("invalid status got = %d, want = %d", res["s"], success)
		}
	}

	for i := 1; i <= limit; i++ {
		res, err := store.inc(key, limit, windowSize, now)
		if err != nil {
			t.Error("inc call failed = ", err)
		}
		if res["ts"] != now {
			t.Errorf("invalid ts got = %d, want = %d", res["ts"], now)
		}
		if res["c"] != limit {
			t.Errorf("invalid counter got = %d, want = %d", res["c"], i)
		}
		if res["s"] != fail {
			t.Errorf("invalid status got = %d, want = %d", res["s"], fail)
		}
	}

	// lets increase the timestamp by 1000 and it should still fail
	expectedNow := now
	now = 1999
	for i := 1; i <= limit; i++ {
		res, err := store.inc(key, limit, windowSize, now)
		if err != nil {
			t.Error("inc call failed = ", err)
		}
		if res["ts"] != expectedNow {
			t.Errorf("invalid ts got = %d, want = %d", res["ts"], expectedNow)
		}
		if res["c"] != limit {
			t.Errorf("invalid counter got = %d, want = %d", res["c"], limit)
		}
		if res["s"] != fail {
			t.Errorf("invalid status got = %d, want = %d", res["s"], fail)
		}
	}

	// lets elapse a second and now it should pass again
	now = 2000
	for i := 1; i <= limit; i++ {
		res, err := store.inc(key, limit, windowSize, now)
		if err != nil {
			t.Error("inc call failed = ", err)
		}
		if res["ts"] != now {
			t.Errorf("invalid ts got = %d, want = %d", res["ts"], now)
		}
		if res["c"] != i {
			t.Errorf("invalid counter got = %d, want = %d", res["c"], i)
		}
		if res["s"] != success {
			t.Errorf("invalid status got = %d, want = %d", res["s"], success)
		}
	}
}
