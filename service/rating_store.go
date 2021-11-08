package service

import "sync"

// RatingStore is an interface to store laptop ratings
type RatingStore interface {
	// Add adds a new laptop score to store and returns its rating
	Add(laptopID string, score float64) (*Rating, error)
}

// Rating contains the rating information of a laptop
type Rating struct {
	Count uint32
	Sum	float64
}

// InMemoryRatingStore stores laptop rating in memory
type InMemoryRatingStore struct {
	mutex sync.RWMutex
	rating map[string]*Rating
}
// NewInMemoryRatingStore returns a InMemoryRatingStore
func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

// Add add a new laptop score to the store and returns its rating
func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	// 去RatingStore中查找元素是否存在
	rating := store.rating[laptopID]
	// 不存在则创建一个
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum: score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopID] = rating
	return rating, nil
}