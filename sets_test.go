package lpg

import (
	"testing"
)

func BenchmarkFastSet_Iterator(b *testing.B) {
	g := NewGraph()
	numNodes := 10000
	fs := newFastSet()

	for i := 0; i < numNodes; i++ {
		x := g.NewNode([]string{"a", "b", "c"}, map[string]interface{}{"a": "b", "c": "d"}, nil)
		fs.add(i, x)
	}

	for n := 0; n < b.N; n++ {
		itr := fs.iterator()
		for itr.Next() {
			itr.Value()
		}
	}
}
