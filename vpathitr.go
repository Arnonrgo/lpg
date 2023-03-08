// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lpg

// CollectAllPaths iterates the variable length paths that have the
// edges in firstLeg. For each edge, it calls the edgeFilter
// function. If the edge is accepted, it recursively descends and
// calls accumulator.AddPath for each discovered path until AddPath
// returns false
func CollectAllPaths(graph *Graph, fromNode *Node, firstLeg EdgeIterator, edgeFilter func(*Edge) bool, dir EdgeDir, min, max int, accumulator func([]*Edge, *Node) bool) {
	var recurse func([]*Edge, *Node) bool
	// isLoop := func(path Path) {

	// }
	// if appending nextpath to path - is loop formed (going around twice)?
	isLoop := func(path []PathElement, nextPath PathElement) bool {
		var lastOccurrenceIdx int
		var loopCount int
		p := NewPathFromElements(path...)
		for i := 0; i < p.NumNodes(); i++ {
			node := p.GetNode(i)
			if nextPath.GetTargetNode() == node {
				lastOccurrenceIdx = i
				loopCount++
			}
		}
		if loopCount < 2 {
			return false
		}
		// ShortPath: path starting at node lastOccurenceIx + nextEdge
		// Long path: path starting at node i
		shortPath := &Path{path: make([]PathElement, 0)}
		for i := 0; i < p.NumNodes(); i++ {
			node := p.GetNode(i)
			// pontentially a loop
			if nextPath.GetTargetNode() == node {
				if i == lastOccurrenceIdx {
					return false
				}
				return p.Slice(i, -1).HasPrefix(append(shortPath.path[lastOccurrenceIdx:], nextPath))
			}
		}
		return false
	}
	recurse = func(prefix []*Edge, lastNode *Node) bool {
		var endNode *Node
		// var endEdge *Edge
		switch dir {
		case OutgoingEdge:
			endNode = prefix[len(prefix)-1].GetTo()
			// endEdge = prefix[len(prefix)-1]
		case IncomingEdge:
			endNode = prefix[len(prefix)-1].GetFrom()
			// endEdge = prefix[len(prefix)-1]
		case AnyEdge:
			if prefix[len(prefix)-1].GetTo() == lastNode {
				endNode = prefix[len(prefix)-1].GetFrom()
			} else {
				endNode = prefix[len(prefix)-1].GetTo()
			}
			// endEdge = prefix[len(prefix)-1]
		}

		if (min == -1 || len(prefix) >= min) && (max == -1 || len(prefix) <= max) {
			// A valid path
			entry := make([]*Edge, len(prefix))
			copy(entry, prefix)
			if !accumulator(entry, endNode) {
				return false
			}
		}

		if max != -1 && len(prefix) >= max {
			return true
		}

		// if isLoop(endEdge, prefix[:len(prefix)-1], endNode) {
		// 	return true
		// }
		itr := edgeIterator{
			&filterIterator{
				itr: endNode.GetEdges(dir),
				filter: func(item interface{}) bool {
					return edgeFilter(item.(*Edge))
				},
			},
		}

		for itr.Next() {
			edge := itr.Edge()
			if isLoop(NewPathElementsFromEdges(prefix), PathElement{Edge: edge}) {
				return true
			}
			if !recurse(append(prefix, edge), endNode) {
				return false
			}
		}
		return true
	}

	for firstLeg.Next() {
		edge := firstLeg.Edge()
		if !recurse([]*Edge{edge}, fromNode) {
			break
		}
	}
}
