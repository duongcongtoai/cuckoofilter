package cuckoo

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_ScalableCuckooFilter_FalseNegative(t *testing.T) {
	filter := NewScalableCuckooFilter()

	// filter := NewFilter(100000)
	exist := []string{}
	removed := []string{}
	for i := 0; i < 15000; i++ {
		id := fmt.Sprintf("%d-", i) + uuid.NewString()
		exist = append(exist, id)
		removed = append(removed, id+"to-removed")
	}
	for i := 0; i < len(exist); i++ {
		filter.Insert([]byte(exist[i]))
	}
	for i := 0; i < len(exist); i++ {
		filter.Insert([]byte(removed[i]))
	}

	for i := 0; i < len(exist); i++ {
		filter.Delete([]byte(removed[i]))
	}
	for i := 0; i < len(exist); i++ {
		filter.Insert([]byte(removed[i]))
	}
	falseNegatives := []int{}
	for i := 0; i < len(exist); i++ {
		exist := filter.Lookup([]byte(exist[i]))
		if !exist {
			falseNegatives = append(falseNegatives, i)
		}
	}
	assert.Equal(t, 0, len(falseNegatives))
}

func TestTodo(t *testing.T) {
	filter := NewScalableCuckooFilterV2()
	batchInsert := func(filter *ScalableCuckooFilterv2, exist []string,
		historicalData func() []string) {
	TRYLOOP:
		for {
			for _, item := range exist {
				_, err := filter.Insert([]byte(item))
				if err != nil {
					filter.ResetAndScale()
					for _, oldItem := range historicalData() {
						_, err := filter.Insert([]byte(oldItem))
						if err != nil {
							panic(err)
						}
					}
					continue TRYLOOP
				}
			}
			return
		}
	}
	// filter := NewFilter(100000)
	exist := []string{}
	removed := []string{}
	for i := 0; i < 15000; i++ {
		id := fmt.Sprintf("%d-", i) + uuid.NewString()
		exist = append(exist, id)
		removed = append(removed, id+"1")
	}
	batchInsert(filter, exist, func() []string { return nil })
	batchInsert(filter, removed, func() []string { return exist })

	for i := 0; i < len(exist); i++ {
		filter.Delete([]byte(removed[i]))
	}

	batchInsert(filter, removed, func() []string { return exist })

	falsePositive := []int{}
	for i := 0; i < len(exist); i++ {
		exist := filter.Lookup([]byte(exist[i]))
		if !exist {
			falsePositive = append(falsePositive, i)
		}
	}
	if len(falsePositive) > 0 {
		panic(fmt.Sprintf("%d", len(falsePositive)))
	}
	fmt.Println(len(filter.filters.buckets) * 4)
}

func TestNormalUse(t *testing.T) {
	filter := NewScalableCuckooFilter()
	for i := 0; i < 100000; i++ {
		filter.Insert([]byte("NewScalableCuckooFilter_" + strconv.Itoa(i)))
	}
	testStr := []byte("NewScalableCuckooFilter")
	b := filter.Insert(testStr)
	assert.True(t, b)
	b = filter.Lookup(testStr)
	assert.True(t, b)
	b = filter.Delete(testStr)
	assert.True(t, b)
	b = filter.Lookup(testStr)
	assert.False(t, b)
	b = filter.Lookup([]byte("NewScalableCuckooFilter_233"))
	assert.True(t, b)
	b = filter.InsertUnique([]byte("NewScalableCuckooFilter_599"))
	assert.False(t, b)
}

func TestScalableCuckooFilter_DecodeEncode(t *testing.T) {
	filter := NewScalableCuckooFilter(func(filter *ScalableCuckooFilter) {
		filter.loadFactor = 0.8
	})
	for i := 0; i < 100000; i++ {
		filter.Insert([]byte("NewScalableCuckooFilter_" + strconv.Itoa(i)))
	}
	bytes := filter.Encode()
	decodeFilter, err := DecodeScalableFilter(bytes)
	assert.Nil(t, err)
	assert.Equal(t, decodeFilter.loadFactor, float32(0.8))
	b := decodeFilter.Lookup([]byte("NewScalableCuckooFilter_233"))
	assert.True(t, b)
	for i, f := range decodeFilter.filters {
		assert.Equal(t, f.count, filter.filters[i].count)
	}

}
