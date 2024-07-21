package main

import (
	"crypto/sha512"
	"math/big"

	graph "github.com/dominikbraun/graph"
)

const ROOT_NODE = "\000\001\002\000"

func SuggestWord(previousWords []string) {
	var vertex = 0
	for _, val := range previousWords {
		var hashedVal = hash(val)
		_, err := g.Edge(vertex, hashedVal)
		for err != nil {
			hashedVal += 1
			_, err = g.Edge(vertex, hashedVal)
		}
		vertex = hashedVal
	}
	var adjMap, err = g.AdjacencyMap()
	if err != nil {
		panic(err)
	}
	var nextWords = adjMap[vertex]
	// weight is 0+, probability of word being selected

	var selected graph.Edge[int]
	for _, edge := range nextWords {

	}
}

func sum(array [64]byte) int {
	result := 0
	for _, v := range array {
		result += int(v)
	}
	return result
}

var power = big.NewInt(2)

var g graph.Graph[int, string]

func hash(value string) int {
	if value == ROOT_NODE {
		return 0
	}
	var hashNum = sum(sha512.Sum512([]byte(value)))
	for _, err := g.Vertex(hashNum); err != graph.ErrVertexNotFound; {
		hashNum += 1
	}
	return int(hashNum)
}

func hashDeterministic(value string) int {
	if value == ROOT_NODE {
		return 0
	}
	var hashNum = sum(sha512.Sum512([]byte(value)))
	return int(hashNum)
}

func main() {
	power.Exp(big.NewInt(2), big.NewInt(16), nil)
	g = graph.New(hash, graph.Directed(), graph.Rooted(), graph.Weighted())

	g.AddVertex(ROOT_NODE, graph.VertexWeight(0), graph.VertexAttribute("root", "yes"), graph.VertexAttribute("word", ""))
	g.AddVertex("test", graph.VertexWeight(0), graph.VertexAttribute("root", "no"), graph.VertexAttribute("word", "test"))

	g.AddEdge(0, hashDeterministic("test"), graph.EdgeWeight(100))

	SuggestWord([]string{ROOT_NODE})
}
