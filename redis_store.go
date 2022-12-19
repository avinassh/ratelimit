package ratelimit

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedigoStore struct {
	pool   *redis.Pool
	script *redis.Script
}

func newRedisPool(address string) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				address,
			)
		},
	}
}

func NewRedigoStore(pool *redis.Pool) RedigoStore {
	// we will initialise with the script
	conn := pool.Get()
	defer conn.Close()

	var script = redis.NewScript(1, tokenBucketScript)
	err := script.Load(conn)
	if err != nil {
		panic(err)
	}
	return RedigoStore{pool: pool, script: script}
}

func NewRedigoSWStore(pool *redis.Pool) RedigoStore {
	// we will initialise with the script
	conn := pool.Get()
	defer conn.Close()

	var script = redis.NewScript(1, slidingWindowScript)
	err := script.Load(conn)
	if err != nil {
		panic(err)
	}
	return RedigoStore{pool: pool, script: script}
}

func (s *RedigoStore) inc(key string, rate, windowSize, now int) (map[string]int, error) {
	conn := s.pool.Get()
	defer conn.Close()

	r, err := redis.IntMap(s.script.Do(conn, key, rate, windowSize, now))
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *RedigoStore) Inc(key string, rate, windowSize, now int) (StoreResponse, error) {
	return buildStoreResponse(s.inc(key, rate, windowSize, now))
}

func buildStoreResponse(result map[string]int, err error) (StoreResponse, error) {
	if err != nil {
		return StoreResponse{}, err
	}
	response := StoreResponse{
		Counter:    result["c"],
		LastRefill: MillisToTime(int64(result["ts"])),
	}
	if result["s"] == 1 {
		response.Allowed = true
	}
	return response, nil
}

func TimeMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func MillisToTime(m int64) time.Time {
	return time.Unix(0, m*int64(time.Millisecond))
}
