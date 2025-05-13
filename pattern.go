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
	"math"
)

type ErrNodeVariableExpected string

func (e ErrNodeVariableExpected) Error() string {
	return "Node variable expected: " + string(e)
}

type ErrEdgeVariableExpected string

func (e ErrEdgeVariableExpected) Error() string {
	return "Edge variable expected:" + string(e)
}

// Pattern contains pattern items, with even numbered elements
// corresponding to nodes, and odd numbered elements corresponding to
// edges
type Pattern []PatternItem

// A PatternSymbol contains either nodes, or edges.
type PatternSymbol struct {
	Nodes *NodeSet
	Edges *EdgeSet
}

// A PatternItem can be a node or an edge element of a pattern
// It specifies the labels, contexts (for nodes), and properties to match.
// For edges, it also defines the direction and cardinality (Min, Max hops).
// The MatchAnyContext field dictates how contexts are matched: if false (default),
// all specified contexts must be present; if true, at least one must be present.
type PatternItem struct {
	Labels          *StringSet
	Contexts        *StringSet // Contexts to match for node items. Ignored for edge items.
	MatchAnyContext bool       // If true, match if any context in Contexts is present. If false (default), all must be present.
	Properties      map[string]interface{}
	// Min=-1 and Max=-1 for variable length
	Min int
	Max int
	// ToLeft is true if this is a relationship of the form <-
	ToLeft bool
	// Undirected is true if this is a relationship of the form --
	Undirected bool
	// Name of the variable associated with this processing node. If the
	// name is defined, it is used to constrain values. If not, it is
	// used to store values
	Name string
}

func (p PatternItem) getEdgeFilter() func(*Edge) bool {
	return GetEdgeFilterFunc(p.Labels, p.Properties)
}

func (p PatternItem) getNodeFilter() func(*Node) bool {
	return GetNodeFilterFunc(p.Labels, p.Contexts, p.Properties, p.MatchAnyContext)
}

// Returns the set of nodes constraining the pattern item. That is,
// the set of nodes in the symbols
func (p PatternItem) isConstrainedNodes(ctx *MatchContext) (*NodeSet, error) {
	if len(p.Name) == 0 {
		return nil, nil
	}
	// Is this a symbol created in this pattern?
	sym, exists := ctx.LocalSymbols[p.Name]
	if exists {
		if sym.Edges != nil {
			return nil, ErrNodeVariableExpected(p.Name)
		}
		return sym.Nodes, nil
	}
	sym, exists = ctx.Symbols[p.Name]
	if exists {
		if sym.Edges != nil {
			return nil, ErrNodeVariableExpected(p.Name)
		}
		return sym.Nodes, nil
	}
	return nil, nil
}

// Returns the set of edges constraining the pattern item. That is,
// the set of edges in the symbols
func (p PatternItem) isConstrainedEdges(ctx *MatchContext) (*EdgeSet, error) {
	if len(p.Name) == 0 {
		return nil, nil
	}
	// Is this a symbol created in this pattern?
	sym, exists := ctx.LocalSymbols[p.Name]
	if exists {
		if sym.Nodes != nil {
			return nil, ErrEdgeVariableExpected(p.Name)
		}
		return sym.Edges, nil
	}
	sym, exists = ctx.Symbols[p.Name]
	if exists {
		if sym.Nodes != nil {
			return nil, ErrEdgeVariableExpected(p.Name)
		}
		return sym.Edges, nil
	}
	return nil, nil
}

func (p *PatternItem) estimateNodeSize(g *Graph, symbols map[string]*PatternSymbol) (NodeIterator, int) {
	max := -1
	var ret NodeIterator

	// If labels are present, use label index
	if p.Labels != nil && p.Labels.Len() > 0 {
		itr := g.index.nodesByLabel.IteratorAllLabels(p.Labels)
		if sz := itr.MaxSize(); sz != -1 {
			max = sz
			ret = itr
		}
	}

	// If properties are present, use property index
	// This can refine the estimate if properties are more restrictive or if no labels were given
	if p.Properties != nil && len(p.Properties) > 0 {
		for k, v := range p.Properties {
			prop := fmt.Sprintf("%v", v)
			itr, _ := g.index.GetIteratorForNodeProperty(k, prop)
			if itr == nil {
				continue
			}
			maxSize := itr.MaxSize()
			if maxSize == -1 { // Cannot determine size from this property iterator
				continue
			}
			if max == -1 || maxSize < max {
				max = maxSize
				ret = itr
			}
		}
	}

	// NEW: If contexts are present, use context index.
	// This can refine the estimate further if contexts are more restrictive
	// or if no labels/properties were given.
	// This estimation is only performed if there are no labels and no properties,
	// or if context estimation is more restrictive.
	// For now, let's place it to run if no specific iterator has been chosen yet (ret == nil) OR
	// if it can provide a better estimate.
	// To be more precise, we calculate context estimate and then compare.
	if p.Contexts != nil && p.Contexts.Len() > 0 && g.index.nodesByContext != nil {
		contextEstimateMax := -1
		var contextEstimateRet NodeIterator

		if p.MatchAnyContext {
			iteratorsForReturn := make([]Iterator, 0, p.Contexts.Len())
			estimatedSizeSum := 0
			hasAnyValidIter := false

			for _, contextValue := range p.Contexts.Slice() {
				foundIter := g.index.nodesByContext.find(contextValue) // Returns Iterator
				if foundIter != nil {
					iteratorsForReturn = append(iteratorsForReturn, foundIter)
					currentSize := foundIter.MaxSize()
					if currentSize > 0 {
						estimatedSizeSum += currentSize
					}
					// Whether currentSize is 0 or >0, if we found an iterator for a context, it's a valid part of MatchAny processing.
					hasAnyValidIter = true
				}
			}

			if !hasAnyValidIter || len(iteratorsForReturn) == 0 {
				contextEstimateMax = 0
				contextEstimateRet = &nodeIterator{emptyIterator{}}
			} else {
				contextEstimateMax = estimatedSizeSum // Overestimate, but preserves iterator integrity
				multiItRet := MultiIterator(iteratorsForReturn...)
				uniqueItRet := makeUniqueIterator(multiItRet)
				contextEstimateRet = nodeIterator{uniqueItRet}
			}
		} else { // MatchAll contexts
			minIterSize := math.MaxInt32
			var bestIterForMatchAll NodeIterator
			foundValidContext := false

			for _, contextValue := range p.Contexts.Slice() {
				genericIter := g.index.nodesByContext.find(contextValue) // Returns Iterator

				// Wrap the generic Iterator to make it a NodeIterator for use in this block.
				// If genericIter is emptyIterator{}, nodeIterator{genericIter}.MaxSize() will be 0.
				wrappedNodeIter := nodeIterator{genericIter}

				// if nodeIter == nil { ... } // Original check based on nil, now use MaxSize or type assertion if needed
				// For robustness, explicitly handle if find returns an iterator that effectively means 'not found' or 'empty'.
				// An emptyIterator will have MaxSize 0. A valid iterator from find() also reports MaxSize.
				if wrappedNodeIter.MaxSize() == 0 { // Check MaxSize on the wrapped iterator
					minIterSize = 0
					bestIterForMatchAll = &nodeIterator{emptyIterator{}} // Consistent empty iterator
					foundValidContext = true                             // Mark as found to use this 0 result
					break
				}

				currentIterSize := wrappedNodeIter.MaxSize()
				// Note: Original code had a check: if currentIterSize == 0 after nodeIter != nil.
				// This is now covered by wrappedNodeIter.MaxSize() == 0 check above.

				if currentIterSize < minIterSize {
					minIterSize = currentIterSize
					bestIterForMatchAll = wrappedNodeIter // Assign the NodeIterator compliant wrappedNodeIter
				}
				foundValidContext = true
			}

			if foundValidContext {
				contextEstimateMax = minIterSize
				contextEstimateRet = bestIterForMatchAll
			} else { // Should not happen if p.Contexts.Len() > 0, but as a safeguard
				contextEstimateMax = 0
				contextEstimateRet = &nodeIterator{emptyIterator{}}
			}
		}

		// Compare context-based estimate with any prior estimate (label/property)
		if contextEstimateRet != nil && (ret == nil || (contextEstimateMax != -1 && contextEstimateMax < max)) {
			max = contextEstimateMax
			ret = contextEstimateRet
		}
	}

	// If a variable name is defined, it further constrains the selection
	if len(p.Name) > 0 {
		sym, ok := symbols[p.Name]
		if ok {
			if sym.Nodes == nil { // Variable is bound, but to an empty set of nodes
				max = 0
				ret = &nodeIterator{emptyIterator{}}
			} else if ret == nil || (sym.Nodes.Len() < max) { // Variable is more restrictive
				max = sym.Nodes.Len()
				ret = sym.Nodes.Iterator()
			}
		}
	}

	// If no specific iterator was chosen by labels, properties, contexts, or bound variable
	if ret == nil {
		ret = g.GetNodes()
		// max = ret.MaxSize() // Avoid re-assigning max if it's already -1 from ret.MaxSize()
		if sz := ret.MaxSize(); sz != -1 {
			max = sz
		} else {
			max = -1 // Ensure max is -1 if GetNodes().MaxSize() is -1
		}
	}
	return ret, max
}

func (p PatternItem) estimateEdgeSize(g *Graph, symbols map[string]*PatternSymbol) (EdgeIterator, int) {
	max := -1
	var ret EdgeIterator

	allEdges := func() (EdgeIterator, int) {
		ret := g.GetEdges()
		max := -1
		if sz := ret.MaxSize(); sz != -1 {
			max = sz
		}
		return ret, max
	}

	if p.Min > 1 || p.Max > 1 || p.Min == -1 || p.Max == -1 {
		return g.GetEdges(), -1
	}

	if p.Labels != nil && p.Labels.Len() > 0 {
		itr := g.GetEdgesWithAnyLabel(p.Labels)
		if sz := itr.MaxSize(); sz != -1 {
			max = sz
			ret = itr
		}
	}
	if p.Properties != nil && len(p.Properties) > 0 {
		for k, v := range p.Properties {
			prop := fmt.Sprintf("%v", v)
			itr, _ := g.index.GetIteratorForEdgeProperty(k, prop)
			if itr == nil {
				continue
			}
			maxSize := itr.MaxSize()
			if maxSize == -1 {
				continue
			}
			if maxSize < max {
				max = maxSize
				ret = itr
			}
		}
	}
	if len(p.Name) > 0 {
		sym, ok := symbols[p.Name]
		if ok {
			if sym.Edges == nil {
				max = 0
				ret = &edgeIterator{emptyIterator{}}
			} else if max == -1 || sym.Edges.Len() < max {
				max = sym.Edges.Len()
				ret = sym.Edges.Iterator()
			}
		}
	}
	if ret == nil {
		return allEdges()
	}
	return ret, max
}

func (p *PatternSymbol) Add(item interface{}) bool {
	switch k := item.(type) {
	case *Node:
		p.AddNode(k)

	case *Edge:
		if p.Edges == nil {
			p.Edges = NewEdgeSet()
		}
		p.Edges.Add(k)

	case *Path:
		p.AddPath(k)
	}
	return true
}

func (p *PatternSymbol) AddNode(item *Node) {
	if p.Nodes == nil {
		p.Nodes = NewNodeSet()
	}
	p.Nodes.Add(item)
}

func (p *PatternSymbol) AddPath(path *Path) {
	if p.Edges == nil {
		p.Edges = NewEdgeSet()
	}
	for _, x := range path.path {
		p.Edges.Add(x.Edge)
	}
}

func (p *PatternSymbol) NodeSlice() []*Node {
	if p.Nodes != nil {
		return p.Nodes.Slice()
	}
	return nil
}

func (p *PatternSymbol) EdgeSlice() *Path {
	if p.Edges != nil {
		path := &Path{path: make([]PathElement, 0)}
		for itr := p.Edges.Iterator(); itr.Next(); {
			path.path = append(path.path, PathElement{Edge: itr.Edge()})
		}
		return path
	}
	return nil
}

type MatchPlan struct {
	steps    []planProcessor
	nForward int
}

type planProcessor interface {
	Run(*MatchContext, matchAccumulator) error
	GetResult() interface{}
	GetPatternItem() PatternItem
}

type MatchAccumulator interface {
	// path is either a Node or []Edge, the matching path symbols
	// contains the current values for each symbol. The values of the
	// map is either Node or []Edge
	StoreResult(ctx *MatchContext, path *Path, symbols map[string]interface{})
}

type MatchContext struct {
	Graph *Graph
	// These are symbols that are used as constraints in the matching process.
	Symbols map[string]*PatternSymbol

	// localSymbols are symbols defined in the pattern.
	LocalSymbols map[string]*PatternSymbol

	variablePathNode *Node
}

// If the current step has a local symbol, it will be recorded in the context
func (ctx *MatchContext) recordStepResult(step planProcessor) {
	name := step.GetPatternItem().Name
	if len(name) == 0 {
		return
	}
	if _, global := ctx.Symbols[name]; global {
		return
	}
	result := &PatternSymbol{}
	result.Add(step.GetResult())
	ctx.LocalSymbols[name] = result
}

// resetStepResult will remove the step's local symbol from the context
func (ctx *MatchContext) resetStepResult(step planProcessor) {
	name := step.GetPatternItem().Name
	if len(name) == 0 {
		return
	}
	if _, global := ctx.Symbols[name]; global {
		return
	}
	delete(ctx.LocalSymbols, name)
}

func (pattern Pattern) Run(graph *Graph, symbols map[string]*PatternSymbol, result MatchAccumulator) error {
	plan, err := pattern.GetPlan(graph, symbols)
	if err != nil {
		return err
	}
	logf("Starting plan run\n")
	return plan.Run(graph, symbols, result)
}

func (pattern Pattern) FindPaths(graph *Graph, symbols map[string]*PatternSymbol) (DefaultMatchAccumulator, error) {
	acc := DefaultMatchAccumulator{}
	if err := pattern.Run(graph, symbols, &acc); err != nil {
		return acc, err
	}
	return acc, nil
}

// FindNodes runs the pattern with the given symbols, and returns all the head nodes found
func (pattern Pattern) FindNodes(graph *Graph, symbols map[string]*PatternSymbol) ([]*Node, error) {
	acc := DefaultMatchAccumulator{}
	if err := pattern.Run(graph, symbols, &acc); err != nil {
		return nil, err
	}
	return acc.GetHeadNodes(), nil
}

func (pattern Pattern) getFastestElement(graph *Graph, symbols map[string]*PatternSymbol) (Iterator, int) {
	maxSize := -1
	index := 0
	var itr Iterator
	for i := range pattern {
		var sz int
		var t Iterator
		if (i % 2) == 0 {
			t, sz = pattern[i].estimateNodeSize(graph, symbols)
		} else {
			t, sz = pattern[i].estimateEdgeSize(graph, symbols)
		}
		if sz != -1 {
			if maxSize == -1 || sz < maxSize {
				maxSize = sz
				index = i
				itr = t
			}
		}
	}
	return itr, index
}

func (pattern Pattern) GetSymbolNames() *StringSet {
	ret := NewStringSet()
	for _, p := range pattern {
		if len(p.Name) > 0 {
			ret.Add(p.Name)
		}
	}
	return ret
}

// GetPlan returns a match execution plan
func (pattern Pattern) GetPlan(graph *Graph, symbols map[string]*PatternSymbol) (MatchPlan, error) {
	itr, index := pattern.getFastestElement(graph, symbols)
	plan := MatchPlan{}
	processors := make([]planProcessor, len(pattern))
	if (index % 2) == 0 {
		// start with a node
		processors[index] = &iterateNodes{itr: itr.(NodeIterator), patternItem: pattern[index]}
		plan.steps = append(plan.steps, processors[index])
		// Go forward
		for i := index + 1; i < len(pattern); i++ {
			if (i % 2) == 1 {
				// Patern is an edge
				// There is a node before this edge.
				if pattern[i].ToLeft {
					// n<--
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], IncomingEdge)
				} else if !pattern[i].Undirected {
					// n-->
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], OutgoingEdge)
				} else {
					// n--
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], AnyEdge)
				}
			} else {
				// Pattern is a node
				// There is an edge before this node, and that determines the direction
				if pattern[i-1].ToLeft {
					// <--n
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useFromNode)
				} else if !pattern[i-1].Undirected {
					// -->n
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useToNode)
				} else {
					// --n
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useAnyNode, processors[i-2])
				}
			}
			plan.steps = append(plan.steps, processors[i])
			plan.nForward++
		}
		// Go backwards
		for i := index - 1; i >= 0; i-- {
			if (i % 2) == 1 {
				// There is a node after this edge
				if pattern[i].ToLeft {
					// <--n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], OutgoingEdge)
				} else if !pattern[i].Undirected {
					// -->n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], IncomingEdge)
				} else {
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], AnyEdge)
				}
			} else {
				// There is an edge after this node, and that determines the direction
				if pattern[i+1].ToLeft {
					// n<--
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useToNode)
				} else if !pattern[i+1].Undirected {
					// n-->
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useFromNode)
				} else {
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useAnyNode, processors[i+2])
				}
			}
			plan.steps = append(plan.steps, processors[i])
		}
	} else {
		// start with an edge
		processors[index] = &iterateEdges{itr: itr.(EdgeIterator), patternItem: pattern[index]}
		plan.steps = append(plan.steps, processors[index])
		// Go forward
		for i := index + 1; i < len(pattern); i++ {
			if (i % 2) == 1 {
				// There is a node before this edge
				if pattern[i].ToLeft {
					// n<--
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], IncomingEdge)
				} else if !pattern[i].Undirected {
					// n-->
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], OutgoingEdge)
				} else {
					processors[i] = newIterateConnectedEdges(processors[i-1], pattern[i], AnyEdge)
				}
			} else {
				// There is an edge before this node, and that determines the direction
				if pattern[i-1].ToLeft {
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useFromNode)
				} else if !pattern[i-1].Undirected {
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useToNode)
				} else {
					processors[i] = newIterateConnectedNodes(processors[i-1], pattern[i], useAnyNode, processors[i-2])
				}
			}
			plan.steps = append(plan.steps, processors[i])
			plan.nForward++
		}
		// Go backwards
		for i := index - 1; i >= 0; i-- {
			if (i % 2) == 1 {
				// There is a node after this edge
				if pattern[i].ToLeft {
					// <--n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], OutgoingEdge)
				} else if !pattern[i].Undirected {
					// -->n
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], IncomingEdge)
				} else {
					processors[i] = newIterateConnectedEdges(processors[i+1], pattern[i], AnyEdge)
				}
			} else {
				// There is an edge after this node, and that determines the direction
				if pattern[i+1].ToLeft {
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useToNode)
				} else if !pattern[i+1].Undirected {
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useFromNode)
				} else {
					processors[i] = newIterateConnectedNodes(processors[i+1], pattern[i], useAnyNode, processors[i+2])
				}
			}
			plan.steps = append(plan.steps, processors[i])
		}
	}
	return plan, nil
}

type matchAccumulator interface {
	Run(*MatchContext) error
}

type nextAccumulator struct {
	run  planProcessor
	next matchAccumulator
}

func (n nextAccumulator) Run(ctx *MatchContext) error {
	return n.run.Run(ctx, n.next)
}

type resultAccumulator struct {
	acc  MatchAccumulator
	plan MatchPlan
}

// Capture the current results
func (n *resultAccumulator) Run(ctx *MatchContext) error {
	n.acc.StoreResult(ctx, n.plan.GetCurrentPath(), n.plan.CaptureSymbolValues())
	return nil
}

// GetCurrentPath returns the current path recoded in the stages of the pattern. The result is either a single node, or a path
func (plan MatchPlan) GetCurrentPath() *Path {
	if len(plan.steps) == 1 {
		if path, ok := plan.steps[0].GetResult().(*Path); ok {
			return path
		}
		return &Path{only: plan.steps[0].GetResult().(*Node)}
	}
	out := &Path{}
	for i := range plan.steps {
		if i > plan.nForward {
			break
		}
		if path, ok := plan.steps[i].GetResult().(*Path); ok {
			out.AppendPath(path)
		}
	}
	revPathOut := &Path{}
	for i := len(plan.steps) - 1; i >= plan.nForward; i-- {
		if path, ok := plan.steps[i].GetResult().(*Path); ok {
			revPathOut.Append(PathElement{Edge: path.path[0].Edge, Reverse: !path.path[0].Reverse})
			// revPathOut.AppendPath(path)
		}
	}
	return revPathOut.AppendPath(out)
}

// CaptureSymbolValues captures the current symbol values as nodes or []Edges
func (plan MatchPlan) CaptureSymbolValues() map[string]interface{} {
	ret := make(map[string]interface{})
	for _, step := range plan.steps {
		if len(step.GetPatternItem().Name) > 0 {
			if _, exists := ret[step.GetPatternItem().Name]; !exists {
				ret[step.GetPatternItem().Name] = step.GetResult()
			}
		}
	}
	return ret
}

type DefaultMatchAccumulator struct {
	// Each element of the paths is either a Node or []Edge
	Paths   []*Path
	Symbols []map[string]interface{}
}

func (acc *DefaultMatchAccumulator) StoreResult(_ *MatchContext, path *Path, symbols map[string]interface{}) {
	// set := sm.SliceMap[*Path, struct{}]{}
	// for _, p := range acc.Paths {
	// 	set.Put([]*Path{{path: []PathElement{{Edge: p.GetEdge(0)}}}}, struct{}{})
	// }
	// ps := make([]*Path, 0)
	// ps = append(ps, path.(*Path))
	// _, seen := set.Get(ps)
	// if !seen {
	acc.Paths = append(acc.Paths, path)
	// }
	acc.Symbols = append(acc.Symbols, symbols)
}

// Returns the unique nodes in the accumulator that start a path
func (acc *DefaultMatchAccumulator) GetHeadNodes() []*Node {
	ret := make(map[*Node]struct{})
	for _, x := range acc.Paths {
		ret[x.GetNode(0)] = struct{}{}
	}
	arr := make([]*Node, 0, len(ret))
	for x := range ret {
		arr = append(arr, x)
	}
	return arr
}

// Returns the unique nodes in the accumulator that ends a path
func (acc *DefaultMatchAccumulator) GetTailNodes() []*Node {
	ret := make(map[*Node]struct{})
	for _, x := range acc.Paths {
		ret[x.GetNode(x.NumNodes()-1)] = struct{}{}
	}
	arr := make([]*Node, 0, len(ret))
	for x := range ret {
		arr = append(arr, x)
	}
	return arr
}

func (plan MatchPlan) Run(graph *Graph, symbols map[string]*PatternSymbol, result MatchAccumulator) error {
	ctx := &MatchContext{
		Graph:        graph,
		Symbols:      symbols,
		LocalSymbols: make(map[string]*PatternSymbol),
	}

	res := resultAccumulator{acc: result, plan: plan}
	acc := matchAccumulator(&res)
	for i := len(plan.steps) - 1; i > 0; i-- {
		acc = nextAccumulator{
			run:  plan.steps[i],
			next: acc,
		}
	}
	logf("Plan run: steps: %+v, ctx: %+v\n", plan.steps, ctx)
	return plan.steps[0].Run(ctx, acc)
}
