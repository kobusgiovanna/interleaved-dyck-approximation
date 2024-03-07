package main

import (
	"strconv"
)

func otherLabel (s string) string {
	sp := []rune(s)
	if sp[0] == 'o' {
		sp[0] = 'c'
	} else {
		sp[0] = 'o'
	}
	return string(sp) 
}

func parseDyckComponent (g *graph) ([]int, []int, *graph) {
	seen := make(map[string]bool)
	parId := []int{}
	braId := []int{}
	for _, e := range g.GetEdges() {
		label := string(e.Label)
		if len(label)>0 && label!="normal" && !seen[label] && seen[otherLabel(label)] {
			idString := label[4:]
			currId, _ := strconv.Atoi(idString)
			if label[1] == 'p' {
				parId = append(parId, currId)
			} else {
				braId = append(braId, currId)
			}
		}
		seen[label] = true
	}
	parsedDyck := MakeGraph()
	for _, e := range g.GetEdges() {
		label := string(e.Label)
		if label=="normal" || (len(label)!=0 && seen[label] && seen[otherLabel(label)]) {
			parsedDyck.AddEdge(e.From,e.To,e.Label)
		}
	}
	return parId, braId, parsedDyck
}

func parseDyckComponentNaive (g *graph) ([]int, []int, *graph) {
	seen := make(map[string]bool)
	parId := []int{}
	braId := []int{}
	for _, e := range g.GetEdges() {
		label := string(e.Label)
		if label == "normal" || len(label)==0 {
			continue
		}
		idString := label[1:]
		//fmt.Println(idString)
		if len(label)>0 && !seen[idString] {
			currId, _ := strconv.Atoi(label[4:])
			if label[1] == 'p' {
				parId = append(parId, currId)
			} else {
				braId = append(braId, currId)
			}
		}
		seen[idString] = true
	}
	parsedDyck := MakeGraph()
	for _, e := range g.GetEdges() {
		label := string(e.Label)
		if len(label)!=0 {
			parsedDyck.AddEdge(e.From,e.To,e.Label)
		}
	}
	return parId, braId, parsedDyck
}

//remove normal edges for valueflow analysis

func findCondensateDyck(v Vertex, parent *map[Vertex]Vertex) Vertex {
	if v != (*parent)[v] {
		(*parent)[v] = findCondensateDyck((*parent)[v], parent)
	} 
	return (*parent)[v]
}

func joinCondensateDyck(v Vertex, u Vertex, parent *map[Vertex]Vertex, weight *map[Vertex]int) {
	v = findCondensateDyck(v, parent)
	u = findCondensateDyck(u, parent)
	if v == u {
		return
	}
	if (*weight)[v] < (*weight)[u] {
		v, u = u, v
	}
	(*weight)[v] += (*weight)[u]
	(*parent)[u] = v
}

func (g *graph) notCondensateDyck(underApprox [][]Vertex, c byte) (*graph, map[Vertex][]Vertex, map[Vertex]Vertex) {


	parent := make(map[Vertex]Vertex)

	for v, _ := range g.vertices {
		parent[v] = v
	}

	findToVertex := make(map[Vertex][]Vertex)
	for v, _ := range g.vertices {
		findToVertex[v] = append(findToVertex[v], v)
	}

	newGraph := g

	return newGraph, findToVertex, parent
}

func (g *graph) condensateDyck(underApprox [][]Vertex, c byte) (*graph, map[Vertex][]Vertex, map[Vertex]Vertex) {

	parent := make(map[Vertex]Vertex)
	weight := make(map[Vertex]int)

	for v, _ := range g.vertices {
		parent[v] = v
		weight[v] = 1
	}

	for _, e := range g.GetEdges() {
		label := string(e.Label)
		//join if both endpoints have degree one and its normal or ignorable

		if len(g.OutEdgesUnlabeled(e.From))<=2 && len(g.InEdgesUnlabeled(e.To))<=2 {
			if (len(label)>1 && label[1]==c) || label=="normal" {
				//fmt.Println("joining label ", label)
				joinCondensateDyck(e.From, e.To, &parent, &weight)
			}
		}
	}
	

	//there is something weird with the initialization of the undermap
	underMap := make(map[Vertex]map[Vertex]bool)
	for _, pair := range underApprox {
		if _, ok := parent[pair[0]]; !ok {
			continue
		}
		if _, ok := parent[pair[1]]; !ok {
			continue
		}
		fv := findCondensateDyck(pair[0], &parent)
		lv := findCondensateDyck(pair[1], &parent)
		if _, ok := underMap[fv]; !ok {
			underMap[fv] = make(map[Vertex]bool)
		}
		if _, ok := underMap[lv]; !ok {
			underMap[lv] = make(map[Vertex]bool)
		}
		underMap[fv][lv] = true
		if underMap[fv][lv] && underMap[lv][fv] {
			joinCondensateDyck(fv, lv, &parent, &weight)
		}
	}

	findToVertex := make(map[Vertex][]Vertex)
	for v, _ := range g.vertices {
		p := findCondensateDyck(v, &parent)
		findToVertex[p] = append(findToVertex[p], v)
		//fmt.Println("pai de ", int(v), "eh ", int(p))
	}

	newGraph := MakeGraph()
	exists := make(map[Edge]bool)

	for _, e := range g.GetEdges() {
		fV := findCondensateDyck(e.From, &parent)
		tV := findCondensateDyck(e.To, &parent)
		newEdge := Edge{
			From:  fV,
			To:    tV,
			Label: e.Label,
		}
		if (len(e.Label)>1 && e.Label[1]==c) || e.Label=="normal" {
			newEdge.Label = "normal"
		}
		if exists[newEdge] {
			continue
		}
		exists[newEdge] = true
		newGraph.AddEdge(newEdge.From, newEdge.To, newEdge.Label)
	}

	return newGraph, findToVertex, parent
}


//return paths that can have form [s]
func (g *graph) filterBracketPaths(paths []path) []path {
	comp, reach := g.findSccs()
	ans := []path{}
	for _, currPath := range paths {
		for _, outEdge := range g.OutEdges(currPath.start, "ob--0") {
			for _, inEdge := range g.InEdges(currPath.end, "cb--0") {
				if reach[[2]int{comp[outEdge],comp[inEdge]}] {
					ans = append(ans, currPath)
				}
			}
		}
	}
	return ans
}

func (g *graph) removeValueflowUnreachable() *graph {

	comp, reach := g.findSccs()
	source := make(map[int]bool)
	sink := make(map[int]bool)

	for _, edge := range g.edgeList {
		if len(edge.Label) < 2 {
			continue
		}
		if edge.Label[:2] == "ob" {
			source[comp[edge.From]] = true
		}
		if edge.Label[:2] == "cb" {
			sink[comp[edge.To]] = true
		}
	}

	keep := map[Vertex]bool{}
	for vertex, _ := range g.vertices {
		goodSource := false
		goodSink := false
		for src, _ := range source {
			if reach[[2]int{src,comp[vertex]}] {
				goodSource = true
				break
			}
		}
		if !goodSource {
			continue
		}
		for snk, _ := range sink {
			if reach[[2]int{comp[vertex],snk}] {
				goodSink = true
				break
			}
		}
		if !goodSink {
			continue
		}
		keep[vertex]=true
	}
	processed := MakeGraph()
	for _, edge := range g.edgeList {
		if len(edge.Label)>0 && keep[edge.From] && keep[edge.To]{
			processed.AddEdge(edge.From,edge.To,edge.Label)
			//fmt.Println(strconv.Itoa(int(edge.From))+"->" +strconv.Itoa(int(edge.To))+"[label=\""+string(edge.Label)+"\"]");
		}
	}
	return processed
}

func (g *graph) condensateValueflow() *graph {

	toAdd := []Edge{}
	deleted := make(map[Vertex]bool)

	outEdges := make(map[Vertex]int)
	inEdges := make(map[Vertex]int)
	outEdgesN := make(map[Vertex]int)
	inEdgesN := make(map[Vertex]int)
	outEdgesNLabel := make(map[Vertex]Vertex)
	inEdgesNLabel := make(map[Vertex]Vertex)

	for _, e := range g.GetEdges() {
		if len(e.Label)!=0 {
			outEdges[e.From]++;
			inEdges[e.To]++;
		}
		if e.Label == "normal" && e.From!=e.To {
			outEdgesN[e.From]++;
			inEdgesN[e.To]++;
			outEdgesNLabel[e.From]=e.To;
			inEdgesNLabel[e.To]=e.From;
		}
	}

	for v, _ := range g.vertices {
		if inEdges[v]==1 && outEdges[v]==1 && inEdgesN[v]==1 && outEdgesN[v]==1 {
			outEdgesNLabel[inEdgesNLabel[v]] = outEdgesNLabel[v]
			inEdgesNLabel[outEdgesNLabel[v]] = inEdgesNLabel[v]
			deleted[v] = true
			edge := Edge{
				From:  inEdgesNLabel[v],
				To:    outEdgesNLabel[v],
				Label: "normal",
			}
			toAdd = append(toAdd,edge)
		}
	}

	newGraph := MakeGraph()
	exists := make(map[Edge]bool)

	for _, e := range g.GetEdges() {
		fV := e.From
		tV := e.To
		if fV==tV && (len(e.Label)==0 || e.Label=="normal") {
			continue
		}
		if deleted[fV] || deleted[tV] {
			continue
		}
		newEdge := Edge{
			From:  fV,
			To:    tV,
			Label: e.Label,
		}
		if exists[newEdge] {
			continue
		}
		exists[newEdge] = true
		newGraph.AddEdge(newEdge.From, newEdge.To, newEdge.Label)
	}

	for _, e := range toAdd {
		fV := e.From
		tV := e.To
		if fV==tV && (len(e.Label)==0 || e.Label=="normal") {
			continue
		}
		if deleted[fV] || deleted[tV] {
			continue
		}
		newEdge := Edge{
			From:  fV,
			To:    tV,
			Label: e.Label,
		}
		if exists[newEdge] {
			continue
		}
		exists[newEdge] = true
		newGraph.AddEdge(newEdge.From, newEdge.To, newEdge.Label)
	}

	return newGraph
}

func (g *graph) getAllPaths() []path {
	paths := []path{}
	components := g.splitComponents()

	for _, comp := range components {

		scc, reach := comp.findSccs()

		for u, _ := range comp.vertices {
			for v, _ := range comp.vertices {
				if u != v && reach[[2]int{scc[u],scc[v]}] {
					makePath(u,v)
				}
			}
		}
	}

	return paths
}

func (g *graph) graphReaches(u Vertex, v Vertex, component *map[Vertex]int, reaches *map[[2]int]bool) bool {
	return (*reaches)[[2]int{(*component)[u],(*component)[v]}]
}

func (g *graph) removeNotPath(overApprox []path) (*graph){
	pathMatrix := [][]Vertex{}
	for _, path := range overApprox {
		pathMatrix = append(pathMatrix, []Vertex{path.start,path.end})
	}
	return g.removeNotPathMatrix(&pathMatrix)
}


func (g *graph) removeNotPathMatrix(overApprox *[][]Vertex) (*graph){
	if len(*overApprox) == len(g.vertices)*len(g.vertices) {
		return g
	}

	comp, reach := g.findSccs()
	overMap := map[[2]int]bool{}

	for _, pair := range (*overApprox) {
		overMap[[2]int{comp[pair[0]],comp[pair[1]]}] = true
		//fmt.Println("added ", comp[pair[0]], comp[pair[1]])
	}

	keep := map[[2]Vertex]bool{}
	for _, edge := range g.edgeList {
		keep[[2]Vertex{edge.From,edge.To}] = false
		//fmt.Println("try ", comp[edge.From], comp[edge.To])
		if overMap[[2]int{comp[edge.From],comp[edge.To]}] {
			keep[[2]Vertex{edge.From,edge.To}] = true
			continue
		}
		for pair,_ := range overMap {
			if reach[[2]int{pair[0],comp[edge.From]}] && reach[[2]int{comp[edge.To],pair[1]}] {
				keep[[2]Vertex{edge.From,edge.To}] = true
				break
			}
		}
	}
	processed := MakeGraph()
	for _, edge := range g.edgeList {
		if keep[[2]Vertex{edge.From,edge.To}] && (edge.From != edge.To || len(edge.Label) > 0) {
			processed.AddEdge(edge.From,edge.To,edge.Label)
			//fmt.Println(strconv.Itoa(int(edge.From))+"->" +strconv.Itoa(int(edge.To))+"[label=\""+string(edge.Label)+"\"]");
		}
	}
	return processed
}

func (g *graph) reachablePairs(refinedPairs *[][]Vertex) (map[[2]Vertex]bool) {
	comp, reach := g.findSccs()
	viable := map[[2]Vertex]bool{}

	overMap := map[[2]int]bool{}

	for _, pair := range (*refinedPairs) {
		//fmt.Println("pair ", pair[0], pair[1], comp[pair[0]], comp[pair[1]])
		overMap[[2]int{comp[pair[0]],comp[pair[1]]}] = true
	}

	for u, _ := range g.vertices {
		for v, _ := range g.vertices {
			if !reach[[2]int{comp[u],comp[v]}] {
				continue
			}
			if overMap[[2]int{comp[u],comp[v]}] {
				viable[[2]Vertex{u,v}] = true
				continue
			}
			for pair, _ := range overMap {
				//fmt.Println("overmap ", pair[0], pair[1], comp[u], comp[v])
				if reach[[2]int{pair[0],comp[u]}] && reach[[2]int{comp[v],pair[1]}] {
					viable[[2]Vertex{u,v}] = true
					break
				}
			}
		}
	}

	return viable
}

func (g *graph) valueflowTransformation() (*graph) {

	newGraph := MakeGraph()

	for _, e := range g.GetEdges() {
		if e.From == e.To && len(e.Label)==0 {
			continue
		}
		newGraph.AddEdge(Vertex(3*int(e.From))+1,Vertex(3*int(e.To)+1),e.Label)
		if len(e.Label)>1 && e.Label[:2] == "ob" {
			newGraph.AddEdge(Vertex(3*int(e.From)),Vertex(3*int(e.To)+1),e.Label)
		}
		if len(e.Label)>1 && e.Label[:2] == "cb" {
			newGraph.AddEdge(Vertex(3*int(e.From)+1),Vertex(3*int(e.To)+2),e.Label)
		}
	}

	return newGraph
}

func pathValueflowTransformation(overApprox *[]Vertex) ([]Vertex) {
	return []Vertex{Vertex(3*int((*overApprox)[0])),Vertex(3*int((*overApprox)[1])+2)}
}

func pathsValueflowTransformation(overApprox *[][]Vertex) ([][]Vertex) {
	newOverApprox := [][]Vertex{}
	for _, pair := range (*overApprox) {
		newOverApprox = append(newOverApprox, []Vertex{Vertex(3*int(pair[0])),Vertex(3*int(pair[1])+2)})
	}
	return newOverApprox
}

func filterValueflowPaths(paths []path) []path{
	ans := []path{}
	for _, vf := range paths {
		if vf.start%3 == 0 && vf.end%3 == 2 {
			ansPath := makePath(vf.start/3,vf.end/3)
			if ansPath.start != ansPath.end {
				ans = append(ans, ansPath)
			}
		}
	}
	return ans
}

func filterUsedEdges(sDerivations *[]path) (map[Edge]bool) {

	//fmt.Println("finding used edges")

	newDeriToEdge := map[uint64][]Edge{}
	newDeriToDeri := map[[2]uint64]bool{}

	deriToDeriList := map[uint64][]uint64{}
	for deriEdge , _ := range deriToDeri {
		deriToDeriList[deriEdge[0]]=append(deriToDeriList[deriEdge[0]], deriEdge[1])
	}

	seenDeri := map[uint64]bool{}
	seenEdge := map[Edge]bool{}

	var recursive func(curr uint64) 
	recursive = func(curr uint64) {
		if seenDeri[curr] {
			return
		}
		seenDeri[curr]= true
		//fmt.Println("printing one deri to edge")
		for _, edge := range deriToEdge[curr] {
			if len(edge.Label) == 0 {
				continue
			}
			newDeriToEdge[curr] = append(newDeriToEdge[curr], edge)
			seenEdge[edge] = true
		}
		//fmt.Println("done")
		for _, deri := range deriToDeriList[curr] {
			newDeriToDeri[[2]uint64{curr,deri}] = true
			//fmt.Println("usedEdges", curr, "called", deri)
			recursive(deri)
		}
	}

	for _, s := range (*sDerivations) {
		targetDerivation := derivation{
			name:       "S",
			segments:   []path{s},
		}
		recursive(targetDerivation.Hash())
	}
	//fmt.Println("filtered edges ", len(deriToEdge),len( newDeriToEdge) )
	deriToEdge = newDeriToEdge
	deriToDeri = newDeriToDeri
	//fmt.Println("filtered deri ", len(deriToDeri),len( newDeriToDeri) )

	//fmt.Println("found")

	return seenEdge

}

//not change parity for normal edges 
func usedEdges(sDerivations *[]path) (map[Edge]bool) {

	//fmt.Println("finding used edges")

	seenDeri := map[uint64]bool{}
	seenEdge := map[Edge]bool{}

	deriToDeriList := map[uint64][]uint64{}
	for deriEdge , _ := range deriToDeri {
		deriToDeriList[deriEdge[0]]=append(deriToDeriList[deriEdge[0]], deriEdge[1])
	}

	var recursive func(curr uint64) 
	recursive = func(curr uint64) {
		if seenDeri[curr] {
			return
		}
		seenDeri[curr]= true
		//fmt.Println("printing one deri to edge")
		for _, edge := range deriToEdge[curr] {
			if len(edge.Label) == 0 {
				continue
			}
			//fmt.Println(edge.From, edge.To, edge.Label)
			seenEdge[edge] = true
		}
		//fmt.Println("done")
		for _, deri := range deriToDeriList[curr] {
			//fmt.Println("usedEdges", curr, "called", deri)
			recursive(deri)
		}
	}

	for _, s := range (*sDerivations) {
		if s.start == s.end {
			continue
		}
		targetDerivation := derivation{
			name:       "S",
			segments:   []path{s},
		}
		recursive(targetDerivation.Hash())
	}

	return seenEdge

}


// automaton depends on benchmark
func (g *graph) multiplyByAutomaton(labelsB []int) (*graph) {

	newGraph := MakeGraph()

	//automata for one bracket only and takes care of s = [ s'] condition
	if directoryInput == "valueflow" {
		k := 6
		for _, e := range g.GetEdges() {
			if e.From == e.To && len(e.Label)==0 {
				continue
			}
			if len(e.Label)>1 && e.Label[1] == 'b' {
				if e.Label[:2] == "ob" {
					newGraph.AddEdge(Vertex(k*int(e.From)+0),Vertex(k*int(e.To)+3),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+1),Vertex(k*int(e.To)+3),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+2),Vertex(k*int(e.To)+3),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+3),Vertex(k*int(e.To)+4),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+4),Vertex(k*int(e.To)+4),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+5),Vertex(k*int(e.To)+4),"normal")
				}
				if e.Label[:2] == "cb" {
					newGraph.AddEdge(Vertex(k*int(e.From)+3),Vertex(k*int(e.To)+2),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+4),Vertex(k*int(e.To)+5),"normal")
					newGraph.AddEdge(Vertex(k*int(e.From)+5),Vertex(k*int(e.To)+5),"normal")
				}
			} else {
				newGraph.AddEdge(Vertex(k*int(e.From)+1),Vertex(k*int(e.To)+1),e.Label)
				newGraph.AddEdge(Vertex(k*int(e.From)+2),Vertex(k*int(e.To)+1),e.Label)
				newGraph.AddEdge(Vertex(k*int(e.From)+3),Vertex(k*int(e.To)+3),e.Label)
				newGraph.AddEdge(Vertex(k*int(e.From)+4),Vertex(k*int(e.To)+4),e.Label)
				newGraph.AddEdge(Vertex(k*int(e.From)+5),Vertex(k*int(e.To)+4),e.Label)
			}
		}
		return newGraph
	}

	k := len(labelsB)+2

	labelGraph := make(map[int]int)

	for i, label := range labelsB {
		labelGraph[label] = i+1
	}

	for _, e := range g.GetEdges() {
		if e.From == e.To && len(e.Label)==0 {
			continue
		}
		//fmt.Println("aaaa ",int(e.From),int(e.To),k,e.Label,e.Labe)
		if len(e.Label)>1 && e.Label[1] == 'b' {
			//fmt.Println("bbbbb ",int(e.From),int(e.To),k,e.Label)
			labelNum, _ := strconv.Atoi(string(e.Label)[4:])
			if e.Label[:2] == "ob" {
				//fmt.Println("open ",int(e.From),int(e.To),k,e.Label)
				for i := 1; i < k; i++ {
					//fmt.Println("adding edge here",int(e.From),int(e.To),k,e.Label)
					newGraph.AddEdge(Vertex(k*int(e.From)+i),Vertex(k*int(e.To)+k-1),"normal")
				}
				targetGraph := labelGraph[labelNum]
				newGraph.AddEdge(Vertex(k*int(e.From)),Vertex(k*int(e.To)+targetGraph),"normal")
			}
			if e.Label[:2] == "cb" {
				targetGraph := labelGraph[labelNum]
				newGraph.AddEdge(Vertex(k*int(e.From)+targetGraph),Vertex(k*int(e.To)),"normal")
				newGraph.AddEdge(Vertex(k*int(e.From)+k-1),Vertex(k*int(e.To)+k-1),"normal")
			}
		} else {
			for i := 0; i < k; i++ {
				newGraph.AddEdge(Vertex(k*int(e.From)+i),Vertex(k*int(e.To)+i),e.Label)
			}
		}
	}

	return newGraph


}

func filterAutomatonPaths(paths []path, labelsB []int) []path{
	s := 0
	k := len(labelsB)+2
	if directoryInput == "valueflow" {
		s = 2
		k = 6
	}
	ans := []path{}
	seen := make(map[path]bool)
	for _, vf := range paths {
		if int(vf.start)%k == 0 && (int(vf.end)%k == s || int(vf.end)%k == k-1){
			ansPath := makePath(Vertex(int(vf.start)/k),Vertex(int(vf.end)/k))
			if ansPath.start != ansPath.end && !seen[ansPath] {
				seen[ansPath] = true
				ans = append(ans, ansPath)
			}
		}
	}
	return ans
}









