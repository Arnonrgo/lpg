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
	"github.com/dolthub/swiss"
	"sort"
	"strings"
)

type StringSet struct {
	M *swiss.Map[string, bool]
}

func NewStringSet(s ...string) *StringSet {
	ret := swiss.NewMap[string, bool](uint32(len(s)))
	newSet := StringSet{M: ret}
	for _, x := range s {
		newSet.M.Put(x, true)
	}
	return &newSet
}
func (set *StringSet) CloneN(n int) *StringSet {
	ret := swiss.NewMap[string, bool](uint32(n))
	i := 0
	set.M.Iter(func(x string, _ bool) bool {
		ret.Put(x, true)
		i++
		return i == n
	})
	return &StringSet{M: ret}
}

func (set *StringSet) Iter(f func(string) bool) {
	set.M.Iter(func(x string, _ bool) bool {
		return f(x)
	})
}

func (set *StringSet) Clone() *StringSet {
	newSet := NewStringSet()
	set.M.Iter(func(x string, _ bool) bool {
		newSet.M.Put(x, true)
		return false
	})
	return newSet
}

func (set *StringSet) IsEqual(s *StringSet) bool {
	return set.M.Count() == s.M.Count() && set.HasAllSet(s)
}

func (set *StringSet) Has(s string) bool {
	return set.M.Has(s)
}

func (set *StringSet) HasAny(s ...string) bool {
	for _, x := range s {
		if set.M.Has(x) {
			return true
		}
	}
	return false
}
func (set *StringSet) Intersect(s *StringSet) *StringSet {
	newSet := StringSet{M: swiss.NewMap[string, bool](uint32(set.M.Count()))}
	setToIterate := set
	other := s
	if set.M.Count() > s.M.Count() {
		setToIterate = s
		other = set
	}
	setToIterate.M.Iter(func(x string, _ bool) bool {
		if other.M.Has(x) {
			newSet.M.Put(x, true)
		}
		return false
	})
	return &newSet
}

func (set *StringSet) HasAnySet(s *StringSet) bool {
	res := false
	s.M.Iter(func(x string, _ bool) bool {
		if set.M.Has(x) {
			res = true
			return true
		}
		return false
	})
	return res
}

func (set *StringSet) HasAll(s ...string) bool {
	if len(s) == 0 || set.M.Count() < len(s) {
		return false
	}
	for _, x := range s {
		if !set.M.Has(x) {
			return false
		}
	}
	return true
}

func (set *StringSet) HasAllSet(s *StringSet) bool {
	if set.M.Count() < s.M.Count() {
		return false
	}
	res := true
	s.M.Iter(func(x string, _ bool) bool {
		if !set.M.Has(x) {
			res = false
			return true
		}
		return false
	})
	return res
}

func (set *StringSet) Add(s ...string) *StringSet {
	for _, x := range s {
		set.M.Put(x, true)
	}
	return set
}

func (set *StringSet) AddSet(s StringSet) *StringSet {
	s.M.Iter(func(x string, _ bool) bool {
		set.M.Put(x, true)
		return false
	})
	return set
}

func (set *StringSet) Remove(s ...string) *StringSet {
	for _, x := range s {
		set.M.Delete(x)
	}
	return set
}

func (set *StringSet) Slice() []string {
	ret := make([]string, 0, set.M.Count())
	set.M.Iter(func(x string, _ bool) bool {
		ret = append(ret, x)
		return false
	})
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
	return set.M.Count()
}

func (set *StringSet) Replace(other *StringSet, handleRemoved, handleAdded func(string)) {
	set.M.Iter(func(x string, _ bool) bool {
		if !other.M.Has(x) {
			handleRemoved(x)
		}
		return false
	})
	newSet := swiss.NewMap[string, bool](uint32(other.M.Count()))
	other.M.Iter(func(x string, _ bool) bool {
		if !set.M.Has(x) {
			handleAdded(x)
		}
		newSet.Put(x, true)
		return false
	})
	set.M = newSet
}
