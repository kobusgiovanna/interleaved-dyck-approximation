package main

import (
	"fmt"
	"strings"
)

const _epsilonLabel = "" //empty string

type graph struct {
	outEdges     VertexMap
	inEdges      VertexMap
	edgeList     []Edge
	labelToEdges map[Label][]Edge
	vertices     map[Vertex]bool
}

type Edge struct {
	From  Vertex
	To    Vertex
	Label Label
}

type Vertex int

const ANY_VERTEX Vertex = -1

type Label string

type VertexList []Vertex

type LabelToVertexList map[Label]VertexList

type VertexMap map[Vertex]LabelToVertexList

func MakeGraph() *graph {
	return &graph{
		outEdges:     VertexMap{},
		inEdges:      VertexMap{},
		edgeList:     []Edge{},
		labelToEdges: map[Label][]Edge{},
		vertices:     map[Vertex]bool{},
	}
}

func (g *graph) AddEdge(from Vertex, to Vertex, label Label) {
	g.outEdges.addEdge(from, to, label)
	g.inEdges.addEdge(to, from, label)

	edge := Edge{
		From:  from,
		To:    to,
		Label: label,
	}

	g.edgeList = append(g.edgeList, edge)

	if _, ok := g.labelToEdges[label]; !ok {
		g.labelToEdges[label] = []Edge{}
	}
	g.labelToEdges[label] = append(g.labelToEdges[label], edge)

	if _, ok := g.vertices[from]; !ok {
		g.vertices[from] = true
		g.AddEdge(from, from, _epsilonLabel) //disabled for now
	}
	if _, ok := g.vertices[to]; !ok {
		g.vertices[to] = true
		g.AddEdge(to, to, _epsilonLabel) //disabled for now
	}
}

func (v VertexMap) addEdge(from Vertex, to Vertex, label Label) {
	if _, ok := v[from]; !ok {
		v[from] = LabelToVertexList{}
	}
	v[from].addEdge(to, label)
}

func (l LabelToVertexList) addEdge(to Vertex, label Label) {
	if _, ok := l[label]; !ok {
		l[label] = VertexList{}
	}
	l[label] = append(l[label], to)
}

// InEdges returns the incoming edges to 'to' with a given label
func (g *graph) InEdges(to Vertex, label Label) VertexList {
	if to == ANY_VERTEX {
		res := []Vertex{}
		for _, el := range g.GetEdgesWithLabel(label) {
			res = append(res, el.From)
		}
		return res
	}
	return g.inEdges.getEdges(to, label)
}

// InEdgesUnlabeled returns the incoming edges to 'to'
func (g *graph) InEdgesUnlabeled(to Vertex) VertexList {
    if to == ANY_VERTEX {
        res := []Vertex{}
        for _, el := range g.GetEdges() {
            res = append(res, el.From)
        }
        return res
    }
    return g.inEdges.getEdgesUnlabeled(to)
}

// OutEdges returns the outgoing edges from 'from' with a given label
func (g *graph) OutEdges(from Vertex, label Label) VertexList {
	if from == ANY_VERTEX {
		res := []Vertex{}
		for _, el := range g.GetEdgesWithLabel(label) {
			res = append(res, el.To)
		}
		return res
	}
	return g.outEdges.getEdges(from, label)
}

// OutEdges returns the outgoing edges from 'from'
func (g *graph) OutEdgesUnlabeled(from Vertex) VertexList {
	if from == ANY_VERTEX {
		res := []Vertex{}
		for _, el := range g.GetEdges() {
			res = append(res, el.To)
		}
		return res
	}
	return g.outEdges.getEdgesUnlabeled(from)
}

// getEdges returns the edges from a with a given label, as stored in the VertexMap
func (v VertexMap) getEdges(a Vertex, label Label) VertexList {
	lbl2vtxList, ok := v[a]
	if !ok {
		return VertexList{}
	}
	vtxList, ok := lbl2vtxList[label]
	if !ok {
		return VertexList{}
	}
	return vtxList
}

func (v VertexMap) getEdgesUnlabeled(a Vertex) VertexList {
	lbl2vtxList, ok := v[a]
	if !ok {
		return VertexList{}
	}
	vtxList := VertexList{}
	for _, labeledVtxList := range lbl2vtxList {
		vtxList = append(vtxList, labeledVtxList...)
	}
	return vtxList
}

func (g *graph) GetEdges() []Edge {
	return g.edgeList
}

func (g *graph) GetEdgesWithLabel(label Label) []Edge {
	return g.labelToEdges[label]
}

// Takes input in two modes:
func MakeLinearGraph(path string) *graph {
	if strings.Contains(path, " ") {
		return makeLinearGraphMultiCharAlphabet(path)
	}
	return makeLinearGraphSingleCharAlphabet(path)
}

func makeLinearGraphSingleCharAlphabet(path string) *graph {
	g := MakeGraph()
	for pos, char := range path {
		g.AddEdge(Vertex(pos), Vertex(pos+1), Label(char))
	}
	return g
}

func makeLinearGraphMultiCharAlphabet(path string) *graph {
	g := MakeGraph()
	for pos, char := range strings.Split(path, " ") {
		g.AddEdge(Vertex(pos), Vertex(pos+1), Label(char))
	}
	return g
}

func (g *graph) NumVertices() int {
	return len(g.vertices)
}

func (g *graph) ShortDescription() string {
	return fmt.Sprintf("%d vertices, %d edges", g.NumVertices(), len(g.edgeList))
}

func (g *graph) splitComponents() []*graph {

	currentComponent := 0
	vertexComponent := map[Vertex]int{}
	seen := map[Vertex]bool{}

	var dfs func(v Vertex)
	dfs = func(v Vertex) {
		if seen[v] {
			return
		}
		seen[v] = true
		vertexComponent[v] = currentComponent
		for _, w := range g.OutEdgesUnlabeled(v) { dfs(w) }
		for _, w := range g.InEdgesUnlabeled(v) { dfs(w) }
	}

	for v, _ := range g.vertices {
		if _, ok := seen[v]; !ok { 
			dfs(v)
			currentComponent++;
		}
	}

	components := make([]*graph, currentComponent)
	for i, _ := range components {
		components[i] = MakeGraph() 
	}

	for _, e := range g.edgeList {
		if len(e.Label) == 0 && e.From == e.To {
			continue
		}
		components[vertexComponent[e.From]].AddEdge(e.From,e.To,e.Label)
	}

	return components

} 

func (g *graph) findSccs() (map[Vertex]int, map[[2]int]bool) {
	index := 0
	vertexIndex := map[Vertex]int{}
	vertexLowlink := map[Vertex]int{}
	vertexOnStack := map[Vertex]bool{}

	sccIndex := 0
	vertexToSCC := map[Vertex]int{}

	S := []Vertex{}

	var connected func(v Vertex)
	connected = func(v Vertex) {
		vertexIndex[v] = index
		vertexLowlink[v] = index
		index++
		S = append(S,v)
		vertexOnStack[v] = true
		//fmt.Println("edges", int(v), len(g.OutEdgesUnlabeled(v)))
		//check outedgesunlabeled
		for _, w := range g.OutEdgesUnlabeled(v) {
			//fmt.Println("--",int(v),int(w))
			if _, ok := vertexIndex[w]; !ok {
				connected(w)
				if vertexLowlink[w] < vertexLowlink[v] {
					vertexLowlink[v] = vertexLowlink[w]
				}
			} else if vertexOnStack[w] {
				if vertexIndex[w]< vertexLowlink[v] {
					vertexLowlink[v] = vertexIndex[w]
				}
			}
		}
		if vertexLowlink[v] == vertexIndex[v] {
			//oldlen := len(S)
			for {
				w := S[len(S)-1]
				S = S[:len(S)-1]
				vertexOnStack[w] = false
				vertexToSCC[w] = sccIndex
				if w == v {
					break
				}
			}
			sccIndex++
		}
	}

	for v, _ := range g.vertices {
		if _, ok := vertexIndex[v]; !ok {
			connected(v)
		}
	} 

	reach := map[[2]int]bool{}
	outEdges := map[int][]int{}

	for i := 0; i < sccIndex; i++ {
		outEdges[i] = []int{}
	}

	for v, _ := range g.vertices {
		vcc := vertexToSCC[v]
		reach[[2]int{vcc,vcc}] = true
		for _, w := range g.OutEdgesUnlabeled(v) {
			wcc := vertexToSCC[w]
			if !reach[[2]int{vcc,wcc}] {
				reach[[2]int{vcc,wcc}] = true
				outEdges[vcc] = append(outEdges[vcc],wcc)
			}
		}
	}

	for i := 0; i < sccIndex; i++ {
		for _, j := range outEdges[i] {
			for k := 0; k < j; k++ {
				if reach[[2]int{i,j}] && reach[[2]int{j,k}] {
					reach[[2]int{i,k}] = true
				}
			}
		}
	}

	return vertexToSCC, reach
}

