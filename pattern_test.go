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
	"testing"
)

func TestPattern(t *testing.T) {
	graph := NewGraph()
	graph.index.NodePropertyIndex("key", graph, BtreeIndex)
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < 9; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
	}
	nodes[5].SetProperty("key", "value")
	symbols := make(map[string]*PatternSymbol)
	pat := Pattern{
		{},
		{Min: 1, Max: 1},
		{Name: "nodes", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	if _, i := pat.getFastestElement(graph, map[string]*PatternSymbol{}); i != 2 {
		t.Errorf("Expecting 2, got %d", i)
	}
	plan, err := pat.GetPlan(graph, symbols)
	if err != nil {
		t.Error(err)
		return
	}
	acc := &DefaultMatchAccumulator{}
	plan.Run(graph, symbols, acc)
	if _, ok := acc.Symbols[0]["nodes"].(*Node); !ok {
		t.Errorf("Expecting one node, got: %v", acc)
	}

	pat = Pattern{
		{Labels: NewStringSet("bogus")},
		{Min: 1, Max: 1},
		{Name: "nodes", Properties: map[string]interface{}{"key": "value"}},
	}

	symbols = make(map[string]*PatternSymbol)
	plan, err = pat.GetPlan(graph, symbols)
	if err != nil {
		t.Error(err)
		return
	}
	acc = &DefaultMatchAccumulator{}
	plan.Run(graph, symbols, acc)
	if len(acc.Paths) != 0 {
		t.Errorf("Expecting 0 node, got: %+v", acc)
	}

	pat = Pattern{
		{},
		{},
		{Properties: map[string]interface{}{"key": "value2"}},
	}
	if _, i := pat.getFastestElement(graph, map[string]*PatternSymbol{}); i != 2 {
		t.Errorf("Expecting 2, got %d", i)
	}
	pat = Pattern{
		{Properties: map[string]interface{}{"key": "value2"}},
		{},
		{},
	}
	if _, i := pat.getFastestElement(graph, map[string]*PatternSymbol{}); i != 0 {
		t.Errorf("Expecting 0, got %d", i)
	}
}

func TestPathPattern(t *testing.T) {
	graph, nodes := GetLineGraph(10, true)
	nodes[4].SetLabels(NewStringSet("n4"))
	nodes[5].SetLabels(NewStringSet("n5"))
	nodes[6].SetLabels(NewStringSet("n6"))

	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}

	pat := Pattern{
		{},
		{Min: 1, Max: 1},
		{Name: "bnode", Labels: NewStringSet("n5")},
		{Min: 1, Max: 1},
		{},
	}

	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 1 {
		t.Errorf("expected path of length 1, got %v", len(acc.Paths))
	}
	if acc.Paths[0].String() != "(:n4 {})->(:n5 {}) (:n5 {})->(:n6 {})" {
		t.Errorf("expected path to be (:n4 {})->(:n5 {}) (:n5 {})->(:n6 {}) got %s", acc.Paths[0].String())
	}

	// ()->(b)->()->(d)->()
	nodes[7].SetLabels(NewStringSet("n7"))
	nodes[8].SetLabels(NewStringSet("n8"))
	pat = Pattern{
		{},
		{Min: 1, Max: 1},
		{Name: "bnode", Labels: NewStringSet("n5")},
		{Min: 1, Max: 1},
		{},
		{Min: 1, Max: 1},
		{Name: "dnode", Labels: NewStringSet("n7")},
		{Min: 1, Max: 1},
		{},
	}
	symbols = make(map[string]*PatternSymbol)
	acc = &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 1 {
		t.Errorf("expected path of length 1, got %v", len(acc.Paths))
	}
	if acc.Paths[0].String() != "(:n4 {})->(:n5 {}) (:n5 {})->(:n6 {}) (:n6 {})->(:n7 {}) (:n7 {})->(:n8 {})" {
		t.Errorf("expected path to be (:n4 {})->(:n5 {}) (:n5 {})->(:n6 {}) (:n6 {})->(:n7 {}) (:n7 {})->(:n8 {}) got %s", acc.Paths[0].String())
	}

	// ()->()->()->()->(e)
	pat = Pattern{
		{},
		{Min: 1, Max: 1},
		{},
		{Min: 1, Max: 1},
		{},
		{Min: 1, Max: 1},
		{},
		{Min: 1, Max: 1},
		{Name: "enode", Labels: NewStringSet("n8")},
	}
	symbols = make(map[string]*PatternSymbol)
	acc = &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 1 {
		t.Errorf("expected path of length 1, got %v", len(acc.Paths))
	}
	if acc.Paths[0].String() != "(:n4 {})->(:n5 {}) (:n5 {})->(:n6 {}) (:n6 {})->(:n7 {}) (:n7 {})->(:n8 {})" {
		t.Errorf("expected path to be (:n4 {})->(:n5 {}) (:n5 {})->(:n6 {}) (:n6 {})->(:n7 {}) (:n7 {})->(:n8 {}) got %s", acc.Paths[0].String())
	}
}

func TestReverseSimplePath(t *testing.T) {
	graph, nodes := GetLineGraph(2, true)
	nodes[0].SetProperty("key", "value")
	nodes[1].SetProperty("key", "value")
	nodes[0].SetLabels(NewStringSet("a"))
	nodes[1].SetLabels(NewStringSet("b"))
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: -1, ToLeft: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 1 {
		t.Errorf("Expected length of path to be: %d got: %d", 1, len(acc.Paths))
	}
	if !acc.Paths[0].path[0].Reverse {
		t.Errorf("Expected path to be in reverse")
	}
}

func TestOCGetPattern(t *testing.T) {
	g := NewGraph()
	n1 := g.NewNode([]string{"root"}, nil, nil)
	n2 := g.NewNode([]string{"c1"}, nil, nil)
	n3 := g.NewNode([]string{"c2"}, nil, nil)
	n4 := g.NewNode([]string{"c3"}, nil, nil)

	g.NewEdge(n1, n2, "n1n2", nil, nil)
	g.NewEdge(n1, n3, "n1n3", nil, nil)
	g.NewEdge(n3, n4, "n3n4", nil, nil)
	pat := Pattern{
		{Name: "this", Labels: NewStringSet("c1"), Properties: map[string]interface{}{}},
		{Min: 1, Max: 1, ToLeft: true},
		{Name: "", Labels: NewStringSet(), Properties: map[string]interface{}{}},
		{Min: 1, Max: 1},
		{Name: "target", Labels: NewStringSet("c2"), Properties: map[string]interface{}{}},
	}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(g, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 1 {
		t.Errorf("Length of accumulated paths should be %d: got %d", 1, len(acc.Paths))
	}
}

func TestOCGetPattern2(t *testing.T) {
	g := NewGraph()
	n1 := g.NewNode([]string{"root"}, nil, nil)
	n2 := g.NewNode([]string{"c1"}, nil, nil)
	n3 := g.NewNode([]string{"c2"}, nil, nil)
	n4 := g.NewNode([]string{"c3"}, nil, nil)

	g.NewEdge(n1, n2, "n1n2", nil, nil)
	g.NewEdge(n1, n3, "n1n3", nil, nil)
	g.NewEdge(n3, n4, "n3n4", nil, nil)
	pat := Pattern{
		{Name: "this", Labels: NewStringSet("c1"), Properties: map[string]interface{}{}},
		{Min: 1, Max: 1, ToLeft: true},
		{Name: "", Labels: NewStringSet(), Properties: map[string]interface{}{}},
		{Min: 1, Max: 1},
		{Name: "target", Labels: NewStringSet("c2"), Properties: map[string]interface{}{}},
	}
	symbols := make(map[string]*PatternSymbol)
	symbols["this"] = &PatternSymbol{}
	symbols["this"].Add(n2)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(g, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 1 {
		t.Errorf("Length of accumulated paths should be %d: got %d", 1, len(acc.Paths))
	}
}

func TestLoopPattern(t *testing.T) {
	graph := NewGraph()
	graph.index.NodePropertyIndex("key", graph, BtreeIndex)
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < 9; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
	}
	symbols := make(map[string]*PatternSymbol)
	symbols["n"] = &PatternSymbol{}
	symbols["n"].Add(nodes[0])
	pat := Pattern{
		{Name: "n"},
		{Min: 1, Max: 1},
		{Name: "n"},
	}
	out := DefaultMatchAccumulator{}
	err := pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Symbols) > 0 {
		t.Errorf("Expecting 0 node, got: %+v", out)
	}

	// Create a loop
	graph.NewEdge(nodes[0], nodes[0], "label", nil, nil)
	out = DefaultMatchAccumulator{}
	err = pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Symbols) != 1 {
		t.Errorf("Expecting 1 node, got: %v", out)
	}
}

func TestVariableLengthPath(t *testing.T) {
	graph := NewGraph()
	graph.index.NodePropertyIndex("key", graph, BtreeIndex)
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < 9; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
	}

	nodes[1].SetProperty("property", "value")
	nodes[4].SetProperty("property", "value")

	symbols := make(map[string]*PatternSymbol)
	pat := Pattern{
		{Name: "n", Properties: map[string]interface{}{"property": "value"}},
		{Min: 1, Max: 1},
		{Properties: map[string]interface{}{"property": "value"}},
	}
	out := DefaultMatchAccumulator{}
	err := pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out.Paths) != 0 {
		t.Errorf("Expecting 0 nodes")
	}
	pat = Pattern{
		{Name: "n", Properties: map[string]interface{}{"property": "value"}},
		{Min: 1, Max: 4},
		{Properties: map[string]interface{}{"property": "value"}},
	}
	out = DefaultMatchAccumulator{}
	err = pat.Run(graph, symbols, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if out.Paths[0].NumNodes() != 4 {
		t.Errorf("Expecting 4 nodes: %+v, got num nodes: %d", out, out.Paths[0].NumNodes())
	}
}

// (n1)->(n2)->(n3)...
func GetLineGraph(n int, withIndex bool) (*Graph, []*Node) {
	graph := NewGraph()
	if withIndex {
		graph.index.NodePropertyIndex("key", graph, BtreeIndex)
	}
	nodes := make([]*Node, 0)
	for i := 0; i < n; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < n-1; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
	}
	return graph, nodes
}

// / (n1)->(n1)->(n2)->(n2)->(n3)->(n3)...
func GetLineGraphWithSelfLoops(n int, withIndex bool) (*Graph, []*Node) {
	graph := NewGraph()
	if withIndex {
		graph.index.NodePropertyIndex("key", graph, BtreeIndex)
	}
	nodes := make([]*Node, 0)
	for i := 0; i < n; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < n-1; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
		graph.NewEdge(nodes[i], nodes[i], "label", nil, nil)
	}
	graph.NewEdge(nodes[n-1], nodes[n-1], "label", nil, nil)
	return graph, nodes
}

// (n1)->(n2)->(n3)->(n1)
func GetCircleGraph(n int, withIndex bool) (*Graph, []*Node) {
	graph := NewGraph()
	if withIndex {
		graph.index.NodePropertyIndex("key", graph, BtreeIndex)
	}
	nodes := make([]*Node, 0)
	for i := 0; i < n; i++ {
		nodes = append(nodes, graph.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < n-1; i++ {
		graph.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
	}
	graph.NewEdge(nodes[n-1], nodes[0], "label", nil, nil)
	return graph, nodes
}

func TestSimpleDirectedPathPatternWithIndex(t *testing.T) {
	testSimpleDirectedPathPattern(t, true)
}
func TestSimpleDirectedPathPatternWithoutIndex(t *testing.T) {
	testSimpleDirectedPathPattern(t, false)
}

// (n:{prop:val}) -[*]-> (m:{prop:val})
func testSimpleDirectedPathPattern(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[5].SetProperty("key", "value")
	nodes[6].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: 1, Max: 1},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n5, n6 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[5] {
			n5++
		}
		if p.GetEdge(0).GetFrom() == nodes[6] {
			n6++
		}
	}
	if n5 != 1 && n6 != 1 {
		t.Errorf("Expected number of paths to be 3, got %d", n6)
	}
}

func TestSimpleDirectedPathPatternSelfLoopsWithIndex(t *testing.T) {
	testSimpleDirectedPathPatternWithSelfLoops(t, true)
}

func TestSimpleDirectedPathPatternSelfLoopsWithoutIndex(t *testing.T) {
	testSimpleDirectedPathPatternWithSelfLoops(t, false)
}

func testSimpleDirectedPathPatternWithSelfLoops(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[5].SetProperty("key", "value")
	nodes[6].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: 1, Max: 1},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 4 {
		t.Errorf("Expected number of paths to be 4, got %d", len(acc.Paths))
	}
	n5 := 0
	n6 := 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[5] {
			n5++
		}
		if p.GetEdge(0).GetFrom() == nodes[6] {
			n6++
		}
	}
	if n5 != 2 {
		t.Errorf("Expected number of paths through n5 to be 2, got %d", n5)
	}
	if n6 != 2 {
		t.Errorf("Expected number of paths through n6 to be 2, got %d", n6)
	}
}

func TestSimpleDirectedPathPatternCircleGraphWithIndex(t *testing.T) {
	testSimpleDirectedPathPatternCircleGraph(t, true)
}

func TestSimpleDirectedPathPatternCircleGraphWithoutIndex(t *testing.T) {
	testSimpleDirectedPathPatternCircleGraph(t, false)
}

func testSimpleDirectedPathPatternCircleGraph(t *testing.T, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[5].SetProperty("key", "value")
	nodes[5].SetLabels(NewStringSet("n5"))
	nodes[6].SetLabels(NewStringSet("n6"))
	nodes[6].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: 1, Max: 1},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n5, n6 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[5] {
			n5++
		}
		if p.GetEdge(0).GetFrom() == nodes[6] {
			n6++
		}
	}
	if n5 != 1 && n6 != 1 {
		t.Errorf("Expected number of paths to be 3, got %d", n6)
	}
}

func TestSimplePathPatternPatternWithIndex(t *testing.T) {
	testSimplePathPattern(t, true)
}

func TestSimplePathPatternWithoutIndex(t *testing.T) {
	testSimplePathPattern(t, false)
}

// (n:{prop:val})-[]-(m:{prop:val})
func testSimplePathPattern(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: 1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n2, n8 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetFrom() == nodes[8] {
			n8++
		}
	}
	if n2 != 1 && n8 != 1 {
		t.Errorf("Expected number of paths to be 3, got %d", n8)
	}
}

func TestSimplePathPatternWithSelfLoopsWithIndex(t *testing.T) {
	testSimplePathPatternWithSelfLoops(t, true)
}
func TestSimplePathPatternWithSelfLoopsWithoutIndex(t *testing.T) {
	testSimplePathPatternWithSelfLoops(t, false)
}

func testSimplePathPatternWithSelfLoops(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[3].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: 1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 6 {
		t.Errorf("expected length of path accumulator to be 6, got %d", len(acc.Paths))
	}
	n2, n3 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetTo() == nodes[3] {
			n3++
		}
	}
	if n2 != 3 {
		t.Errorf("Expected number of paths through n2 to be 2, got %d", n2)
	}
	if n3 != 3 {
		t.Errorf("Expected number of paths through n3 to be 2, got %d", n3)
	}
}

func TestSimplePathPatternCircleGraphWithIndex(t *testing.T) {
	testSimplePathPatternCircleGraph(t, true)
}
func TestSimplePathPatternCircleGraphWithoutIndex(t *testing.T) {
	testSimplePathPatternCircleGraph(t, false)
}

func testSimplePathPatternCircleGraph(t *testing.T, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: 1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n2, n8 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetFrom() == nodes[8] {
			n8++
		}
	}
	if n2 != 1 {
		t.Errorf("Expected number of paths to be 1, got %d", n2)
	}
	if n8 != 1 {
		t.Errorf("Expected number of paths to be 1, got %d", n8)
	}
}

func TestVariablePathPatternWithIndex(t *testing.T) {
	testVariablePathPattern(t, true)
}

func TestVariablePathPatternWithoutIndex(t *testing.T) {
	testVariablePathPattern(t, false)
}

// (n:{prop:val})-[*]-(m:{prop:val})
func testVariablePathPattern(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[3].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: -1, Undirected: true},
		{Name: "n2", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n2, n3 := 0, 0
	if len(acc.Paths) != 2 {
		t.Errorf("Expected length of paths to be 2 got %d", len(acc.Paths))
	}
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetTo() == nodes[3] {
			n3++
		}
	}
	if n2 != 2 {
		t.Errorf("Expected number of paths to be 2, got %d", n2)
	}
	if n3 != 2 {
		t.Errorf("Expected number of paths to be 2, got %d", n3)
	}
}

func TestVariablePathPatternSelfLoopsWithIndex(t *testing.T) {
	testVariablePathPatternWithSelfLoops(t, true)
}

func TestVariablePathPatternSelfLoopsWithoutIndex(t *testing.T) {
	testVariablePathPatternWithSelfLoops(t, false)
}

func testVariablePathPatternWithSelfLoops(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(4, withIndex)
	nodes[1].SetProperty("key", "value")
	nodes[1].SetLabels(NewStringSet("b"))
	nodes[2].SetLabels(NewStringSet("c"))
	nodes[2].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: -1},
		{Name: "n2", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 6 {
		t.Errorf("Expected length of paths to be 6 got %d", len(acc.Paths))
	}
	n2, n3 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[1] {
			n2++
		}
		if p.GetEdge(0).GetTo() == nodes[2] {
			n3++
		}
	}
}

func TestVariablePathPatternCircleGraphWithIndex(t *testing.T) {
	testVariablePathPatternCircleGraph(t, true)
}

func TestVariablePathPatternCircleGraphWithoutIndex(t *testing.T) {
	testVariablePathPatternCircleGraph(t, false)
}

func testVariablePathPatternCircleGraph(t *testing.T, withIndex bool) {
	graph, nodes := GetCircleGraph(4, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[3].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: -1, Max: -1, Undirected: true},
		{Name: "n2", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 8 {
		t.Errorf("Expected length of paths to be 8 got %d", len(acc.Paths))
	}
	n2, n3 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetTo() == nodes[3] {
			n3++
		}
	}
	if n2 != 4 {
		t.Errorf("Expected number of paths to be 4, got %d", n2)
	}
	if n3 != 4 {
		t.Errorf("Expected number of paths to be 4, got %d", n3)
	}
}

func TestPathLengthTwoPatternWithIndex(t *testing.T) {
	testPathLengthTwoPattern(t, true)
}

func TestPathLengthTwoPatternWithoutIndex(t *testing.T) {
	testPathLengthTwoPattern(t, false)
}

func testPathLengthTwoPattern(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[4].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: 2, Max: 2, Undirected: true},
		{Name: "n2", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n2, n4 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetTo() == nodes[3] {
			n4++
		}
	}
	if n2 != 1 && n4 != 1 {
		t.Errorf("Expected number of paths to be 1, got %d", n4)
	}
}

func TestPathLengthTwoPatternWithSelfLoopsWithIndex(t *testing.T) {
	testPathLengthTwoPatternWithSelfLoops(t, true)
}
func TestPathLengthTwoPatternWithSelfLoopsWithoutIndex(t *testing.T) {
	testPathLengthTwoPatternWithSelfLoops(t, false)
}

func testPathLengthTwoPatternWithSelfLoops(t *testing.T, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[4].SetProperty("key", "value")
	nodes[6].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: 2, Max: 2, Undirected: true},
		{Name: "n2", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	n4, n6 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[4] {
			n4++
		}
		if p.GetEdge(0).GetTo() == nodes[6] {
			n6++
		}
	}
	if n4 != 1 {
		t.Errorf("Expected number of paths through n4 to be 1, got %d", n4)
	}
	if n6 != 1 {
		t.Errorf("Expected number of paths through n6 to be 1, got %d", n6)
	}
}

func TestPathLengthTwoPatternCircleGraphWithIndex(t *testing.T) {
	testPathLengthTwoPatternCircleGraph(t, true)
}
func TestPathLengthTwoPatternCircleGraphWithoutIndex(t *testing.T) {
	testPathLengthTwoPatternCircleGraph(t, false)
}

func testPathLengthTwoPatternCircleGraph(t *testing.T, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[2].SetLabels(NewStringSet("n4"))
	nodes[4].SetLabels(NewStringSet("n7"))
	nodes[4].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}},
		{Min: 2, Max: 2, Undirected: true},
		{Name: "n2", Labels: NewStringSet(), Properties: map[string]interface{}{"key": "value"}}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	if err := pat.Run(graph, symbols, acc); err != nil {
		t.Error(err)
		return
	}
	if len(acc.Paths) != 2 {
		t.Errorf("expected length of path accumulator to be 2, got %d", len(acc.Paths))
	}
	n2, n4 := 0, 0
	for _, p := range acc.Paths {
		if p.GetEdge(0).GetFrom() == nodes[2] {
			n2++
		}
		if p.GetEdge(0).GetTo() == nodes[3] {
			n4++
		}
	}
	if n2 != 1 && n4 != 1 {
		t.Errorf("Expected number of paths to be 1, got %d", n4)
	}
}

func BenchmarkSimpleDirectedPathPatternWithIndex(b *testing.B) {
	benchmarkSimpleDirectedPathPattern(b, true)
	benchmarkSimpleDirectedPathPatternWithSelfLoops(b, true)
	benchmarkSimpleDirectedPathPatternCircleGraph(b, true)
}
func BenchmarkSimpleDirectedPathPatternWithoutIndex(b *testing.B) {
	benchmarkSimpleDirectedPathPattern(b, false)
	benchmarkSimpleDirectedPathPatternWithSelfLoops(b, false)
	benchmarkSimpleDirectedPathPatternCircleGraph(b, false)
}

func benchmarkSimpleDirectedPathPattern(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[5].SetProperty("key", "value")
	nodes[6].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: 1, Max: 1},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func benchmarkSimpleDirectedPathPatternWithSelfLoops(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: 1},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}
func benchmarkSimpleDirectedPathPatternCircleGraph(b *testing.B, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: 1},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func BenchmarkSimplePathPatternWithIndex(b *testing.B) {
	benchmarkSimplePathPattern(b, true)
	benchmarkSimplePathPatternWithSelfLoops(b, true)
	benchmarkSimplePathPatternCircleGraph(b, true)
}
func BenchmarkSimplePathPatternWithoutIndex(b *testing.B) {
	benchmarkSimplePathPattern(b, false)
	benchmarkSimplePathPatternWithSelfLoops(b, false)
	benchmarkSimplePathPatternCircleGraph(b, false)
}

func benchmarkSimplePathPattern(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: 1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func benchmarkSimplePathPatternWithSelfLoops(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: 1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}
func benchmarkSimplePathPatternCircleGraph(b *testing.B, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: 1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func BenchmarkVariablePathPatternWithIndex(b *testing.B) {
	benchmarkVariablePathPattern(b, true)
	benchmarkVariablePathPatternCircleGraph(b, true)
	benchmarkVariablePathPatternWithSelfLoops(b, true)
}
func BenchmarkVariablePathPatternWithoutIndex(b *testing.B) {
	benchmarkVariablePathPattern(b, false)
	benchmarkVariablePathPatternCircleGraph(b, false)
	benchmarkVariablePathPatternWithSelfLoops(b, false)
}

func benchmarkVariablePathPattern(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: -1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func benchmarkVariablePathPatternCircleGraph(b *testing.B, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: -1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func benchmarkVariablePathPatternWithSelfLoops(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[2].SetProperty("key", "value")
	nodes[8].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: -1, Max: -1, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func BenchmarkPathLengthTwoPatternWithIndex(b *testing.B) {
	benchmarkPathLengthTwoPattern(b, true)
	benchmarkPathLengthTwoPatternWithSelfLoops(b, true)
	benchmarkPathLengthTwoPatternCircleGraph(b, true)
}
func BenchmarkPathLengthTwoPatternWithoutIndex(b *testing.B) {
	benchmarkPathLengthTwoPattern(b, false)
	benchmarkPathLengthTwoPatternWithSelfLoops(b, false)
	benchmarkPathLengthTwoPatternCircleGraph(b, false)
}

func benchmarkPathLengthTwoPattern(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraph(10, withIndex)
	nodes[4].SetProperty("key", "value")
	nodes[7].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: 2, Max: 2, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}
func benchmarkPathLengthTwoPatternWithSelfLoops(b *testing.B, withIndex bool) {
	graph, nodes := GetLineGraphWithSelfLoops(10, withIndex)
	nodes[4].SetProperty("key", "value")
	nodes[7].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: 2, Max: 2, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}
func benchmarkPathLengthTwoPatternCircleGraph(b *testing.B, withIndex bool) {
	graph, nodes := GetCircleGraph(10, withIndex)
	nodes[4].SetProperty("key", "value")
	nodes[7].SetProperty("key", "value")
	pat := Pattern{
		{Name: "n", Labels: NewStringSet()},
		{Min: 2, Max: 2, Undirected: true},
		{Name: "n2", Labels: NewStringSet()}}
	symbols := make(map[string]*PatternSymbol)
	acc := &DefaultMatchAccumulator{}
	for n := 0; n < b.N; n++ {
		pat.Run(graph, symbols, acc)
	}
}

func TestPatternWithContexts(t *testing.T) {
	graph := NewGraph()

	// Nodes with contexts
	node1 := graph.NewNode([]string{"Person"}, map[string]interface{}{"name": "Alice"}, NewStringSet("ctxA"))
	node2 := graph.NewNode([]string{"Person"}, map[string]interface{}{"name": "Bob"}, NewStringSet("ctxA", "ctxB"))
	node3 := graph.NewNode([]string{"Location"}, map[string]interface{}{"city": "London"}, NewStringSet("ctxB"))
	node4 := graph.NewNode([]string{"Person"}, map[string]interface{}{"name": "Charlie"}, NewStringSet("ctxC"))
	node5 := graph.NewNode([]string{"Document"}, map[string]interface{}{"title": "DocX"}, NewStringSet("ctxD", "ctxE"))
	node6 := graph.NewNode([]string{"Document"}, map[string]interface{}{"title": "DocY"}, NewStringSet("ctxE"))

	// Test Case 1: Match all - single context (MatchAnyContext = false by default)
	pat1 := Pattern{
		{Name: "n1", Contexts: NewStringSet("ctxA")},
	}
	symbols1 := make(map[string]*PatternSymbol)
	acc1 := &DefaultMatchAccumulator{}
	if err := pat1.Run(graph, symbols1, acc1); err != nil {
		t.Errorf("TestPatternWithContexts Case 1 (MatchAll): Run failed: %v", err)
	}
	if len(acc1.Paths) != 2 { // Alice (ctxA), Bob (ctxA, ctxB)
		t.Errorf("TestPatternWithContexts Case 1 (MatchAll): Expected 2 paths, got %d. Paths: %v", len(acc1.Paths), acc1.Paths)
	}

	// Test Case 2: Match all - multiple contexts (MatchAnyContext = false by default)
	pat2 := Pattern{
		{Name: "n2", Contexts: NewStringSet("ctxA", "ctxB")},
	}
	symbols2 := make(map[string]*PatternSymbol)
	acc2 := &DefaultMatchAccumulator{}
	if err := pat2.Run(graph, symbols2, acc2); err != nil {
		t.Errorf("TestPatternWithContexts Case 2 (MatchAll): Run failed: %v", err)
	}
	if len(acc2.Paths) != 1 { // Only Bob (ctxA, ctxB)
		t.Errorf("TestPatternWithContexts Case 2 (MatchAll): Expected 1 path, got %d. Paths: %v", len(acc2.Paths), acc2.Paths)
	} else if acc2.Paths[0].GetNode(0) != node2 {
		t.Errorf("TestPatternWithContexts Case 2 (MatchAll): Expected node Bob, got %v", acc2.Paths[0].GetNode(0))
	}

	// Test Case 3: Match all - label and context (MatchAnyContext = false by default)
	pat3 := Pattern{
		{Name: "n3", Labels: NewStringSet("Location"), Contexts: NewStringSet("ctxB")},
	}
	symbols3 := make(map[string]*PatternSymbol)
	acc3 := &DefaultMatchAccumulator{}
	if err := pat3.Run(graph, symbols3, acc3); err != nil {
		t.Errorf("TestPatternWithContexts Case 3 (MatchAll): Run failed: %v", err)
	}
	if len(acc3.Paths) != 1 { // Only London (Location, ctxB)
		t.Errorf("TestPatternWithContexts Case 3 (MatchAll): Expected 1 path, got %d. Paths: %v", len(acc3.Paths), acc3.Paths)
	} else if acc3.Paths[0].GetNode(0) != node3 {
		t.Errorf("TestPatternWithContexts Case 3 (MatchAll): Expected node London, got %v", acc3.Paths[0].GetNode(0))
	}

	// Test Case 4: Match all - context that does not match (MatchAnyContext = false by default)
	pat4 := Pattern{
		{Name: "n4", Contexts: NewStringSet("ctxDoesNotExist")},
	}
	symbols4 := make(map[string]*PatternSymbol)
	acc4 := &DefaultMatchAccumulator{}
	if err := pat4.Run(graph, symbols4, acc4); err != nil {
		t.Errorf("TestPatternWithContexts Case 4 (MatchAll): Run failed: %v", err)
	}
	if len(acc4.Paths) != 0 {
		t.Errorf("TestPatternWithContexts Case 4 (MatchAll): Expected 0 paths, got %d", len(acc4.Paths))
	}

	// Test Case 5: Match all - properties and context (MatchAnyContext = false by default)
	pat5 := Pattern{
		{Name: "n5", Properties: map[string]interface{}{"name": "Alice"}, Contexts: NewStringSet("ctxA")},
	}
	symbols5 := make(map[string]*PatternSymbol)
	acc5 := &DefaultMatchAccumulator{}
	if err := pat5.Run(graph, symbols5, acc5); err != nil {
		t.Errorf("TestPatternWithContexts Case 5 (MatchAll): Run failed: %v", err)
	}
	if len(acc5.Paths) != 1 { // Only Alice (name:Alice, ctxA)
		t.Errorf("TestPatternWithContexts Case 5 (MatchAll): Expected 1 path, got %d. Paths: %v", len(acc5.Paths), acc5.Paths)
	} else if acc5.Paths[0].GetNode(0) != node1 {
		t.Errorf("TestPatternWithContexts Case 5 (MatchAll): Expected node Alice, got %v", acc5.Paths[0].GetNode(0))
	}

	// Test Case 6: Match all - requires context not present (MatchAnyContext = false by default)
	pat6 := Pattern{
		{Name: "n6", Labels: NewStringSet("Person"), Properties: map[string]interface{}{"name": "Alice"}, Contexts: NewStringSet("ctxB")},
	}
	symbols6 := make(map[string]*PatternSymbol)
	acc6 := &DefaultMatchAccumulator{}
	if err := pat6.Run(graph, symbols6, acc6); err != nil {
		t.Errorf("TestPatternWithContexts Case 6 (MatchAll): Run failed: %v", err)
	}
	if len(acc6.Paths) != 0 { // Alice (name:Alice, ctxA) does not have ctxB
		t.Errorf("TestPatternWithContexts Case 6 (MatchAll): Expected 0 paths, got %d", len(acc6.Paths))
	}

	// Test Case 7: Match Any - single context in pattern, node has it
	pat7 := Pattern{
		{Name: "n7", Contexts: NewStringSet("ctxC"), MatchAnyContext: true},
	}
	symbols7 := make(map[string]*PatternSymbol)
	acc7 := &DefaultMatchAccumulator{}
	if err := pat7.Run(graph, symbols7, acc7); err != nil {
		t.Errorf("TestPatternWithContexts Case 7 (MatchAny): Run failed: %v", err)
	}
	if len(acc7.Paths) != 1 { // Charlie (ctxC)
		t.Errorf("TestPatternWithContexts Case 7 (MatchAny): Expected 1 path, got %d. Paths: %v", len(acc7.Paths), acc7.Paths)
	} else if acc7.Paths[0].GetNode(0) != node4 {
		t.Errorf("TestPatternWithContexts Case 7 (MatchAny): Expected node Charlie, got %v", acc7.Paths[0].GetNode(0))
	}

	// Test Case 8: Match Any - multiple contexts in pattern, node has one of them
	// Pattern wants ctxD OR ctxA. Alice (ctxA), Bob (ctxA, ctxB), DocX (ctxD, ctxE)
	pat8 := Pattern{
		{Name: "n8", Contexts: NewStringSet("ctxD", "ctxA"), MatchAnyContext: true},
	}
	symbols8 := make(map[string]*PatternSymbol)
	acc8 := &DefaultMatchAccumulator{}
	if err := pat8.Run(graph, symbols8, acc8); err != nil {
		t.Errorf("TestPatternWithContexts Case 8 (MatchAny): Run failed: %v", err)
	}
	if len(acc8.Paths) != 3 { // Alice, Bob, DocX
		t.Errorf("TestPatternWithContexts Case 8 (MatchAny): Expected 3 paths, got %d. Paths: %v", len(acc8.Paths), acc8.Paths)
	} else {
		// Check that one of the matched nodes is node5 (DocX)
		foundNode5 := false
		for _, pth := range acc8.Paths {
			if pth.GetNode(0) == node5 {
				foundNode5 = true
				break
			}
		}
		if !foundNode5 {
			t.Errorf("TestPatternWithContexts Case 8 (MatchAny): Expected node5 (DocX) to be among the matches.")
		}
	}

	// Test Case 9: Match Any - multiple contexts in pattern, node has a different one (no match)
	// Pattern wants ctxD OR ctxA. London (ctxB)
	pat9 := Pattern{
		{Name: "n9", Labels: NewStringSet("Location"), Contexts: NewStringSet("ctxD", "ctxA"), MatchAnyContext: true},
	}
	symbols9 := make(map[string]*PatternSymbol)
	acc9 := &DefaultMatchAccumulator{}
	if err := pat9.Run(graph, symbols9, acc9); err != nil {
		t.Errorf("TestPatternWithContexts Case 9 (MatchAny): Run failed: %v", err)
	}
	if len(acc9.Paths) != 0 { // London (ctxB) does not have ctxD or ctxA
		t.Errorf("TestPatternWithContexts Case 9 (MatchAny): Expected 0 paths, got %d. Paths: %v", len(acc9.Paths), acc9.Paths)
	}

	// Test Case 10: Match Any - with properties
	// Pattern wants (name:Bob) AND (ctxD OR ctxB). Bob (name:Bob, ctxA, ctxB) -> Match
	pat10 := Pattern{
		{Name: "n10", Properties: map[string]interface{}{"name": "Bob"}, Contexts: NewStringSet("ctxD", "ctxB"), MatchAnyContext: true},
	}
	symbols10 := make(map[string]*PatternSymbol)
	acc10 := &DefaultMatchAccumulator{}
	if err := pat10.Run(graph, symbols10, acc10); err != nil {
		t.Errorf("TestPatternWithContexts Case 10 (MatchAny): Run failed: %v", err)
	}
	if len(acc10.Paths) != 1 { // Bob
		t.Errorf("TestPatternWithContexts Case 10 (MatchAny): Expected 1 path, got %d. Paths: %v", len(acc10.Paths), acc10.Paths)
	} else if acc10.Paths[0].GetNode(0) != node2 {
		t.Errorf("TestPatternWithContexts Case 10 (MatchAny): Expected node Bob, got %v", acc10.Paths[0].GetNode(0))
	}

	// Test Case 11: Match Any - empty context set in pattern (should match no nodes if interpreted strictly, or all if lenient? Let's assume strict: needs a context from the set)
	// Current GetNodeFilterFunc: if contexts != nil && contexts.Len() > 0. So empty set means the context check is skipped.
	// This behavior is consistent for both matchAll and matchAny. If context set is empty, context check is bypassed.
	pat11 := Pattern{
		{Name: "n11", Contexts: NewStringSet(), MatchAnyContext: true},
	}
	symbols11 := make(map[string]*PatternSymbol)
	acc11 := &DefaultMatchAccumulator{}
	if err := pat11.Run(graph, symbols11, acc11); err != nil {
		t.Errorf("TestPatternWithContexts Case 11 (MatchAny, Empty Contexts): Run failed: %v", err)
	}
	// Expect all nodes, as context check is skipped if PatternItem.Contexts is empty.
	// Nodes: node1, node2, node3, node4, node5, node6 (6 nodes)
	if len(acc11.Paths) != 6 {
		t.Errorf("TestPatternWithContexts Case 11 (MatchAny, Empty Contexts): Expected 6 paths, got %d. Paths: %v", len(acc11.Paths), acc11.Paths)
	}

	// Test Case 12: Match All - empty context set in pattern (MatchAnyContext = false)
	// Should also skip context check and match all nodes.
	pat12 := Pattern{
		{Name: "n12", Contexts: NewStringSet(), MatchAnyContext: false}, // Or just omit MatchAnyContext
	}
	symbols12 := make(map[string]*PatternSymbol)
	acc12 := &DefaultMatchAccumulator{}
	if err := pat12.Run(graph, symbols12, acc12); err != nil {
		t.Errorf("TestPatternWithContexts Case 12 (MatchAll, Empty Contexts): Run failed: %v", err)
	}
	if len(acc12.Paths) != 6 {
		t.Errorf("TestPatternWithContexts Case 12 (MatchAll, Empty Contexts): Expected 6 paths, got %d. Paths: %v", len(acc12.Paths), acc12.Paths)
	}

	// Test Case 13: Match Any - specifically for node6 with ctxE
	pat13 := Pattern{
		{Name: "n13", Contexts: NewStringSet("ctxE"), MatchAnyContext: true},
	}
	symbols13 := make(map[string]*PatternSymbol)
	acc13 := &DefaultMatchAccumulator{}
	if err := pat13.Run(graph, symbols13, acc13); err != nil {
		t.Errorf("TestPatternWithContexts Case 13 (MatchAny): Run failed: %v", err)
	}
	if len(acc13.Paths) != 2 { // DocX (ctxD, ctxE), DocY (ctxE)
		t.Errorf("TestPatternWithContexts Case 13 (MatchAny): Expected 2 paths, got %d. Paths: %v", len(acc13.Paths), acc13.Paths)
	} else {
		foundNode6 := false
		for _, pth := range acc13.Paths {
			if pth.GetNode(0) == node6 {
				foundNode6 = true
				break
			}
		}
		if !foundNode6 {
			t.Errorf("TestPatternWithContexts Case 13 (MatchAny): Expected node6 (DocY) to be among the matches.")
		}
	}
}
