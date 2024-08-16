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

type Item interface {
	any
}

type index[V ordered, I Item] interface {
	add(value V, id int, item I)
	remove(value V, id int)
	find(value V) Iterator
	valueItr() Iterator
}

type IndexType int

const (
	BtreeIndex IndexType = 0
	HashIndex  IndexType = 1
)

type graphIndex struct {
	nodesByLabel   NodeMap
	nodesByContext index[string, *Node]
	edgesByContext index[string, *Edge]
	nodeProperties map[string]index[string, *Node]
	edgeProperties map[string]index[string, *Edge]
}

func newGraphIndex() graphIndex {
	return graphIndex{
		nodesByLabel:   *NewNodeMap(),
		nodesByContext: &setTree[string, *Node]{},
		edgesByContext: &setTree[string, *Edge]{},
		nodeProperties: make(map[string]index[string, *Node]),
		edgeProperties: make(map[string]index[string, *Edge]),
	}
}

// todo enforce string properies only
// NodePropertyIndex sets up an index for the given node property (only support String properties
func (g *graphIndex) NodePropertyIndex(propertyName string, graph *Graph, it IndexType) {
	_, exists := g.nodeProperties[propertyName]
	if exists {
		return
	}
	var ix index[string, *Node]
	if it == BtreeIndex {
		ix = &setTree[string, *Node]{}
	} else {
		ix = &hashIndex[string, *Node]{}
	}
	g.nodeProperties[propertyName] = ix
	// Reindex
	for nodes := graph.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		value, ok := node.properties[propertyName]
		val := value.(string)
		if ok {
			ix.add(val, node.id, node)
		}
	}
}

func (g *graphIndex) isNodePropertyIndexed(propertyName string) index[string, *Node] {
	return g.nodeProperties[propertyName]
}

func (g *graphIndex) isEdgePropertyIndexed(propertyName string) index[string, *Edge] {
	return g.edgeProperties[propertyName]
}

// GetIteratorForNodeProperty returns an iterator for the given
// key/value, and the max size of the resultset. If no index found,
// returns nil,-1
func (g *graphIndex) GetIteratorForNodeProperty(key string, value string) NodeIterator {
	index, found := g.nodeProperties[key]
	if !found {
		return nil
	}
	itr := index.find(value)
	return nodeIterator{itr}
}

// NodesWithProperty returns an iterator that will go through the
// nodes that has the property
func (g *graphIndex) NodesWithProperty(key string) NodeIterator {
	index, found := g.nodeProperties[key]
	if !found {
		return nil
	}
	return nodeIterator{index.valueItr()}
}

// EdgesWithProperty returns an iterator that will go through the
// edges that has the property
func (g *graphIndex) EdgesWithProperty(key string) EdgeIterator {
	index, found := g.edgeProperties[key]
	if !found {
		return nil
	}
	return edgeIterator{index.valueItr()}
}

func (g *graphIndex) addNodeToIndex(node *Node) {
	g.nodesByLabel.Add(node)

	for k, v := range node.properties {
		index, found := g.nodeProperties[k]
		if !found {
			continue
		}
		val := v.(string)
		index.add(val, node.id, node)
	}
}

func (g *graphIndex) removeNodeFromIndex(node *Node) {
	g.nodesByLabel.Remove(node)

	node.contexts.Iter(func(k string) bool {
		g.nodesByContext.remove(k, node.id)
		return false
	})

	for k, v := range node.properties {
		index, found := g.nodeProperties[k]
		if !found {
			continue
		}
		val := v.(string)
		index.remove(val, node.id)
	}
}

// EdgePropertyIndex sets up an index for the given edge property
func (g *graphIndex) EdgePropertyIndex(propertyName string, graph *Graph, it IndexType) {
	_, exists := g.edgeProperties[propertyName]
	if exists {
		return
	}
	var ix index[string, *Edge]
	if it == BtreeIndex {
		ix = &setTree[string, *Edge]{}
	} else {
		ix = &hashIndex[string, *Edge]{}
	}
	g.edgeProperties[propertyName] = ix
	// Reindex
	for edges := graph.GetEdges(); edges.Next(); {
		edge := edges.Edge()
		value, ok := edge.properties[propertyName]
		val := value.(string)
		if ok {
			ix.add(val, edge.id, edge)
		}
	}
}

func (g *graphIndex) addEdgeToIndex(edge *Edge) {
	edge.contexts.Iter(func(k string) bool {
		g.edgesByContext.add(k, edge.id, edge)
		return false
	})
	for k, v := range edge.properties {
		index, found := g.edgeProperties[k]
		if !found {
			continue
		}
		val := v.(string)
		index.add(val, edge.id, edge)
	}
}

func (g *graphIndex) removeEdgeFromIndex(edge *Edge) {
	edge.contexts.Iter(func(k string) bool {
		g.edgesByContext.remove(k, edge.id)
		return false
	})

	for k, v := range edge.properties {
		index, found := g.edgeProperties[k]
		if !found {
			continue
		}
		val := v.(string)
		index.remove(val, edge.id)
	}
}

// GetIteratorForEdgeProperty returns an iterator for the given
// key/value, and the max size of the resultset. If no index found,
// returns nil,-1
func (g *graphIndex) GetIteratorForEdgeProperty(key string, value string) EdgeIterator {
	index, found := g.edgeProperties[key]
	if !found {
		return nil
	}
	itr := index.find(value)
	return edgeIterator{itr}
}
