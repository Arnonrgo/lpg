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
	"context"
	"reflect"
	"testing"
)

func BenchmarkClone(b *testing.B) {
	source := NewGraph()
	target := NewGraph()
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, source.NewNode([]string{"a"}, nil, nil))
	}
	for i := 0; i < 9; i++ {
		source.NewEdge(nodes[i], nodes[i+1], "label", nil, nil)
	}

	for n := 0; n < b.N; n++ {
		CopyGraph(source, target, func(key string, value interface{}) interface{} {
			return value
		})
	}
}

func TestClone(t *testing.T) {
	source := NewGraph()
	target := NewGraph() // target graph has empty strtable
	nodes := make([]*Node, 0)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, source.NewNode([]string{"a"}, map[string]interface{}{"key": i}, nil))
	}
	for i := 0; i < 9; i++ {
		source.NewEdge(nodes[i], nodes[i+1], "label", map[string]interface{}{"key": i}, nil)
	}

	CopyGraph(source, target, func(key string, value interface{}) interface{} {
		return value
	})

	if ok, _ := CheckIsomorphism(context.Background(), source, target, func(n1, n2 *Node) bool {
		result := n1.GetLabels().HasAll(n2.GetLabels().Slice()...) && reflect.DeepEqual(n1.properties, n2.properties)
		return result
	},
		func(e1, e2 *Edge) bool {
			result := e1.label == e2.label && reflect.DeepEqual(e1.properties, e2.properties)
			return result
		}); !ok {
		t.Errorf("Clone result not isomorphic")
	}
}
