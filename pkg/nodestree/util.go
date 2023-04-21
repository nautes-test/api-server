// Copyright 2023 Nautes Authors
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

package nodestree

import (
	"reflect"

	structs "github.com/fatih/structs"
)

func NodesToMapping(nodes *Node, mapping map[string]*Node) {
	mapping[nodes.Path] = nodes
	for _, n := range nodes.Children {
		if n.IsDir {
			NodesToMapping(n, mapping)
		}

		mapping[n.Path] = n
	}
}

func IsInSlice(slice []string, s string) (isIn bool) {
	if len(slice) == 0 {
		return false
	}

	isIn = false
	for _, f := range slice {
		if f == s {
			isIn = true
			break
		}
	}

	return
}

func GetResourceValue(c interface{}, field, key string) string {
	t := reflect.TypeOf(c)
	t = t.Elem()
	v := reflect.ValueOf(c)
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Name == field {
			m := structs.Map(v.Field(i).Interface())
			v := m[key]

			if v != nil {
				return v.(string)
			}
		}
	}

	return ""
}

func ListsResourceNodes(nodes Node, kind string) (list []*Node) {
	for _, node := range nodes.Children {
		if node.IsDir {
			list = append(list, ListsResourceNodes(*node, kind)...)
		} else {
			if node.Kind == kind {
				list = append(list, node)
			}
		}
	}

	return list
}
