package interfaces

type Cache[K comparable, V any] interface {
	Set(key K, value V)
	Get(K) (*V, bool)
}
