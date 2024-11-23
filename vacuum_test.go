package cuckoo

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func Test_Vacuum(t *testing.T) {
	var a = []uint32{1, 2, 3, 4, 5, 6, 7, 8}
	casted := *(*[]uint64)(unsafe.Pointer(&a))
	fmt.Println(casted)
	fmt.Println(a)
	fmt.Println(casted[3])
	casted[0] = 5
	fmt.Println(a)
	f := VacuumFilter{}
	n := 1 << 12
	f.Init(n, 4, 400)

	fmt.Println(n)
	// filter := NewFilter(100000)
	exist := []uint64{}
	// removed := []uint64{}
	for i := 0; i < n; i++ {
		exist = append(exist, uint64(i))
		inserted := f.Insert(uint64(i))
		if !inserted {
			panic("no")
		}
	}

	// for i := 15000; i < 30000; i++ {
	// 	removed = append(removed, uint64(i))
	// 	f.Insert(uint64(i))
	// }

	// for i := 0; i < len(removed); i++ {
	// 	f.Delete(removed[i])
	// }
	// for i := 0; i < len(removed); i++ {
	// 	f.Insert(removed[i])
	// }
	falseNegatives := []int{}
	for i := 0; i < len(exist); i++ {
		exist := f.Lookup(exist[i])
		if !exist {
			falseNegatives = append(falseNegatives, i)
		}
	}
	assert.Equal(t, 0, len(falseNegatives))
}
