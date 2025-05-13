// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lpg

import (
	"iter"
	"sort"
	"strings"
)

// Shared instance for empty string sets to avoid allocations.
// Revert to original fastMap.
var emptyStringSet = NewStringSet()

type StringSet struct {
	// Use original fastMap
	M *fastMap
}

// uses a pre existing stringset or creates a new one if empty
func FastNewStringSet(set *StringSet) *StringSet {
	if set == nil {
		return emptyStringSet
	}
	return set
}

func NewStringSet(s ...string) *StringSet {
	// Use original newFastMap
	set := newFastMap()
	for _, x := range s {
		set.add(x, x)
	}
	return &StringSet{M: set}
}

func (set *StringSet) CloneN(n int) *StringSet {
	if n <= 0 || set == nil || set.M == nil {
		return emptyStringSet
	}
	newSet := emptyFastMap(n)

	current := set.M.l.Front()
	for i := 0; i < n && current != nil; i++ {
		str := current.Value.(string)
		newSet.add(str, str) // Avoid duplicate type assertion
		current = current.Next()
	}
	return &StringSet{M: newSet}
}

func (set *StringSet) Iter(f func(string) bool) {
	if set == nil || set.M == nil {
		return
	}
	// Use list iteration
	current := set.M.l.Front()
	for current != nil {
		if f(current.Value.(string)) {
			break
		}
		current = current.Next()
	}
}

func (set *StringSet) Clone() *StringSet {
	if set == nil || set.M == nil || set.M.size() == 0 {
		return emptyStringSet
	}
	return set.CloneN(set.M.size())
}

func (set *StringSet) IsEqual(s *StringSet) bool {
	// Use M.size()
	return set.Len() == s.Len() && set.HasAllSet(s) // Keep Len() for consistency?
}

func (set *StringSet) Has(s string) bool {
	if set == nil || set.M == nil {
		return false
	}
	return set.M.has(s)
}

func (set *StringSet) HasAny(s ...string) bool {
	for _, x := range s {
		if set.Has(x) {
			return true
		}
	}
	return false
}

func (set *StringSet) Intersect(s *StringSet) *StringSet {
	newSet := NewStringSet()
	if set == nil || s == nil || set.M == nil || s.M == nil {
		return newSet
	}
	setToIterate := set
	other := s
	if set.Len() > s.Len() {
		setToIterate = s
		other = set
	}
	// Use list iteration
	current := setToIterate.M.l.Front()
	for current != nil {
		key := current.Value.(string)
		if other.Has(key) {
			newSet.M.add(key, key)
		}
		current = current.Next()
	}
	return newSet
}

func (set *StringSet) HasAnySet(s *StringSet) bool {
	if set == nil || s == nil || set.M == nil || s.M == nil {
		return false
	}
	// Use list iteration
	current := set.M.l.Front()
	for current != nil {
		if s.Has(current.Value.(string)) {
			return true
		}
		current = current.Next()
	}
	return false
}

func (set *StringSet) HasAll(s ...string) bool {
	if len(s) == 0 {
		return true
	}
	if set == nil || set.M == nil || set.Len() < len(s) {
		return false
	}
	for _, x := range s {
		if !set.Has(x) {
			return false
		}
	}
	return true
}

func (set *StringSet) HasAllSet(s *StringSet) bool {
	if s == nil || s.Len() == 0 {
		return true
	}
	if set == nil || set.M == nil || set.Len() < s.Len() {
		return false
	}
	// Use list iteration
	current := s.M.l.Front() // Iterate the set we are checking against
	for current != nil {
		if !set.Has(current.Value.(string)) {
			return false
		}
		current = current.Next()
	}
	return true
}

func (set *StringSet) Add(s ...string) *StringSet {
	if set == nil { // Should not happen if constructed with NewStringSet
		return nil // Or panic? Original code didn't handle nil receiver here.
	}
	if set.M == nil { // Need to initialize map if receiver exists but M is nil?
		set.M = newFastMap()
	}
	for _, x := range s {
		set.M.add(x, x)
	}
	return set
}

func (set *StringSet) AddSet(s StringSet) *StringSet {
	if set == nil {
		return nil
	}
	if set.M == nil {
		set.M = newFastMap()
	}
	// Use list iteration on the input set 's'
	current := s.M.l.Front()
	for current != nil {
		set.M.add(current.Value.(string), current.Value.(string))
		current = current.Next()
	}
	return set
}

func (set *StringSet) Remove(s ...string) *StringSet {
	if set == nil || set.M == nil {
		return set
	}
	for _, x := range s {
		set.M.remove(x)
	}
	return set
}

func (set *StringSet) Slice() []string {
	if set == nil || set.M == nil {
		return nil
	}
	// Use list iteration
	ret := make([]string, 0, set.M.size())
	current := set.M.l.Front()
	for current != nil {
		ret = append(ret, current.Value.(string))
		current = current.Next()
	}
	return ret
}

func (set *StringSet) SortedSlice() []string {
	ret := set.Slice()
	sort.Strings(ret)
	return ret
}

func (set *StringSet) String() string {
	return strings.Join(set.Slice(), ",")
}

func (set *StringSet) Len() int {
	if set == nil || set.M == nil {
		return 0
	}
	return set.M.size()
}

func (set *StringSet) Replace(other *StringSet, handleRemoved, handleAdded func(string)) {
	if set == nil || set.M == nil {
		*set = *NewStringSet()
	}
	originalKeys := make(map[string]struct{}) // Need original keys to check against `other` later
	current := set.M.l.Front()
	for current != nil {
		key := current.Value.(string)
		originalKeys[key] = struct{}{} // Store original keys
		if other == nil || !other.Has(key) {
			handleRemoved(key)
		}
		current = current.Next()
	}

	// Create the new map/list state based on 'other'
	newMap := newFastMap()
	if other != nil {
		otherCurrent := other.M.l.Front()
		for otherCurrent != nil {
			key := otherCurrent.Value.(string)
			newMap.add(key, key)
			if _, existed := originalKeys[key]; !existed {
				handleAdded(key)
			}
			otherCurrent = otherCurrent.Next()
		}
	}
	set.M = newMap // Replace internal map/list entirely
}

func (f *StringSet) Range() iter.Seq[string] {
	return func(yield func(k string) bool) {
		f.Iter(func(k string) bool {
			return !yield(k)
		})
	}
}

func (f *StringSet) Iterator() Iterator {
	if f == nil || f.M == nil {
		// Return iterator for empty list
		return &listIterator{size: 0}
	}
	// Return original list iterator
	return &listIterator{next: f.M.l.Front(), size: f.M.size()}
}
