package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	graph "gonum.org/v1/gonum/graph"
	simple "gonum.org/v1/gonum/graph/simple"
)

type MarkovGraph struct {
	*simple.WeightedDirectedGraph
	data  map[int64]string
	data2 map[int64]int64
}

func (mg *MarkovGraph) Init() {
	mg.data = make(map[int64]string, 5000)
	mg.data2 = make(map[int64]int64, 5000)

	var node = mg.NewNode()
	ROOT_NODE = mg.NewNode().ID()
	mg.AddNode(node)
}

func (mg *MarkovGraph) SetData(id int64, data string) {
	mg.data[id] = data
}
func (mg *MarkovGraph) GetData(id int64) (string, bool) {
	var data, ok = mg.data[id]
	return data, ok
}
func (mg *MarkovGraph) SetData2(id int64, data int64) {
	mg.data2[id] = data
}
func (mg *MarkovGraph) GetData2(id int64) (int64, bool) {
	var data, ok = mg.data2[id]
	return data, ok
}
func (mg *MarkovGraph) BFS(start int64, fun func(int64) bool) int64 {
	var visited = make(map[int64]bool)
	var queue []int64
	visited[start] = true
	queue = append(queue, start)
	for len(queue) != 0 {
		var s = queue[0]
		queue = queue[1:]
		if fun(s) {
			return s
		}
		var neighbors = mg.From(s)
		for neighbors.Next() {
			var node = neighbors.Node()
			if val, ok := visited[node.ID()]; ok && !val {
				visited[node.ID()] = true
				queue = append(queue, node.ID())
			}
		}
	}
	return int64(math.Inf(-1))
}

func NewMarkovGraph() *MarkovGraph {
	return &MarkovGraph{
		WeightedDirectedGraph: simple.NewWeightedDirectedGraph(0, 0),
		data:                  make(map[int64]string),
	}
}

var ROOT_NODE int64

var g MarkovGraph

func GetEdgesFromParent(parent int64) []graph.WeightedEdge {
	var edges = g.WeightedEdges()
	var out []graph.WeightedEdge
	for edges.Next() {
		var val = edges.WeightedEdge()
		if val.From().ID() == parent {
			out = append(out, val)
		}
	}
	return out
}

func GetHashFromWords(words []string) int64 {
	var vertex int64 = 0
	for _, val := range words {
		vertex = GetChildsHash(val, vertex)
	}
	return vertex
}

func SafeGetHashFromWords(words []string) (int64, bool) {
	var vertex int64 = 0
	for _, val := range words {
		if !HasChild(vertex, val) {
			return vertex, false
		}
		vertex = GetChildsHash(val, vertex)
	}
	return vertex, true
}

func HasChild(parent int64, word string) bool {
	var nodes = GetEdgesFromParent(parent)
	for _, node := range nodes {
		var data, ok = g.GetData(node.To().ID())
		if ok && data == word && node.From().ID() == parent {
			return true
		}
	}
	return false
}

func NumChildrenWithWord(parent int64, word string) int {
	var nodes = g.From(parent)
	var num = 0
	for nodes.Next() {
		var node = nodes.Node()
		var data, ok = g.GetData(node.ID())
		if ok && data == word {
			num++
		}
	}
	return num
}

func SuggestWord(previousWords []string, i int) string {
	fmt.Println("single word", i)
	if i >= 100 {
		return ""
	}
	var vertex = ""
	var successful = false
	var _, safe = SafeGetHashFromWords(previousWords)
	if !safe {
		fmt.Println("Not safe, removing last word")
		return SuggestWord(previousWords[:len(previousWords)-1], i+1)
	}
	g.BFS(GetHashFromWords(previousWords), func(v int64) bool {
		var nextWords = g.From(v)
		// weight is 0-100%, probability of word being selected

		var selected graph.Node = nil
		var iterations = 0
		if nextWords.Len() == 0 {
			selected = g.NewNode()
			var node, _ = SafeGetHashFromWords(previousWords)
			var iter = 1
			for g.From(node).Len() < 2 {
				node = GetParent(node)
				iter += 1
			}
			g.SetData(selected.ID(), SuggestWord(previousWords[:len(previousWords)-iter], i+1))
		}
		for selected == nil {
			iterations++
			nextWords.Reset()
			for nextWords.Next() {
				var node = nextWords.Node()
				var weight64, ok = g.Weight(v, node.ID())
				var weight = int(weight64)
				if !ok {
					weight = 100
				}
				if iterations >= 100 {
					selected = node
					break
				}
				if rand.Int()%100 < weight {
					selected = node
					break
				}
			}
		}
		text, ok := g.GetData(selected.ID())
		if !ok {
			return false
		}
		successful = true
		vertex = text
		return true
	})
	if !successful && len(previousWords) == 0 {
		return ""
	}
	if !successful {
		return SuggestWord(previousWords[:len(previousWords)-1], i+1)
	}
	return vertex
}

func SuggestWords(previousWords []string, numWords int) []string {
	var words []string
	for i := 0; i < numWords; i++ {
		var suggested = SuggestWord(previousWords, 0)
		var iterations = 0
		var notEnough = false
		for contains(words, suggested) {
			iterations++
			fmt.Println(iterations)
			if iterations > 100 {
				notEnough = true
				break
			}
			suggested = SuggestWord(previousWords, iterations)
		}
		if notEnough {
			break
		}
		words = append(words, suggested)
	}
	return words
}

func SuggestWordsAfterEachOther(previousWords []string, numWords int) []string {
	for i := 0; i < numWords; i++ {
		var output = SuggestWord(previousWords, 0)
		previousWords = append(previousWords, output)
	}
	return previousWords
}

func GetParent(node int64) int64 {
	var edges = g.Edges()
	for edges.Next() {
		var edge = edges.Edge()
		if edge.To().ID() == node {
			return edge.From().ID()
		}
	}
	return 0
}

func GetSiblings(node int64) []int64 {
	var nodes = g.From(GetParent(node))
	var siblings []int64
	for nodes.Next() {
		var node = nodes.Node()
		siblings = append(siblings, node.ID())
	}
	return siblings
}

func AddWord(previousWords []string, word string) {
	var vertex = GetHashFromWords(previousWords)
	var children = GetEdgesFromParent(vertex)
	var numChildren = len(children)
	if HasChild(vertex, word) {
		var data2, ok = g.GetData2(vertex)
		if !ok {
			g.SetData2(vertex, 0)
		}
		g.SetData2(vertex, data2+1)
	} else {
		var newNode = g.NewNode()
		g.AddNode(newNode)
		numChildren++
		g.SetData(newNode.ID(), word)
		g.SetData2(newNode.ID(), 1)
		if len(children) > 0 {
			g.SetWeightedEdge(g.NewWeightedEdge(children[0].From(), newNode, 100/float64(numChildren)))
		} else {
			g.SetWeightedEdge(g.NewWeightedEdge(g.Node(vertex), newNode, 100/float64(numChildren)))
		}
	}
	for _, item := range children {
		g.RemoveEdge(item.From().ID(), item.To().ID())
		data2, ok := g.GetData2(item.To().ID())
		if !ok {
			panic("No data2 found")
		}
		var edge = simple.WeightedEdge{F: item.From(), T: item.To(), W: float64(data2) * (100 / float64(numChildren))}
		g.SetWeightedEdge(edge)
	}
}

func Train(words []string) {
	fmt.Print("\n")
	for i := 0; i < len(words); i++ {
		if i >= len(words) {
			break
		}
		fmt.Printf("Training on word %d/%d: %s\n", i+1, len(words), words[i])
		AddWord(words[:i], words[i])
	}
}

func SaveGraph(f string) {
	var nodes = g.Nodes()
	var output = ""
	for nodes.Next() {
		var node = nodes.Node()
		var data, ok = g.GetData(node.ID())
		if !ok {
			if node.ID() == ROOT_NODE {
				data = ""
			} else {
				panic("Data not found")
			}
		}
		data2, ok := g.GetData2(node.ID())
		if !ok {
			panic("Data2 not found")
		}
		var id = node.ID()
		var edges = GetEdgesFromParent(id)
		var edgeOut = ""
		for _, edge := range edges {
			var weight, ok = g.Weight(id, edge.To().ID())
			if !ok {
				panic("Weight not found")
			}
			edgeOut += fmt.Sprintf("%d:%d|", edge.To().ID(), int(weight))
		}
		var out = fmt.Sprintf("%d:%s:%d;%s", id, data, data2, edgeOut)
		output += out + "\n"
	}
	var data = []byte(output)
	file, err := os.Create(f)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(data)
	if err != nil {
		panic(err)
	}
}

func LoadGraph(f string) {
	var file, err = os.Open(f)
	if err != nil {
		panic(err)
	}
	info, err := os.Stat(f)
	if err != nil {
		panic(err)
	}
	var data []byte = make([]byte, info.Size())
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
	file.Close()

	var str = string(data)
	var nodes = strings.Split(str, "\n")
	for _, node := range nodes {
		if node == "" {
			continue
		}
		var parts = strings.Split(node, ";")
		var nodeData = strings.Split(parts[0], ":")
		var id, interr = strconv.ParseInt(nodeData[0], 10, 64)
		if interr != nil {
			fmt.Println(node)
			panic(interr)
		}
		var data = nodeData[1]
		data2, interr := strconv.ParseInt(nodeData[2], 10, 64)
		if interr != nil {
			panic(interr)
		}
		var newNode, new = g.NodeWithID(id)
		if new {
			g.AddNode(newNode)
		}
		g.SetData(newNode.ID(), data)
		g.SetData2(newNode.ID(), int64(data2))
	}
	for _, node := range nodes {
		if node == "" {
			continue
		}
		var parts = strings.Split(node, ";")
		var lines = strings.Split(parts[1], "|")
		var nodeData = strings.Split(parts[0], ":")
		var id, interr = strconv.ParseInt(nodeData[0], 10, 64)
		if interr != nil {
			panic(interr)
		}
		for _, line := range lines {
			if line == "" {
				continue
			}
			var lineData = strings.Split(line, ":")
			lineID, interr := strconv.ParseInt(lineData[0], 10, 64)
			if interr != nil {
				fmt.Println(node)
				panic(interr)
			}
			weight, interr := strconv.ParseFloat(lineData[1], 64)
			if interr != nil {
				panic(interr)
			}
			g.SetWeightedEdge(g.NewWeightedEdge(g.Node(id), g.Node(lineID), weight))
		}
	}
}

func Trainish(sentence string) {
	var words = strings.Split(sentence, " ")
	for i := range words {
		if i >= len(words) {
			break
		}
		if words[i] == "" {
			if i == len(words)-1 {
				words = words[:i]
			} else {
				words = append(words[:i], words[i+1:]...)
			}
		}
	}
	fmt.Println(strings.Join(words, ", "))
	Train(words)
}

func CreateFragments(sentence string) []string {
	var fragments []string
	var words = strings.Split(sentence, "?")
	for i := 0; i < len(words); i++ {
		fragments = append(fragments, strings.Split(words[i], "!")...)
	}
	return fragments
}

func TrainFromFile(f string) {
	var file, err = os.Open(f)
	if err != nil {
		panic(err)
	}
	info, err := os.Stat(f)
	if err != nil {
		panic(err)
	}
	var data []byte = make([]byte, info.Size())
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
	file.Close()
	var str = string(data)
	str = strings.ReplaceAll(str, "\n", " ")
	var sentences = strings.Split(str, ".")
	for _, sentence := range sentences {
		fmt.Println("Training on sentence: " + sentence)
		sentence = strings.ReplaceAll(sentence, ".", "")
		sentence = strings.ReplaceAll(sentence, ",", "")
		sentence = strings.ReplaceAll(sentence, ":", "")
		sentence = strings.ReplaceAll(sentence, ";", "")
		sentence = strings.ReplaceAll(sentence, "(", "")
		sentence = strings.ReplaceAll(sentence, ")", "")
		sentence = strings.ReplaceAll(sentence, "\"", "")
		sentence = strings.ReplaceAll(sentence, "  ", " ")
		sentence = strings.ReplaceAll(sentence, "\t", " ")
		sentence = strings.ToLower(sentence)
		var fragments = CreateFragments(sentence)
		for _, fragment := range fragments {
			Trainish(fragment)
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// GetChildsHash returns the hash of the child node of the parent node with the given value. If the child node does not exist, it will return the next available hash.
//
// THIS WILL NOT WORK IF THERE ARE simplePLE CHILDREN WITH THE SAME VALUE!
func GetChildsHash(value string, parent int64) int64 {
	var nodes = GetEdgesFromParent(parent)
	var out int64 = int64(math.Inf(-1))
	var numChildrenFound = 0
	for _, node := range nodes {
		var data, ok = g.GetData(node.To().ID())
		if ok && data == value {
			numChildrenFound++
			out = node.To().ID()
		}
	}
	if numChildrenFound > 1 {
		panic("multiple children with the same value, value: \"" + value + "\" parent: " + strconv.FormatInt(parent, 10) + " number of identical children found: " + strconv.Itoa(numChildrenFound))
	}
	if out != int64(math.Inf(-1)) {
		return out
	}
	panic("No node found with value " + value + " and parent " + strconv.FormatInt(parent, 10))
}

func main() {
	g = *NewMarkovGraph()
	g.Init()

	TrainFromFile("./asimovstories.txt")
	SaveGraph("asimov.chain")

	fmt.Println(SuggestWordsAfterEachOther([]string{"the"}, 5))
}
