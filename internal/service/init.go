package service

import "web_demoservice/internal/cache"

func NewOrderService(cache *cache.Cache, repo OrderRepository) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}
