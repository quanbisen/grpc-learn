package service

import "sync"

type RateStore interface {
	Add(laptopId string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32
	Sum   float64
}

type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

func NewInMemoryRatingStore() RateStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

func (m *InMemoryRatingStore) Add(laptopId string, score float64) (*Rating, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rating := m.rating[laptopId]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}
	m.rating[laptopId] = rating
	return rating, nil
}
