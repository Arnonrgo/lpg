[![GoDoc](https://godoc.org/github.com/Arnonrgo/lpg?status.svg)](https://godoc.org/github.com/Arnonrgo/lpg/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/Arnonrgo/lpg)](https://goreportcard.com/report/github.com/Arnonrgo/lpg/v3)
[![Build Status](https://github.com/Arnonrgo/lpg/actions/workflows/CI.yml/badge.svg?branch=v3)](https://github.com/Arnonrgo/lpg/actions/workflows/CI.yml)
# Labeled property graphs


This Go module is modified version of the one that is part of the [Layered Schema
Architecture](https://layeredschemas.org).

Main changes vs original
* changed (most) underlying datastructures to more efficient ones (4x-10x improvement for my use case YMMV)
* added support for contexts (something in the middle between labels and properties) 
* changed the behavior of Find to work more logically (e.g. return empty on misses rather than all nodes/edges)
* properties are now converted to strings regardless of original type


This labeled property graph package implements the openCypher model of
labeled property graphs. A labeled property graph (LPG) contains nodes
and directed edges between those nodes. Every node contains:

  * Labels: Set of string tokens that usually identify the type of the
    node,
  * Properties: Key-value pairs.
  
Every edge contains:
  * A label: String token that identifies a relationship, and
  * Properties: Key-value pairs.

A `Graph` objects keeps an index of the nodes and edges included in
it. Create a graph using `NewGraph` function:

```go
g := lpg.NewGraph()
// Create two nodes
n1 := g.NewNode([]string{"label1"},map[string]any{"prop": "value1" }, nil)
n2 := g.FastNewNode(NewStringSet("label2"),map[string]any{"prop": "value2" }, NewStringSet("context1"))
// Connect the two nodes with an edge
edge:=g.NewEdge(n1,n2,"relatedTo",nil,nil)
```

The LPG library uses iterators to address nodes and edges.

```go
for nodes:=graph.GetNodes(); nodes.Next(); {
  node:=nodes.Node()
}
for edges:=graph.GetEdges(); edges.Next(); {
  edge:edges.Edge()
}
```

Every node knows its adjacent edges. 

```go
// Get outgoing edges
for edges:=node1.GetEdges(lpg.OutgoingEdge); edges.Next(); {
  edge:=edges.Edge
}

// Get all edges
for edges:=node1.GetEdges(lpg.AnyEdge); edges.Next(); {
  edge:=edges.Edge
}
```

The graph indexes nodes by label, so access to nodes using labels is
fast. You can add additional indexes on properties:

```go
g := lpg.NewGraph()
// Index all nodes with property 'prop'
g.AddNodePropertyIndex("prop")

// This access should be fast
nodes := g.GetNodesWithProperty("prop")

// This will go through all nodes
slowNodes:= g.GetNodesWithProperty("propWithoutIndex")
```

## Pattern Searches

Graph library supports searching patterns within a graph. The
following example searches for the pattern that match

```
(:label1) -[]->({prop:value})`
```

and returns the head nodes for every matching path:

```go
pattern := lpg.Pattern{ 
 // Node containing label 'label1'
 {
   Labels: lpg.NewStringSet("label1"),
 },
 // Edge of length 1
 {
   Min: 1, 
   Max: 1,
 },
 // Node with property prop=value
 {
   Properties: map[string]interface{} {"prop":"value"},
 }}
nodes, err:=pattern.FindNodes(g,nil)
```

Variable length paths are supported:

```go
pattern := lpg.Pattern{ 
 // Node containing label 'label1'
 {
   Labels: lpg.NewStringSet("label1"),
 },
 // Minimum paths of length 2, no maximum length
 {
   Min: 2, 
   Max: -1,
 },
 // Node with property prop=value
 {
   Properties: map[string]interface{} {"prop":"value"},
 }}
```

All graph nodes are under the `nodes` key as an array. The `n` key
identifies the node using a unique index. All node references in edges
use these indexes. A node may include all outgoing edges embedded in
it, or edges may be included under a separate top-level array
`edges`. If the edge is included in the node, the edge only has a `to`
field that gives the target node index as the node containing the edge
is assumed to be the source node. Edges under the top-level `edges`
array include both a `from` and a `to` index.



