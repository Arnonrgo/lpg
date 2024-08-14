package lpg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringSet_HasAny(t *testing.T) {
	set := NewStringSet("a", "b", "c")
	assert.True(t, set.HasAny("d", "a"))
	assert.False(t, set.HasAny("d", "e"))
}

func TestStringSet_HasAllSet(t *testing.T) {
	set := NewStringSet("a", "b", "c")
	assert.True(t, set.HasAllSet(NewStringSet("c", "b")))
	assert.False(t, set.HasAllSet(NewStringSet("x", "b")))
}

func BenchmarkCloneSet(b *testing.B) {
	set := NewStringSet("a", "b")
	for n := 0; n < b.N; n++ {
		set.Clone()
	}
}

func BenchmarkTakeN(b *testing.B) {
	set := NewStringSet("a", "b", "c", "e", "f", "g", "h")
	for n := 0; n < b.N; n++ {
		set.CloneN(2)
	}
}
