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
	"testing"
)

func BenchmarkPropNonExistsGraph(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h"}))
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		g.NewEdge(nodes[i], nodes[i+1], labels[i%4], nil, nil)
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			nodes.Node().GetProperty("z")
		}
	}
}

func BenchmarkPropExistsGraph(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		if i < 500 {
			nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h"}))
		} else {
			nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h", "z": "zz"}))
		}
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		g.NewEdge(nodes[i], nodes[i+1], labels[i%4], nil, nil)
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			nodes.Node().GetProperty("z")
		}
	}
}

func BenchmarkDeleteEdge(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil))
	}
	labels := []string{"a", "b", "c", "d"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
		}
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			for {
				edgeRemoved := false
				for edges := node.GetEdges(OutgoingEdge); edges.Next(); {
					edges.Edge().Remove()
					edgeRemoved = true
					break
				}
				if !edgeRemoved {
					break
				}
			}
		}
	}
}

func BenchmarkFindEdgeLabel(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil))
	}
	labels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
		}
	}
	b.ResetTimer()
	edgeHasLabel := func(edge *Edge, str string) bool {
		return edge.GetLabel() == str
	}
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			for edges := node.GetEdges(OutgoingEdge); edges.Next(); {
				edgeHasLabel(edges.Edge(), "h")
			}
		}
	}
}

func BenchmarkFindEdgeProp(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil))
	}
	labels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			if i < len(nodes)/2 {
				g.NewEdge(nodes[i], nodes[i+1], label, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h", "z": "zz"}, nil)
			} else {
				g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
			}
		}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			for edges := node.GetEdges(OutgoingEdge); edges.Next(); {
				edges.Edge().GetProperty("z")
			}
		}
	}
}

func BenchmarkAddEdge(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 100_000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil))
	}
	labels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	b.ResetTimer()
	g.AddEdgePropertyIndex("z", BtreeIndex)
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			if i < len(nodes)/2 {
				g.FastNewEdge(nodes[i], nodes[i+1], label, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h", "z": "zz"}, nil)
			} else {
				g.FastNewEdge(nodes[i], nodes[i+1], label, nil, nil)
			}
		}
	}
}
