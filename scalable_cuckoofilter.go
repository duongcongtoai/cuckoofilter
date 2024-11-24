package cuckoo

import (
	"bytes"
	"encoding/gob"

	f16 "github.com/panmari/cuckoofilter"
)

const (
	DefaultLoadFactor = 0.9
	DefaultCapacity   = 10000
)

type ScalableCuckooFilter struct {
	filters    []*f16.Filter
	loadFactor float32
	initialCap uint
	//when scale(last filter size * loadFactor >= capacity) get new filter capacity
	// scaleFactor func(capacity uint) uint
}

func WithLoadFactor(loadFactor float32) option {
	return func(sfilter *ScalableCuckooFilter) {
		sfilter.loadFactor = loadFactor
	}
}

func WithInitialCap(cap uint) option {
	return func(sfilter *ScalableCuckooFilter) {
		sfilter.initialCap = cap
	}
}

type option func(*ScalableCuckooFilter)
type Store struct {
	Bytes      [][]byte
	LoadFactor float32
}

// NewScalableCuckooFilter naive implementation of scalable cuckoo filter
// but guarantees the followings:
// - zero false-negative
// - false-positive rate is less than 0.001 (it is actually even less than 0.0005, but
// the correct rate is not calculated by the author yet)
// Note that even though "github.com/panmari/cuckoofilter" implementation
// provide r ~= 0.0001, the implementation of scalable cuckoo filter
// theoretically has higher r, because it also correlates with the number of underlying
// filters.
//
// It is suggested that user rebuilt the filter overtime to achieve the best performance
// if the size of data scales up/down overtime.
func NewScalableCuckooFilter(opts ...option) *ScalableCuckooFilter {
	sfilter := new(ScalableCuckooFilter)
	for _, opt := range opts {
		opt(sfilter)
	}
	configure(sfilter)
	return sfilter
}

func (sf *ScalableCuckooFilter) Lookup(data []byte) bool {
	for _, filter := range sf.filters {
		if filter.Lookup(data) {
			return true
		}
	}
	return false
}

func (sf *ScalableCuckooFilter) Reset() {
	for _, filter := range sf.filters {
		filter.Reset()
	}
}

func (sf *ScalableCuckooFilter) Insert(data []byte) bool {
	needScale := false
	lastFilter := sf.filters[len(sf.filters)-1]
	if lastFilter.LoadFactor() > float64(sf.loadFactor) {
		needScale = true
	} else {
		b := lastFilter.Insert(data)
		needScale = !b
	}
	if !needScale {
		return true
	}

	newFilter := f16.NewFilter(sf.initialCap)
	sf.filters = append(sf.filters, newFilter)
	return newFilter.Insert(data)
}

func (sf *ScalableCuckooFilter) InsertUnique(data []byte) bool {
	if sf.Lookup(data) {
		return false
	}
	return sf.Insert(data)
}

func (sf *ScalableCuckooFilter) Delete(data []byte) bool {
	for _, filter := range sf.filters {
		if filter.Delete(data) {
			return true
		}
	}
	return false
}

func (sf *ScalableCuckooFilter) Count() uint {
	var sum uint
	for _, filter := range sf.filters {
		sum += filter.Count()
	}
	return sum

}

func (sf *ScalableCuckooFilter) Encode() []byte {
	slice := make([][]byte, len(sf.filters))
	for i, filter := range sf.filters {
		encode := filter.Encode()
		slice[i] = encode
	}
	store := &Store{
		Bytes:      slice,
		LoadFactor: sf.loadFactor,
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(store)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func (sf *ScalableCuckooFilter) DecodeWithParam(fBytes []byte, opts ...option) (*ScalableCuckooFilter, error) {
	instance, err := DecodeScalableFilter(fBytes)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(instance)
	}
	return instance, nil
}

func DecodeScalableFilter(fBytes []byte) (*ScalableCuckooFilter, error) {
	buf := bytes.NewBuffer(fBytes)
	dec := gob.NewDecoder(buf)
	store := &Store{}
	err := dec.Decode(store)
	if err != nil {
		return nil, err
	}
	filterSize := len(store.Bytes)
	instance := NewScalableCuckooFilter(func(filter *ScalableCuckooFilter) {
		filter.filters = make([]*f16.Filter, filterSize)
	}, func(filter *ScalableCuckooFilter) {
		filter.loadFactor = store.LoadFactor
	})
	for i, oneBytes := range store.Bytes {
		filter, err := f16.Decode(oneBytes)
		if err != nil {
			return nil, err
		}
		instance.filters[i] = filter
	}
	return instance, nil

}

func configure(sfilter *ScalableCuckooFilter) {
	if sfilter.loadFactor == 0 {
		sfilter.loadFactor = DefaultLoadFactor
	}
	if sfilter.initialCap == 0 {
		sfilter.initialCap = DefaultCapacity
	}
	// NOTE: in order for scalable cuckfoo filter to provide reliable deletion
	// the size of all children filters must be the same
	// H. Chen, L. Liao, H. Jin and J. Wu, "The dynamic cuckoo filter"
	// if sfilter.scaleFactor == nil {
	// 	sfilter.scaleFactor = func(currentSize uint) uint {
	// 		return currentSize * bucketSize
	// 	}
	// }
	if sfilter.filters == nil {
		initFilter := f16.NewFilter(sfilter.initialCap)
		sfilter.filters = []*f16.Filter{initFilter}
	}
}
