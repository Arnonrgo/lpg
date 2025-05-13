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
	"container/list"
)

type graphElement interface{ *Node | *Edge | any }

// A fastSet is a set of objects with constant-time
// insertion/deletion, with iterator support (iteration order undefined).
type fastSet struct {
	n map[int]interface{} // Using standard Go map now
}

type fastMap struct {
	n map[string]*list.Element
	l *list.List
}

func newFastMap() *fastMap {
	return &fastMap{
		n: make(map[string]*list.Element),
		l: list.New(),
	}
}

func emptyFastMap(size int) *fastMap {
	return &fastMap{
		n: make(map[string]*list.Element, size),
		l: list.New(),
	}
}
func (f *fastMap) init() {
	f.n = make(map[string]*list.Element)
	f.l.Init()
}

func (f *fastMap) size() int {
	return len(f.n)
}

func (f *fastMap) add(id string, item interface{}) bool {
	el, exists := f.n[id]
	if exists {
		el.Value = item
		return false
	}
	el = f.l.PushBack(item)
	f.n[id] = el
	return true
}
func (f *fastMap) get(id string) (interface{}, bool) {
	el, ok := f.n[id]
	if !ok {
		return nil, false
	}
	return el.Value, true
}

func (f *fastMap) remove(id string) bool {
	el, ext := f.n[id]
	if !ext {
		return false
	}
	delete(f.n, id)
	f.l.Remove(el)
	return true
}

func (f *fastMap) has(id string) bool {
	_, ret := f.n[id]
	return ret
}

func (f *fastMap) iterator() Iterator {
	return &listIterator{next: f.l.Front(), size: f.size()}
}

// newFastSet accepts an optional size hint for the underlying map.
func newFastSet(hint ...int) *fastSet {
	initialSize := 1 // Default hint
	if len(hint) > 0 && hint[0] > 0 {
		initialSize = hint[0]
	}
	return &fastSet{
		n: make(map[int]interface{}, initialSize),
	}
}

func (f *fastSet) init(hint ...int) {
	initialSize := 10 // Default hint
	if len(hint) > 0 && hint[0] > 0 {
		initialSize = hint[0]
	}
	f.n = make(map[int]interface{}, initialSize)
	// f.l = list.New() // Remove list creation
}

func (f *fastSet) size() int {
	return len(f.n)
}

// Add a new item. Returns true if added
func (f *fastSet) add(id int, item interface{}) bool {
	if _, exists := f.n[id]; !exists {
		// el := f.l.PushBack(item) // No longer using list
		f.n[id] = item // Directly add/update in map
		return true
	} else {
		// Update the item if it already exists (optional, depends on desired set semantics)
		f.n[id] = item
		return false
	}
}

func (f *fastSet) get(id int) (interface{}, bool) {
	// Get item directly from map
	return f.n[id], true
}

// Remove an item. Returns true if removed
func (f *fastSet) remove(id int) bool {
	if _, exists := f.n[id]; exists {
		delete(f.n, id)
		return true
	}
	return false
}

func (f *fastSet) has(id int) bool {
	_, exists := f.n[id]
	return exists
}

// Iterator for fastSet - iterates directly over the Go map (random order).
// Stores items in a slice upon creation for iteration.
type mapIterator struct { // Renamed from intMapIterator
	items []interface{}
	index int
	sz    int
}

func (it *mapIterator) Next() bool {
	it.index++
	return it.index < it.sz
}

func (it *mapIterator) Value() interface{} {
	if it.index < 0 || it.index >= it.sz {
		return nil // Or panic
	}
	return it.items[it.index]
}

func (it *mapIterator) MaxSize() int {
	return it.sz
}

func (f *fastSet) iterator() Iterator {
	size := len(f.n)
	items := make([]interface{}, 0, size)
	for _, value := range f.n { // Iterate using standard map range
		items = append(items, value)
	}
	return &mapIterator{items: items, index: -1, sz: size} // Use the renamed iterator
}

type NodeSet struct {
	set fastSet
}

func NewNodeSet() *NodeSet {
	nm := &NodeSet{}
	nm.set.init()
	return nm
}

func (set *NodeSet) Add(node *Node) {
	set.set.add(node.id, node)
}

func (set *NodeSet) Remove(node *Node) {
	set.set.remove(node.id)
}

func (set *NodeSet) Has(node *Node) bool {
	return set.set.has(node.id)
}

func (set *NodeSet) Len() int {
	return set.set.size()
}

func (set *NodeSet) Iterator() NodeIterator {
	i := set.set.iterator()
	return nodeIterator{i}
}

func (set *NodeSet) Slice() []*Node {
	return NodeSlice(set.Iterator())
}

// EdgeSet keeps an unordered set of edges
type EdgeSet struct {
	set fastSet
}

func NewEdgeSet() *EdgeSet {
	es := &EdgeSet{}
	es.set.init()
	return es
}

func (set *EdgeSet) Add(edge *Edge) {
	set.set.add(edge.id, edge)
}

func (set *EdgeSet) Remove(edge *Edge) {
	set.set.remove(edge.id)
}

func (set *EdgeSet) Len() int {
	return set.set.size()
}

func (set *EdgeSet) Iterator() EdgeIterator {
	i := set.set.iterator()
	return edgeIterator{i}
}

func (set *EdgeSet) Slice() []*Edge {
	return EdgeSlice(set.Iterator())
}
