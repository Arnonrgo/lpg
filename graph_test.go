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
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGraphCRUD(t *testing.T) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	for i := 0; i < len(nodes)-1; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "e", nil, nil)
	}

	if len(NodeSlice(g.GetNodes())) != len(nodes) {
		t.Errorf("Wrong node count")
	}
	if g.NumNodes() != len(nodes) {
		t.Errorf("Wrong numNodes")
	}
	nodes[2].DetachAndRemove()
	if len(NodeSlice(g.GetNodes())) != len(nodes)-1 {
		t.Errorf("Wrong node count")
	}
	if g.NumNodes() != len(nodes)-1 {
		t.Errorf("Wrong numNodes")
	}
}

func TestContexts(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()

	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "e", nil, NewStringSet("default", "whatever"))
	}
	i := len(nodes) - 2
	g.NewEdge(nodes[i], nodes[i+1], "e", nil, NewStringSet("something", "whatever"))

	edges := make([]*Edge, 0)
	g.ProcessEdgesWithAnyContext(NewStringSet("something"), func(e *Edge) {
		edges = append(edges, e)
	})
	assert.Equal(t, 1, len(edges))
	edges = make([]*Edge, 0)
	g.ProcessEdgesWithAnyContext(NewStringSet("default", "whatever"), func(e *Edge) {
		edges = append(edges, e)
	})
	assert.Equal(t, len(nodes)-1, len(edges))
	edges = make([]*Edge, 0)
	g.ProcessEdgesWithAnyContext(NewStringSet("default"), func(e *Edge) {
		edges = append(edges, e)

	})
	assert.Equal(t, len(nodes)-2, len(edges))

}

func TestRetrieveEdgesWithContexts(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()

	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "edge", nil, NewStringSet("default", "whatever"))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "other", nil, nil)
	}
	edges := g.GetEdgesWithAnyLabel(NewStringSet("other"))
	for edges.Next() {
		fmt.Println("here")
		assert.Equal(t, "other", edges.Edge().GetLabel())
	}
	edges = g.GetEdgesWithAnyLabel(NewStringSet("edge"))
	for edges.Next() {
		fmt.Println("edge")
		assert.Equal(t, "edge", edges.Edge().GetLabel())
	}
}

func TestFind_MatchLabel(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	nodes = append(nodes, g.NewNode([]string{"h"}, nil, nil))

	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "edge", nil, NewStringSet("default", "whatever"))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "other", map[string]interface{}{"a": "test"}, nil)
	}
	g.NewEdge(nodes[0], nodes[1], "special", map[string]interface{}{"b": "test"}, nil)

	// return only matching when label is found
	found := 0
	match, _ := g.FindNodes(NewStringSet("h"), map[string]interface{}{})
	for match.Next() {
		found++
		assert.Equal(t, "h", match.Node().GetLabels().Slice()[0])
	}
	assert.Equal(t, 1, found)
	found = 0
	ematch, _ := g.FindEdges("special", map[string]interface{}{})
	for ematch.Next() {
		found++
		assert.Equal(t, "special", ematch.Edge().label)
	}
	assert.Equal(t, 1, found)
}

func TestFind_MatchProperty(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()
	g.AddNodePropertyIndex("a", BtreeIndex)
	g.AddEdgePropertyIndex("b", BtreeIndex)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	nodes = append(nodes, g.NewNode([]string{"h"}, map[string]interface{}{"a": "test"}, nil))

	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "edge", nil, NewStringSet("default", "whatever"))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "other", map[string]interface{}{"a": "test"}, nil)
	}
	g.NewEdge(nodes[0], nodes[1], "special", map[string]interface{}{"b": "test"}, nil)

	// return only matching when label is found
	found := 0
	match, _ := g.FindNodes(nil, map[string]interface{}{"a": "test"})
	for match.Next() {
		found++
		assert.Equal(t, "h", match.Node().GetLabels().Slice()[0])
	}
	assert.Equal(t, 1, found)
	found = 0
	ematch, _ := g.FindEdges("", map[string]interface{}{"b": "test"})
	for ematch.Next() {
		found++
		assert.Equal(t, "special", ematch.Edge().label)
	}
	assert.Equal(t, 1, found)
}

func TestFind_EmptyOnNonExistingLabels(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	nodes = append(nodes, g.NewNode([]string{"h"}, nil, nil))

	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "edge", nil, NewStringSet("default", "whatever"))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "other", map[string]interface{}{"a": "test"}, nil)
	}
	g.NewEdge(nodes[0], nodes[1], "special", map[string]interface{}{"b": "test"}, nil)

	// can't find labels returns empty
	match, _ := g.FindNodes(NewStringSet("blah", "h"), map[string]interface{}{})
	assert.False(t, match.Next())
	edges, _ := g.FindEdges("blah", map[string]interface{}{})
	assert.False(t, edges.Next())

}
func TestFind_ErrorWhenPropertyNotIndexed(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	nodes = append(nodes, g.NewNode([]string{"h"}, nil, nil))

	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "edge", nil, NewStringSet("default", "whatever"))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "other", map[string]interface{}{"a": "test"}, nil)
	}
	g.NewEdge(nodes[0], nodes[1], "special", map[string]interface{}{"b": "test"}, nil)

	size := 0
	_, err := g.FindNodes(NewStringSet("special"), map[string]interface{}{"b": "test"})
	assert.Error(t, err)
	assert.Equal(t, 0, size)

	size = 0
	_, err = g.FindEdges("special", map[string]interface{}{"b": "test"})
	assert.Error(t, err)
	assert.Equal(t, 0, size)
}

func TestFind_FilterBothLabelAndProperty(t *testing.T) {
	nodes := make([]*Node, 0)
	g := NewGraph()
	g.AddNodePropertyIndex("a", BtreeIndex)
	g.AddEdgePropertyIndex("b", BtreeIndex)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	nodes = append(nodes, g.NewNode([]string{"a"}, map[string]interface{}{"a": "test"}, nil))
	nodes = append(nodes, g.NewNode([]string{"b"}, map[string]interface{}{"a": "test"}, nil))

	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "edge", nil, NewStringSet("default", "whatever"))
	}
	for i := 0; i < len(nodes)-2; i++ {
		g.NewEdge(nodes[i], nodes[i+1], "other", map[string]interface{}{"a": "test"}, nil)
	}
	g.NewEdge(nodes[0], nodes[1], "b", map[string]interface{}{"b": "test"}, nil)
	g.NewEdge(nodes[0], nodes[1], "c", map[string]interface{}{"b": "test"}, nil)

	// return only matching when label is found
	found := 0
	match, _ := g.FindNodes(NewStringSet("a"), map[string]interface{}{"a": "test"})
	for match.Next() {
		found++
		assert.Equal(t, "a", match.Node().GetLabels().Slice()[0])
	}
	assert.Equal(t, 1, found)
	found = 0
	ematch, _ := g.FindEdges("b", map[string]interface{}{"b": "test"})
	for ematch.Next() {
		found++
		assert.Equal(t, "b", ematch.Edge().label)
	}
	assert.Equal(t, 1, found)
}

func BenchmarkGetProperty(b *testing.B) {
	g := NewGraph()
	for i := 0; i < 1000; i++ {
		is := fmt.Sprintf("%v", i)
		node := g.NewNode([]string{"a", "b", "c", is}, map[string]interface{}{"a": "b", "c": "d"}, nil)
		node.GetProperty("a")
		node.GetProperty("blasah")

	}
	node := g.allNodes.head

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		node = node.next
		if node == g.allNodes.tail {
			node = g.allNodes.head
		}
		node.GetProperty("a")
		node.GetProperty("blasah")
	}
}

func BenchmarkAddNode(b *testing.B) {
	g := NewGraph()
	for n := 0; n < b.N; n++ {
		g.NewNode([]string{"a", "b", "c"}, map[string]interface{}{"a": "b", "c": "d"}, nil)
	}
}

func benchmarkItrNodes(numNodes int, b *testing.B) {
	g := NewGraph()
	var x *Node
	for i := 0; i < numNodes; i++ {
		g.NewNode([]string{"a", "b", "c"}, map[string]interface{}{"a": "b", "c": "d"}, nil)
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			x = nodes.Node()
		}
	}
	_ = x
}

func BenchmarkItrNodes1000(b *testing.B)  { benchmarkItrNodes(1000, b) }
func BenchmarkItrNodes10000(b *testing.B) { benchmarkItrNodes(10000, b) }

func benchmarkItrNodesViaIndex(numNodes int, b *testing.B) {
	g := NewGraph()
	var x *Node
	for i := 0; i < numNodes; i++ {
		g.NewNode([]string{"a", "b", "c"}, map[string]interface{}{"a": "b", "c": "d"}, nil)
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.index.nodesByLabel.Iterator(); nodes.Next(); {
			x = nodes.Node()
		}
	}
	_ = x
}

func benchmarkItrNodesViaIndex2(numNodes int, b *testing.B) {
	g := NewGraph()
	var x *Node
	for i := 0; i < numNodes; i++ {
		g.NewNode([]string{"a", "b", "c"}, map[string]interface{}{"a": "b", "c": "d"}, nil)
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.index.nodesByLabel.IteratorAllLabels(NewStringSet("a", "c")); nodes.Next(); {
			x = nodes.Node()
		}
	}
	_ = x
}

func BenchmarkItrNodesViaIndex1000(b *testing.B)     { benchmarkItrNodesViaIndex(1000, b) }
func BenchmarkItrNodesViaIndex10000(b *testing.B)    { benchmarkItrNodesViaIndex(10000, b) }
func BenchmarkItrNodesViaIndex_2_10000(b *testing.B) { benchmarkItrNodesViaIndex2(10000, b) }

func BenchmarkCreateEdge(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	labels := []string{"a", "b", "c", "d"}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(nodes)-1; i++ {
			g.FastNewEdge(nodes[i], nodes[i+1], labels[i%4], nil, nil)
		}
	}
}

func BenchmarkItrAllEdge(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		g.NewEdge(nodes[i], nodes[i+1], labels[i%4], nil, nil)
	}
	var edge *Edge

	for n := 0; n < b.N; n++ {
		for edges := g.GetEdges(); edges.Next(); {
			edge = edges.Edge()
		}
	}
	_ = edge
}

func BenchmarkItrNodeEdges(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
		}
	}
	var edge *Edge

	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			for edges := node.GetEdges(OutgoingEdge); edges.Next(); {
				edge = edges.Edge()
			}
		}
	}
	_ = edge
}

func BenchmarkItrNodeEdgesLabel(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 100000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
		}
	}
	var edge *Edge

	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			for edges := node.GetEdgesWithAnyLabel(OutgoingEdge, NewStringSet("a", "c")); edges.Next(); {
				edge = edges.Edge()
			}
		}
	}
	_ = edge
}

func BenchmarkItrEdgesByLabel(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 10000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
		}
	}
	var edge *Edge

	for n := 0; n < b.N; n++ {
		for edges := g.GetEdgesWithAnyLabel(NewStringSet("a", "c")); edges.Next(); {
			edge = edges.Edge()
		}
	}
	_ = edge
}
