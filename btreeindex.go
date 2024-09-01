// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lpg

import (
	"github.com/tidwall/btree"
)

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// A setTree is a B-Tree of linkedhashsets
type setTree[V ordered, I Item] struct {
	tree *btree.Map[V, *fastSet]
}

func (s *setTree[V, I]) add(key V, id int, item I) {
	if s.tree == nil {
		s.tree = btree.NewMap[V, *fastSet](50)
	}
	v, found := s.tree.Get(key)
	if !found {
		v = newFastSet()
		s.tree.Set(key, v)
	}
	v.add(id, item)
}

func (s *setTree[V, I]) remove(key V, id int) {
	if s.tree == nil {
		return
	}
	v, found := s.tree.Get(key)
	if !found {
		return
	}
	v.remove(id)
	if v.size() == 0 {
		s.tree.Delete(key)
	}
}

// find returns the iterator and expected size.
func (s *setTree[V, I]) find(key V) Iterator {
	if s.tree == nil {
		return emptyIterator{}
	}
	v, found := s.tree.Get(key)
	if !found {
		return emptyIterator{}
	}
	return withSize(v.iterator(), v.size())
}

func (s *setTree[V, I]) valueItr() Iterator {
	if s.tree == nil {
		return emptyIterator{}
	}
	treeItr := s.tree.Iter()
	return &funcIterator{
		iteratorFunc: func() Iterator {
			if !treeItr.Next() {
				return nil
			}
			return withSize(treeItr.Value().iterator(), treeItr.Value().size())
		},
	}
}
