package cuckoo

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_ScalableCuckooFilter_Reliability(t *testing.T) {
	filter := NewScalableCuckooFilter()

	permaElements := make([]string, 0, DefaultCapacity)
	dynamicElements := make([]string, 0, DefaultCapacity)
	aliens := make([]string, 0, DefaultCapacity)

	// so it trigger the scaling
	for i := 0; i < DefaultCapacity; i++ {
		id := fmt.Sprintf("%d-", i) + uuid.NewString()
		permaElements = append(permaElements, id)
		dynamicElements = append(dynamicElements, id+"dynamic")
		aliens = append(aliens, id+"alien")
	}
	for i := 0; i < len(permaElements); i++ {
		filter.Insert([]byte(permaElements[i]))
	}
	for i := 0; i < len(dynamicElements); i++ {
		filter.Insert([]byte(dynamicElements[i]))
	}
	falsePositive := 0
	for i := 0; i < len(aliens); i++ {
		exist := filter.Lookup([]byte(aliens[i]))
		if exist {
			falsePositive++
		}
	}
	assert.Less(t, float64(falsePositive)/float64(len(aliens)), 0.001)

	// Delete and add back
	for i := 0; i < len(permaElements); i++ {
		filter.Delete([]byte(dynamicElements[i]))
	}
	for i := 0; i < len(permaElements); i++ {
		filter.Insert([]byte(dynamicElements[i]))
	}
	falseNegatives := []int{}
	for i := 0; i < len(permaElements); i++ {
		exist := filter.Lookup([]byte(permaElements[i]))
		if !exist {
			falseNegatives = append(falseNegatives, i)
		}
		exist = filter.Lookup([]byte(dynamicElements[i]))
		if !exist {
			falseNegatives = append(falseNegatives, i)
		}
	}
	assert.Equal(t, 0, len(falseNegatives))

	falsePositive = 0
	for i := 0; i < len(aliens); i++ {
		exist := filter.Lookup([]byte(aliens[i]))
		if exist {
			falsePositive++
		}
	}
	fmt.Println(len(filter.Encode()))
	assert.Less(t, float64(falsePositive)/float64(len(aliens)), 0.001)
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
		assert.Equal(t, f.Count(), filter.filters[i].Count())
	}

}
