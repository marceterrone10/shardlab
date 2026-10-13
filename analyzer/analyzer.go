// Package analyzer turns observed joins into candidate co-location groups.
package analyzer

import (
	"fmt"
	"sort"
	"strings"
)

type JoinObservation struct {
	LeftTable  string `json:"left_table"`
	RightTable string `json:"right_table"`
	Count      int64  `json:"count"`
}

type Suggestion struct {
	Tables []string `json:"tables"`
	Weight int64    `json:"weight"`
}

type pair struct {
	left  string
	right string
}

// Suggest groups tables connected by joins whose aggregate frequency reaches
// minFrequency. It is deliberately small: later phases can replace this
// connected-components heuristic while preserving the input/output contract.
func Suggest(observations []JoinObservation, minFrequency int64) ([]Suggestion, error) {
	if minFrequency < 1 {
		return nil, fmt.Errorf("minimum frequency must be greater than zero")
	}

	weights := make(map[pair]int64)
	tables := make(map[string]struct{})
	for _, observation := range observations {
		left := strings.TrimSpace(observation.LeftTable)
		right := strings.TrimSpace(observation.RightTable)
		if left == "" || right == "" {
			return nil, fmt.Errorf("table names cannot be empty")
		}
		if observation.Count < 1 {
			return nil, fmt.Errorf("observation count must be greater than zero")
		}
		if left == right {
			continue
		}
		if left > right {
			left, right = right, left
		}
		weights[pair{left: left, right: right}] += observation.Count
		tables[left] = struct{}{}
		tables[right] = struct{}{}
	}

	parent := make(map[string]string, len(tables))
	for table := range tables {
		parent[table] = table
	}

	var find func(string) string
	find = func(table string) string {
		if parent[table] != table {
			parent[table] = find(parent[table])
		}
		return parent[table]
	}
	union := func(left, right string) {
		leftRoot := find(left)
		rightRoot := find(right)
		if leftRoot != rightRoot {
			parent[rightRoot] = leftRoot
		}
	}

	for joinedTables, weight := range weights {
		if weight >= minFrequency {
			union(joinedTables.left, joinedTables.right)
		}
	}

	groupTables := make(map[string][]string)
	for table := range tables {
		root := find(table)
		groupTables[root] = append(groupTables[root], table)
	}

	groupWeights := make(map[string]int64)
	for joinedTables, weight := range weights {
		if weight < minFrequency {
			continue
		}
		root := find(joinedTables.left)
		groupWeights[root] += weight
	}

	suggestions := make([]Suggestion, 0, len(groupTables))
	for root, groupedTables := range groupTables {
		if len(groupedTables) < 2 {
			continue
		}
		sort.Strings(groupedTables)
		suggestions = append(suggestions, Suggestion{
			Tables: groupedTables,
			Weight: groupWeights[root],
		})
	}

	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Weight == suggestions[j].Weight {
			return strings.Join(suggestions[i].Tables, ",") < strings.Join(suggestions[j].Tables, ",")
		}
		return suggestions[i].Weight > suggestions[j].Weight
	})
	return suggestions, nil
}
