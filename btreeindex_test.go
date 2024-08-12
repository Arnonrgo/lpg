package lpg

import (
	"cmp"
	"github.com/dolthub/swiss"
	"github.com/kamstrup/intmap"
	btree2 "github.com/tidwall/btree"
	"math/rand/v2"
	b2 "modernc.org/b/v2"
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

func BenchmarkSetTree3(b *testing.B) {
	tree := b2.TreeNew[int64, int](cmp2)
	// get random number out of 100000NewWith(16, ComparePropertyValue)
	for i := 0; i < b.N; i++ {
		rnd := rand.Int64N(1_000_000)
		tree.Set(rnd, i)
		//rnd = rand.Int64N(1_000_000)
		//tree.Get(rnd)
	}

}

func BenchmarkGetTree3(b *testing.B) {
	tree := b2.TreeNew[int64, int](cmp2)
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
