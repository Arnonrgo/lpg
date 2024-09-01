package lpg

import (
	"cmp"
	"fmt"
	"github.com/dolthub/swiss"
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/kamstrup/intmap"
	btree2 "github.com/tidwall/btree"
	"math/rand/v2"
	"testing"
)

func compare(a, b interface{}) int {
	return cmp.Compare(a.(int64), b.(int64))
}

func BenchmarkTree2(b *testing.B) {
	tree := *btree2.NewMap[int64, int](16)
	// get random number out of 100000NewWith(16, ComparePropertyValue)
	for i := 0; i < b.N; i++ {
		rnd := rand.Int64N(1_000_000)
		tree.Set(rnd, i)
	}

}

func BenchmarkGetTree2(b *testing.B) {
	tree := *btree2.NewMap[int64, int](16)
	// get random number out of 100000NewWith(16, ComparePropertyValue)
	for i := 0; i < 1_000_000; i++ {
		rnd := rand.Int64N(1_000_000)
		tree.Set(rnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rnd := rand.Int64N(1_000_000)
		tree.Get(rnd)
	}
}

func cmp2(a int64, b int64) int {
	return cmp.Compare(a, b)
}

func BenchmarkFastSet(b *testing.B) {
	// fast but not thread sage
	set := newFastSet()
	set.init()
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(1_000_000)
		set.add(rnd, i)
		set.has(rnd)
	}
}

func BenchmarkSwissMap(b *testing.B) {
	// faster but not thread sage
	set := swiss.NewMap[int, int](42)
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(1_000_000)
		set.Put(rnd, i)
		set.Has(rnd)
	}
}
func BenchmarkIntMap(b *testing.B) {
	// very fast (with no threads),  thread safe, needs integer Keys
	set := intmap.New[int32, int32](10)
	for i := 0; i < b.N; i++ {
		rnd := rand.Int32N(1_000_000)
		set.Put(rnd, int32(i))
		set.Has(rnd)
	}
}

func BenchmarkSwissMap_string(b *testing.B) {
	// faster but not thread sage
	set := swiss.NewMap[string, int](100)
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(10_000_000)
		srnd := fmt.Sprintf("%d", rnd)
		set.Put(srnd, i)
		set.Has(srnd)
	}
}
func BenchmarkSwissMap_Iterate(b *testing.B) {
	set := swiss.NewMap[string, int](100)

	for i := 0; i < 10000; i++ {
		srnd := fmt.Sprintf("%d", i)
		set.Put(srnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Iter(func(key string, value int) bool {
			_ = fmt.Sprintf("%d", value)
			return false
		})
	}
}

func BenchmarkLinkedMap_Iterate(b *testing.B) {
	set := linkedhashmap.New[string, int]()

	for i := 0; i < 10000; i++ {
		srnd := fmt.Sprintf("%d", i)
		set.Put(srnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := set.Iterator()
		for x.Next() {
			_ = fmt.Sprintf("%d", x.Value())

		}
	}
}
func BenchmarkFastMap_string(b *testing.B) {
	// faster but not thread sage
	set := newFastMap()
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(100000)
		srnd := fmt.Sprintf("%d", rnd)
		set.add(srnd, i)
		_ = set.has(srnd)
	}
}
func BenchmarkFastsMap_Iterate(b *testing.B) {
	set := newFastMap()

	for i := 0; i < 10000; i++ {
		srnd := fmt.Sprintf("%d", i)
		set.add(srnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := set.iterator()
		for x.Next() {
			_ = fmt.Sprintf("%d", x.Value())

		}
	}
}
func BenchmarkSwissMap_Instantiate(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = swiss.NewMap[string, int](0)
	}
}

func BenchmarkFastMap(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = newFastMap()
	}
}
func BenchmarkMap_Instantiate(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = make(map[string]int)
	}
}

func BenchmarkMap_Iterate(b *testing.B) {
	set := make(map[string]int)
	for i := 0; i < 10000; i++ {
		srnd := fmt.Sprintf("%d", i)
		set[srnd] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range set {
			_ = fmt.Sprintf("%d", value)
		}
	}
}

func BenchmarkMap_string(b *testing.B) {
	// faster but not thread sage
	set := make(map[string]int)
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(100000)
		srnd := fmt.Sprintf("%d", rnd)
		set[srnd] = i
		_ = set[srnd]

	}
}
