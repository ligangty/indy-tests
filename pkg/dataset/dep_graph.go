/*
 *  Copyright (C) 2021-2023 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package dataset

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type DepGraph struct {
	Vertices map[string]interface{} `json:"vertices"`
	Edges    []Edge                 `json:"edges"`
}
type Edge struct {
	Source string
	Target string
}

type BuildQueue struct {
	Builds []Build
}

func (bq *BuildQueue) append(b Build) {
	bq.Builds = append(bq.Builds, b)
}

type Build struct {
	Id    string
	Items []string `yaml:"build"`
}

func (b *Build) append(item string) {
	b.Items = append(b.Items, item)
}

func getBuildQueueAsYaml(edges []Edge) string {
	fmt.Printf("Make build queue, edges.len: %d\n", len(edges))
	bQueue := getBuildQueue(edges)
	d, _ := yaml.Marshal(&bQueue.Builds)
	var s = string(d)
	fmt.Println(s)
	return s
}

func getBuildQueue(edges []Edge) BuildQueue {
	var bQueue BuildQueue
	level := 1
	roots := make(map[string]bool)
	for ok := true; ok; ok = (len(edges) > 0) {
		keys, values := getKeysAndValues(edges)
		s := make(map[string]bool)
		//fmt.Printf("Keys: %s\n", strings.Join(keys, " "))
		//fmt.Printf("level %d:\n", level)
		for index, v := range edges {
			if !contains(keys, v.Target) {
				s[v.Target] = true
				if isRoot(v.Source, values) {
					roots[v.Source] = true
				}
				edges[index].Target = "" // mark it
			}
		}
		bQueue.append(makeBuild(level, s))
		edges = cleanUp(edges)
		//fmt.Printf("After clean up, len: %d\n", len(edges))
		level += 1
	}
	//fmt.Printf("level %d:\n", level)
	bQueue.append(makeBuild(level, roots))
	return bQueue
}

func makeBuild(level int, m map[string]bool) Build {
	var rb Build
	rb.Id = "id" + fmt.Sprint(level)
	for k := range m {
		rb.append(k)
	}
	return rb
}

// A source is root if it is not in targets
func isRoot(source string, targets []string) bool {
	return !contains(targets, source)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getKeysAndValues(edges []Edge) ([]string, []string) {
	var keys []string
	var values []string
	for _, v := range edges {
		keys = append(keys, v.Source)
		values = append(values, v.Target)
	}
	return keys, values
}

// Remove edges with empty target
func cleanUp(edges []Edge) []Edge {
	var ret []Edge
	for _, v := range edges {
		if v.Target != "" {
			ret = append(ret, v)
		}
	}
	return ret
}
