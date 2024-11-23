package cuckoo

import (
	"errors"
	"fmt"
)

type ScalableCuckooFilterv2 struct {
	filters    *Filter
	loadFactor float32
	//when scale(last filter size * loadFactor >= capacity) get new filter capacity
	scaleFactor func(capacity uint) uint
}

/*
	by default option the grow capacity is:
	capacity , total
	4096  4096
	8192  12288

16384  28672
32768  61440
65536  126,976
*/

type optionv2 func(*ScalableCuckooFilterv2)

func NewScalableCuckooFilterV2(opts ...optionv2) *ScalableCuckooFilterv2 {
	sfilter := new(ScalableCuckooFilterv2)
	for _, opt := range opts {
		opt(sfilter)
	}
	configurev2(sfilter)
	return sfilter
}

func (sf *ScalableCuckooFilterv2) Lookup(data []byte) bool {
	return sf.filters.Lookup(data)
}
func (sf *ScalableCuckooFilterv2) ResetAndScale() {
	newFilter := NewFilter(sf.scaleFactor(uint(len(sf.filters.buckets))))
	sf.filters = newFilter
}

func (sf *ScalableCuckooFilterv2) Insert(data []byte) (bool, error) {
	needScale := false
	lastFilter := sf.filters
	if (float32(lastFilter.count) / float32(len(lastFilter.buckets))) > sf.loadFactor {
		needScale = true
	} else {
		b := lastFilter.Insert(data)
		needScale = !b
	}
	if !needScale {
		return true, nil
	}
	fmt.Printf("%s need scale\n", string(data))
	return false, errors.New("full")
}

func (sf *ScalableCuckooFilterv2) Delete(data []byte) bool {
	return sf.filters.Delete(data)
}

func (sf *ScalableCuckooFilterv2) Count() uint {
	return sf.filters.count
}

func configurev2(sfilter *ScalableCuckooFilterv2) {
	if sfilter.loadFactor == 0 {
		sfilter.loadFactor = DefaultLoadFactor
	}
	if sfilter.scaleFactor == nil {
		sfilter.scaleFactor = func(currentSize uint) uint {
			return currentSize * bucketSize * 2
		}
	}
	if sfilter.filters == nil {
		sfilter.filters = NewFilter(DefaultCapacity)
	}
}
