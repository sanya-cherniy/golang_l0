package order

import (
	"errors"
	"sync"
	"time"
)

type Storage interface {
	Insert(e *Order)
	Get(uid string) (Order, error)
	GetAll() ([]Order, error)
}

type orderStorage struct {
	order       Order
	RequestTime time.Time
}

type MemoryStorage struct {
	data map[string]orderStorage
	sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]orderStorage),
	}
}

func (s *MemoryStorage) Insert(o *Order) {
	s.RLock()
	defer s.RUnlock()
	s.data[o.OrderUID] = orderStorage{
		order:       *o,
		RequestTime: time.Now(),
	}
}

func (s *MemoryStorage) Get(uid string) (Order, error) {
	s.RLock()
	defer s.RUnlock()
	orderStorage, ok := s.data[uid]
	if !ok {
		return orderStorage.order, errors.New("order not found")
	}
	orderStorage.RequestTime = time.Now()
	s.data[uid] = orderStorage
	return orderStorage.order, nil

}

func (s *MemoryStorage) GetAll() (orders []Order, err error) {
	for key, orderStorage := range s.data {
		orders = append(orders, orderStorage.order)
		orderStorage.RequestTime = time.Now()
		s.RLock()
		s.data[key] = orderStorage
		s.RUnlock()
	}
	return orders, nil

}

func (s *MemoryStorage) CashChecker(lifeTime int64) {
	for {
		time.Sleep(time.Duration(lifeTime) * time.Millisecond)
		s.RLock()
		for key := range s.data {
			if time.Now().Sub(s.data[key].RequestTime).Milliseconds() > lifeTime {
				delete(s.data, key)
			}
		}
		s.RUnlock()

	}
}
