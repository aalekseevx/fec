package fecanalysis

import "container/list"

// Graph represents an abstract graph interface
type Graph interface {
	// NumVertices returns the total number of vertices in the graph
	NumVertices() int

	// GetEdges returns a list of edges from the given vertex
	// Each edge is represented as the destination vertex index
	GetEdges(vertex int) []int
}

// BFS performs breadth-first search on the given graph starting from multiple source vertices
// It returns a slice of all vertices reachable from any of the source vertices
func BFS(graph Graph, sources []int) []int {
	if len(sources) == 0 {
		return nil
	}

	// Create internal visited tracking for BFS algorithm
	visited := make([]bool, graph.NumVertices())
	var reachableVertices []int

	// Create a queue for BFS
	queue := list.New()

	// Mark all sources as visited, add to reachable list and queue
	for _, source := range sources {
		// Validate input
		if source < 0 || source >= graph.NumVertices() {
			continue
		}
		if !visited[source] {
			visited[source] = true
			reachableVertices = append(reachableVertices, source)
			queue.PushBack(source)
		}
	}

	// Process vertices in BFS order
	for queue.Len() > 0 {
		// Dequeue a vertex
		element := queue.Front()
		queue.Remove(element)
		current := element.Value.(int)

		// Get all adjacent vertices
		edges := graph.GetEdges(current)

		// Process each adjacent vertex
		for _, neighbor := range edges {
			// Skip invalid vertices
			if neighbor < 0 || neighbor >= graph.NumVertices() {
				continue
			}

			// If not yet visited, mark as visited, add to reachable list, then enqueue
			if !visited[neighbor] {
				visited[neighbor] = true
				reachableVertices = append(reachableVertices, neighbor)
				queue.PushBack(neighbor)
			}
		}
	}

	return reachableVertices
}
