package service

import (
	"context"
	"errors"
	"grpc-go/pb"
	"log"
	"sync"
)

var ErrAlreadyExist = errors.New("record already exist")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() LaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop, 0),
	}
}

func (m *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.data[laptop.Id]; ok {
		return ErrAlreadyExist
	}
	other := laptop
	m.data[other.Id] = other
	return nil
}

func (m *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if v, ok := m.data[id]; ok {
		return v, nil
	}
	return nil, nil
}

func (m *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, laptop := range m.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Println("context is cancelled")
			return nil
		}
		if isQualified(filter, laptop) {
			err := found(laptop)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pb.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3 = 2^13
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}
