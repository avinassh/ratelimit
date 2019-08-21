package ratelimit

type Store interface {
	Inc(key string, rate, windowSize, now int) (StoreResponse, error)
}
