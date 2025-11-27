package common

import (
	"github.com/google/uuid"
)

func HasCycle(graph map[uuid.UUID]map[uuid.UUID]bool) (bool, []uuid.UUID) {
	if len(graph) == 0 {
		return false, nil
	}

	path := []uuid.UUID{}
	visited := make(map[uuid.UUID]GraphNodeStatus)
	for node := range graph {
		visited[node] = GraphNodeStatusNotVisited
	}

	for node := range graph {
		if visited[node] == GraphNodeStatusNotVisited {
			hasCycle, cyclePath := dfs(node, graph, visited, path)
			if hasCycle {
				return true, cyclePath
			}
		}
	}

	return false, nil
}

func dfs(nodeID uuid.UUID, graph map[uuid.UUID]map[uuid.UUID]bool, visited map[uuid.UUID]GraphNodeStatus, path []uuid.UUID) (bool, []uuid.UUID) {
	path = append(path, nodeID)
	visited[nodeID] = GraphNodeStatusVisiting
	for edge := range graph[nodeID] {
		if visited[edge] == GraphNodeStatusVisiting {
			path = append(path, edge)
			return true, path
		}
		if visited[edge] == GraphNodeStatusNotVisited {
			hasCycle, cyclePath := dfs(edge, graph, visited, path)
			if hasCycle {
				return true, cyclePath
			}
		}
	}
	visited[nodeID] = GraphNodeStatusVisited
	return false, nil
}
