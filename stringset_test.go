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

func TestStringSet_Replace(t *testing.T) {
	added := 0
	removed := 0
	set := NewStringSet("a", "b", "c")
	other := NewStringSet("a", "d", "e", "f")
	set.Replace(other, func(s string) { removed++ }, func(s string) { added++ })
	assert.True(t, set.Has("a"))
	assert.True(t, set.Has("f"))
	assert.False(t, set.Has("b"))
	assert.Equal(t, 3, added)
	assert.Equal(t, 2, removed)
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
