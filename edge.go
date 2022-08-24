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

import (
	"fmt"
)

type Edge struct {
	from, to *Node
	label    string
	Properties
	id int
}

// EdgeDir is used to show edge direction
type EdgeDir int

// Incoming and outgoing edge direction constants
const (
	IncomingEdge EdgeDir = -1
	AnyEdge      EdgeDir = 0
	OutgoingEdge EdgeDir = 1
)

func (edge *Edge) GetID() int       { return edge.id }
func (edge *Edge) GetGraph() *Graph { return edge.from.graph }
func (edge *Edge) GetLabel() string { return edge.label }
func (edge *Edge) GetFrom() *Node   { return edge.from }
func (edge *Edge) GetTo() *Node     { return edge.to }

func (edge *Edge) SetLabel(label string) {
	edge.from.graph.SetEdgeLabel(edge, label)
}

func (edge *Edge) SetProperty(key string, value interface{}) {
	edge.from.graph.SetEdgeProperty(edge, key, value)
}

func (edge *Edge) RemoveProperty(key string) {
	edge.from.graph.RemoveEdgeProperty(edge, key)
}

// Remove an edge
func (edge *Edge) Remove() {
	edge.from.graph.RemoveEdge(edge)
}

func (edge *Edge) String() string {
	return fmt.Sprintf("[:%s %s]", edge.label, edge.Properties)
}
