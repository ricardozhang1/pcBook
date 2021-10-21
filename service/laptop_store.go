package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"log"
	"pc_book/pd"
	"sync"
	"time"
)

// ErrAlreadyExists is returned when a record with the same ID already exists in store
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore is an interface to store laptop
type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pd.Laptop) error
	// Find finds a laptop by id
	Find(id string) (*pd.Laptop, error)
	// Search search a laptop by filter, return one by one via the found function
	Search(ctx context.Context, filter *pd.Filter, found func(laptop *pd.Laptop) error) error
}

// InMemoryLaptopStore store laptop inmemory
// 存储在内存中, 使用map
type InMemoryLaptopStore struct {
	mutex 	sync.RWMutex
	data 	map[string]*pd.Laptop
}

// NewInMemoryLaptopStore return a new InMemoryLaptopStore.
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pd.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pd.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	// deep copy
	other, err := deepCopy(laptop)
	if err != nil {
		return fmt.Errorf("can not copy laptop data: %v", err)
	}
	store.data[other.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) Find(id string) (*pd.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}
	return deepCopy(laptop)
}

func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pd.Filter, found func(laptop *pd.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		// heavy process
		time.Sleep(time.Second)
		fmt.Println("check laptop id: ", laptop.GetId())

		// 检查上下文
		if ctx.Err() == context.Canceled || ctx.Err()==context.DeadlineExceeded {
			log.Print("context is canceled")
			return errors.New("context is canceled")
		}

		if isQualified(filter, laptop) {
			// deep copy
			other, err := deepCopy(laptop)
			if err != nil {
				return nil
			}
			err = found(other)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func isQualified(filter *pd.Filter, laptop *pd.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() { return false }
	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() { return false }
	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() { return false }
	// 将内存统一转成bit大小，再进行比较
	if toBit(laptop.Ram) < toBit(filter.MinRam) { return false }
	return true
}

func toBit(memory *pd.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pd.Memory_BIT:
		return value
	case pd.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pd.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3 = 2^13
	case pd.Memory_MEGABYTE:
		return value << 23
	case pd.Memory_GIGABYTE:
		return value << 33
	case pd.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pd.Laptop) (*pd.Laptop, error) {
	other := &pd.Laptop{}

	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}

// DBLaptopStore store laptop DB
// 存储在数据库中
type DBLaptopStore struct {

}
