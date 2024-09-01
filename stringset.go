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

type StringSet struct {
	M *fastMap
}

// uses a pre existing stringset or creates a new one if empty
func FastNewStringSet(set *StringSet) *StringSet {
	if set == nil {
		return NewStringSet()
	}
	return set
}

func NewStringSet(s ...string) *StringSet {
	set := newFastMap()
	for _, x := range s {
		set.add(x, x)
	}
	return &StringSet{M: set}
}

func (set *StringSet) CloneN(n int) *StringSet {
	newSet := newFastMap()
	current := set.M.l.Front()
	for i := 0; i < n && current != nil; i++ {
		newSet.add(current.Value.(string), current.Value.(string))
		current = current.Next()
	}
	return &StringSet{M: newSet}
}

func (set *StringSet) Iter(f func(string) bool) {
	if set == nil {
		return
	}
	current := set.M.l.Front()
	for current != nil {
		if f(current.Value.(string)) {
			break
		}
		current = current.Next()
	}
}

func (set *StringSet) Clone() *StringSet {
	return set.CloneN(set.M.size())
}

func (set *StringSet) IsEqual(s *StringSet) bool {
	return set.M.size() == s.M.size() && set.HasAllSet(s)
}

func (set *StringSet) Has(s string) bool {
	return set.M.has(s)
}

func (set *StringSet) HasAny(s ...string) bool {
	for _, x := range s {
		if set.M.has(x) {
			return true
		}
	}
	return false
}
func (set *StringSet) Intersect(s *StringSet) *StringSet {
	newSet := NewStringSet()
	//newSet := StringSet{M: swiss.NewMap[string, bool](uint32(set.M.Count()))}
	setToIterate := set
	other := s
	if set.M.size() > s.M.size() {
		setToIterate = s
		other = set
	}
	current := setToIterate.M.l.Front()
	for current != nil {
		if other.M.has(current.Value.(string)) {
			newSet.M.add(current.Value.(string), current.Value.(string))
		}
		current = current.Next()
	}
	return newSet
}

func (set *StringSet) HasAnySet(s *StringSet) bool {
	res := false
	current := set.M.l.Front()
	for current != nil {
		if s.M.has(current.Value.(string)) {
			res = true
			break
		}
		current = current.Next()
	}
	return res
}

func (set *StringSet) HasAll(s ...string) bool {
	if len(s) == 0 || set.M.size() < len(s) {
		return false
	}
	current := set.M.l.Front()
	for current != nil {
		if !set.M.has(current.Value.(string)) {
			return false
		}
		current = current.Next()
	}
	return true
}

func (set *StringSet) HasAllSet(s *StringSet) bool {
	if set.M.size() < s.M.size() {
		return false
	}
	current := s.M.l.Front()
	for current != nil {
		if !set.M.has(current.Value.(string)) {
			return false
		}
		current = current.Next()
	}
	return true
}

func (set *StringSet) Add(s ...string) *StringSet {
	for _, x := range s {
		set.M.add(x, x)
	}
	return set
}

func (set *StringSet) AddSet(s StringSet) *StringSet {
	current := s.M.l.Front()
	for current != nil {
		set.M.add(current.Value.(string), current.Value.(string))
		current = current.Next()
	}
	return set
}

func (set *StringSet) Remove(s ...string) *StringSet {
	for _, x := range s {
		set.M.remove(x)
	}
	return set
}

func (set *StringSet) Slice() []string {
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
	if set == nil {
		return 0
	}
	return set.M.size()
}

func (set *StringSet) Replace(other *StringSet, handleRemoved, handleAdded func(string)) {
	current := set.M.l.Front()
	for current != nil {
		if !other.M.has(current.Value.(string)) {
			handleRemoved(current.Value.(string))
		}
		current = current.Next()
	}

	newSet := newFastMap()
	current = other.M.l.Front()
	for current != nil {
		if !set.M.has(current.Value.(string)) {
			handleAdded(current.Value.(string))
		}
		newSet.add(current.Value.(string), true)
		current = current.Next()
	}
	set.M = newSet
}

func (f *StringSet) Range() iter.Seq[string] {
	return func(yield func(k string) bool) {
		f.Iter(func(k string) bool {
			return !yield(k)
		})
	}
}

// //	func (f *StringSet) Iterator() Iterator {
// //		next, stop := iter.Pull[string](f.Range())
// //		return &sIterator{next: next, stop: stop, set: f}
// //	}
func (f *StringSet) Iterator() Iterator {
	return &listIterator{next: f.M.l.Front(), size: f.M.size()}
}
