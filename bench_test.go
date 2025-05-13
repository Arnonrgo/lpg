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
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const (
	benchLoadNodes = 1000
	benchLoadEdges = 2000

	// New constants for context benchmarks
	benchUniqueContexts     = 50
	benchAvgContextsPerNode = 5   // Target average, can vary slightly
	benchHubFactor          = 0.1 // 10% of nodes are hubs

	// Medium Graph Size
	benchMediumNodes = 1000
	benchMediumEdges = 5000

	// Large Graph Size
	benchLargeNodes = 10000
	benchLargeEdges = 50000

	// Node Labels
	labelUser     = "User"
	labelGroup    = "Group"
	labelDocument = "Document"
	labelPost     = "Post"

	// Edge Labels
	edgeConnectsTo = "connectsTo"
	edgeMemberOf   = "memberOf"
	edgeRelatedTo  = "relatedTo"
)

var (
	hubLabels        = []string{labelUser, labelGroup}
	spokeLabels      = []string{labelDocument, labelPost}
	benchContextPool []string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	benchContextPool = make([]string, benchUniqueContexts)
	for i := 0; i < benchUniqueContexts; i++ {
		benchContextPool[i] = "ctx" + strconv.Itoa(i)
	}
}

// createBenchmarkGraph generates a graph for benchmarking purposes.
// - numNodes, numEdges: desired number of nodes and edges.
// - numUniqueContexts: total number of unique context strings available.
// - avgContextsPerNode: target average number of contexts to assign per node.
// - hubFactor: proportion of nodes to be designated as "hubs" (higher connectivity).
func createBenchmarkGraph(numNodes, numEdges, avgContextsPerNode int, hubFactor float64, b *testing.B) *Graph {
	b.Helper()
	g := NewGraph()
	nodes := make([]*Node, numNodes)
	hubNodeIndices := make([]int, 0, int(float64(numNodes)*hubFactor))
	spokeNodeIndices := make([]int, 0, numNodes-int(float64(numNodes)*hubFactor))

	// Create Nodes
	for i := 0; i < numNodes; i++ {
		var nodeLabel string
		if rand.Float64() < hubFactor {
			nodeLabel = hubLabels[rand.Intn(len(hubLabels))]
			hubNodeIndices = append(hubNodeIndices, i)
		} else {
			nodeLabel = spokeLabels[rand.Intn(len(spokeLabels))]
			spokeNodeIndices = append(spokeNodeIndices, i)
		}

		// Assign contexts
		numCtx := avgContextsPerNode - 2 + rand.Intn(5) // Vary contexts: avg-2 to avg+2
		if numCtx < 0 {
			numCtx = 0
		}
		if numCtx > benchUniqueContexts {
			numCtx = benchUniqueContexts
		}
		contexts := NewStringSet()
		if numCtx > 0 {
			selectedCtxIndices := make(map[int]struct{})
			for len(selectedCtxIndices) < numCtx {
				selectedCtxIndices[rand.Intn(benchUniqueContexts)] = struct{}{}
			}
			for idx := range selectedCtxIndices {
				contexts.Add(benchContextPool[idx])
			}
		}
		nodes[i] = g.NewNode([]string{nodeLabel}, nil, contexts)
	}

	if len(nodes) == 0 {
		return g // Avoid panic if numNodes is 0
	}

	// Create Edges
	edgeLabelsAll := []string{edgeConnectsTo, edgeMemberOf, edgeRelatedTo}
	for i := 0; i < numEdges; i++ {
		var fromNode, toNode *Node
		// Preferential attachment to hubs
		if len(hubNodeIndices) > 0 && rand.Float64() < 0.7 { // 70% chance to involve a hub
			fromIdx := hubNodeIndices[rand.Intn(len(hubNodeIndices))]
			fromNode = nodes[fromIdx]
			if len(spokeNodeIndices) > 0 && rand.Float64() < 0.5 { // Connect hub to spoke
				toIdx := spokeNodeIndices[rand.Intn(len(spokeNodeIndices))]
				toNode = nodes[toIdx]
			} else { // Connect hub to any other node (could be another hub)
				toIdx := rand.Intn(numNodes)
				for toIdx == fromIdx { // Avoid self-loops for simplicity here
					toIdx = rand.Intn(numNodes)
				}
				toNode = nodes[toIdx]
			}
		} else { // Connect any two random distinct nodes
			fromIdx := rand.Intn(numNodes)
			toIdx := rand.Intn(numNodes)
			for toIdx == fromIdx {
				toIdx = rand.Intn(numNodes)
			}
			fromNode = nodes[fromIdx]
			toNode = nodes[toIdx]
		}
		g.NewEdge(fromNode, toNode, edgeLabelsAll[rand.Intn(len(edgeLabelsAll))], nil, nil)
	}
	return g
}

func BenchmarkPropNonExistsGraph(b *testing.B) {
	g := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h"}, nil))
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
			nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h"}, nil))
		} else {
			nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, map[string]interface{}{"a": "b", "c": "d", "e": "f", "g": "h", "z": "zz"}, nil))
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
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
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
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
	}
	labels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	for i := 0; i < len(nodes)-1; i++ {
		for _, label := range labels {
			g.NewEdge(nodes[i], nodes[i+1], label, nil, nil)
		}
	}
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
		nodes = append(nodes, g.NewNode([]string{fmt.Sprint(i)}, nil, nil))
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
	for n := 0; n < b.N; n++ {
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			for edges := node.GetEdges(OutgoingEdge); edges.Next(); {
				edges.Edge().GetProperty("z")
			}
		}
	}
}

// --- Single Node Benchmarks --- //

// --- Medium Graph Size --- //

func BenchmarkNode_LabelOnly_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_ContextOnly_MatchAll_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_ContextOnly_MatchAny_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), MatchAnyContext: true, Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_Combined_LabelContext_MatchAll_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_Combined_LabelContext_MatchAny_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), MatchAnyContext: true, Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Large Graph Size --- //

func BenchmarkNode_LabelOnly_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_ContextOnly_MatchAll_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_ContextOnly_MatchAny_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), MatchAnyContext: true, Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_Combined_LabelContext_MatchAll_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNode_Combined_LabelContext_MatchAny_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0], benchContextPool[1]), MatchAnyContext: true, Name: "n"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Two-Hop Path Benchmarks --- //

// --- Medium Graph Size --- //

func BenchmarkPath_LabelsOnly_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Labels: NewStringSet(labelDocument), Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_ContextsOnly_MatchAll_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0]), Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)}, // Edge filter remains label-based for simplicity
		PatternItem{Contexts: NewStringSet(benchContextPool[1]), Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_ContextsOnly_MatchAny_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0]), MatchAnyContext: true, Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Contexts: NewStringSet(benchContextPool[1]), MatchAnyContext: true, Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_Combined_LabelsContexts_MatchAll_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0]), Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Labels: NewStringSet(labelDocument), Contexts: NewStringSet(benchContextPool[1]), Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_Combined_LabelsContexts_MatchAny_Medium(b *testing.B) {
	g := createBenchmarkGraph(benchMediumNodes, benchMediumEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0]), MatchAnyContext: true, Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Labels: NewStringSet(labelDocument), Contexts: NewStringSet(benchContextPool[1]), MatchAnyContext: true, Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Large Graph Size --- //

func BenchmarkPath_LabelsOnly_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Labels: NewStringSet(labelDocument), Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_ContextsOnly_MatchAll_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0]), Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)}, // Edge filter remains label-based for simplicity
		PatternItem{Contexts: NewStringSet(benchContextPool[1]), Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_ContextsOnly_MatchAny_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Contexts: NewStringSet(benchContextPool[0]), MatchAnyContext: true, Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Contexts: NewStringSet(benchContextPool[1]), MatchAnyContext: true, Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_Combined_LabelsContexts_MatchAll_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0]), Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Labels: NewStringSet(labelDocument), Contexts: NewStringSet(benchContextPool[1]), Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPath_Combined_LabelsContexts_MatchAny_Large(b *testing.B) {
	g := createBenchmarkGraph(benchLargeNodes, benchLargeEdges, benchAvgContextsPerNode, benchHubFactor, b)
	pattern := Pattern{
		PatternItem{Labels: NewStringSet(labelUser), Contexts: NewStringSet(benchContextPool[0]), MatchAnyContext: true, Name: "u"},
		PatternItem{Labels: NewStringSet(edgeRelatedTo)},
		PatternItem{Labels: NewStringSet(labelDocument), Contexts: NewStringSet(benchContextPool[1]), MatchAnyContext: true, Name: "d"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pattern.FindPaths(g, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
