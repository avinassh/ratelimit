package ratelimit

import (
	"time"

	"github.com/garyburd/redigo/redis"
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

	var script = redis.NewScript(1, getScript())
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
		LastRefill: time.UnixMilli(int64(result["ts"])),
	}
	if result["s"] == 1 {
		response.Allowed = true
	}
	return response, nil
}
