package lpg

import (
	"cmp"
	"github.com/cockroachdb/swiss"
	"github.com/jaswdr/faker"
	"github.com/kamstrup/intmap"
	btree2 "github.com/tidwall/btree"
	"math/rand/v2"
	"testing"
)

func BenchmarkTree2(b *testing.B) {
	tree := *btree2.NewMap[int64, int](16)
	// get random number out of 100000NewWith(16, ComparePropertyValue)
	for i := 0; i < b.N; i++ {
		rnd := rand.Int64N(1_000_000)
		tree.Set(rnd, i)
	}

}

func BenchmarkSetTree1(b *testing.B) {
	tree := &setTree[string, int]{}
	faker := faker.New()
	models := make([]string, 10000)
	for i := 0; i < 1000; i++ {
		models[i] = faker.Car().Model()
	}
	for i := 0; i < 100000; i++ {
		tree.add(models[rand.IntN(10000)], i, i)
	}
	b.ResetTimer()
	// get random number out of 100000NewWith(16, ComparePropertyValue)
	for i := 0; i < b.N; i++ {
		tree.find(models[rand.IntN(10000)])
	}

}

func BenchmarkGetTree2(b *testing.B) {
	tree := *btree2.NewMap[int64, int](16)
	// get random number out of 100000NewWith(16, ComparePropertyValue)
	for i := 0; i < 10_000_000; i++ {
		rnd := rand.Int64N(10000)
		tree.Set(rnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rnd := rand.Int64N(20000)
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
	for i := 0; i < 10_000_000; i++ {
		rnd := rand.IntN(10000)
		set.add(rnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(20000)
		set.get(rnd)
	}
}

func BenchmarkSwissMap(b *testing.B) {
	// faster but not thread sage
	set := swiss.New[int, int](42)
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(1000)
		set.Put(rnd, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rnd := rand.IntN(2000)
		set.Get(rnd)
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
