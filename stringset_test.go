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
