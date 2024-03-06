package main

import (
	"fmt"
	"strings"
	//"time"
	"hash/fnv"
	"golang.org/x/exp/slices"
	//"strconv"
)

// HYPERPARAMETERS
const _startNonTerminal = "S"

const (
	logging = false
)

//END HYPERPARAMETERS

var recordEdge = true
var deriToEdge = map[uint64][]Edge{}
var deriToDeri = map[[2]uint64]bool{}

type reach struct {
	worklistIdx          int
	worklist             []derivation
	nameToDerivations    nameToDerivations
	nameToDerivationsMap map[string]map[uint64]bool
	derivationVertexMap  map[derivationVertex][]*derivation
	vertexSCC            map[Vertex]int
	reachabilitySCC      map[[2]int]bool
	taintReachable		 map[[2]Vertex]bool

}

type path struct {
	start    Vertex
	end      Vertex
}

type derivation struct {
	name       string
	segments   segments
}

type derivationVertex struct {
	name string
	dimension int
	start bool
	vertex Vertex
}


type segments []path

type nameToDerivations map[string][]*derivation

func (p path) Equals(p2 path) bool {
	return p.start == p2.start && p.end == p2.end
}

func (p path) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%v-%v", p.start, p.end)))
	return h.Sum64()
}

func (s segments) Equals(s2 segments) bool {
	if len(s) != len(s2) {
		return false
	}
	for i, el := range s {
		if !el.Equals(s2[i]) {
			return false
		}
	}
	return true
}

func (s segments) Hash() uint64 {
	hash := fnv.New64a()
	for _, path := range s {
		pathHash := path.Hash()
		hash.Write([]byte(fmt.Sprintf("%d-", pathHash)))
	}
	return hash.Sum64()
}

func (d derivation) Equals(d2 derivation) bool {
	return d.name == d2.name && d.segments.Equals(d2.segments);
}

func (d derivation) Hash() uint64 {
	h := fnv.New64a()
	segmentHash := d.segments.Hash()
	h.Write([]byte(fmt.Sprintf("%v-%d", d.name, segmentHash)))
	return h.Sum64()
}

func logg(a ...any) {
	if logging {
		fmt.Println(a...)
	}
}

func hasEdge(g *graph, from Vertex, to Vertex, label Label) bool {
	for _, edge := range g.GetEdges() {
		if edge.From == from && edge.To == to && edge.Label == label {
			return true
		}
	}
	return false
}

func AllPairsReachability(g *graph, m *MCFG, interleaved bool, refinedPairs [][]Vertex, charList ...[]int) ([]path, nameToDerivations) {
	logg("--- begin all pairs reachability ---")

	//startTime := time.Now()
	vSCC, rSCC := g.findSccs()

	tReachable := map[[2]Vertex]bool{}


	//fmt.Println("Preprocessing time:", time.Since(startTime))
	reachData := &reach{
		worklistIdx:          0,
		worklist:             []derivation{},
		nameToDerivations:    nameToDerivations{},
		nameToDerivationsMap: make(map[string]map[uint64]bool),
		derivationVertexMap:  make(map[derivationVertex][]*derivation),
		vertexSCC:            vSCC,
		reachabilitySCC:      rSCC,
		taintReachable:		  tReachable,
	}

	deriToEdge = map[uint64][]Edge{}
	deriToDeri = map[[2]uint64]bool{}

	//Initialization
	reachData.processBasicRules(g, m)

	//Main loop
	return reachData.allPairsReachabilityMainLoop(g, m)
}

func (reachData *reach) allPairsReachabilityMainLoop(g *graph, m *MCFG) ([]path, nameToDerivations) {

	//startTime := time.Now()

	//Initialize rules containing each body rule name
	prepRuleMap := make(map[string][]PrependRule)
	for _, rule := range m.PrependRules {
		prepRuleMap[rule.BodyName] = append(prepRuleMap[rule.BodyName],rule)
	}
	appeRuleMap := make(map[string][]AppendRule)
	for _, rule := range m.AppendRules {
		appeRuleMap[rule.BodyName] = append(appeRuleMap[rule.BodyName],rule)
	}
	inseRuleMap := make(map[string][]InsertRule)
	for _, rule := range m.InsertRules {
		inseRuleMap[rule.BodyName] = append(inseRuleMap[rule.BodyName],rule)
	}
	concRuleMap := make(map[string][]ConcatenateRule)
	for _, rule := range m.ConcatenateRules {
		for _, name := range rule.BodyNames {
			concRuleMap[name] = append(concRuleMap[name],rule)
		}
	}

	foundPairs := []path{}

	for len(reachData.worklist) != reachData.worklistIdx {

		worklistItem := reachData.popFromWorklist()

		if isStartNonTerminal(worklistItem.name) {
			foundPairs = append(foundPairs, worklistItem.segments[0])
		}

		{
	        rules, ok := prepRuleMap[worklistItem.name]
	        if ok {
	            reachData.processPrependRules(g, &worklistItem, &rules)
	        }
	    }
	    {
	        rules, ok := appeRuleMap[worklistItem.name]
	        if ok {
	            reachData.processAppendRules(g, &worklistItem, &rules)
	        }
	    }
	    {
	        rules, ok := inseRuleMap[worklistItem.name]
	        if ok {
	            reachData.processInsertRules(g, &worklistItem, &rules)
	        }
	    }
	    {
	        rules, ok := concRuleMap[worklistItem.name]
	        if ok {
	            reachData.processConcatenateRules(&worklistItem, &rules)
	        }
	    }
	}

	return foundPairs, reachData.nameToDerivations
}

func (r *reach) processBasicRules(g *graph, m *MCFG) {
	for _, edge := range g.GetEdges() {
		for _, basicRule := range m.BasicRules {
			if basicRule.Label != edge.Label {
				continue
			}
			derivation := derivation{
				name:       basicRule.HeadName,
				segments:   []path{makePath(edge.From, edge.To)},
			}
			if recordEdge {
				deriToEdge[derivation.Hash()] = append(deriToEdge[derivation.Hash()],edge)
			}
			r.addDerivation(&derivation)
		}
	}
}

func (r *reach) processPrependRules(g *graph, worklistItem *derivation, rules *[]PrependRule) {
	for _, prependRule := range *rules {

		segmentNeedingInEdge := worklistItem.segments[prependRule.PrependIdx]
		vertexNeedingInEdge := segmentNeedingInEdge.start
		candidateVertices := g.InEdges(vertexNeedingInEdge, prependRule.Label)

		for _, candidateVertex := range candidateVertices {
			segments := copyPathButReplace(worklistItem.segments, prependRule.PrependIdx,
				makePath(
					candidateVertex,
					worklistItem.segments[prependRule.PrependIdx].end,
				))
			derivation := derivation{
				name:       prependRule.HeadName,
				segments:   segments,
			}
			if recordEdge {
				derivationHash := derivation.Hash()
				edge := Edge{
					From:  candidateVertex,
					To:    vertexNeedingInEdge,
					Label: prependRule.Label,
				}
				deriToEdge[derivationHash] = append(deriToEdge[derivationHash],edge)
				deriToDeri[[2]uint64{derivationHash,(*worklistItem).Hash()}] = true
			}
			r.addDerivation(&derivation)
		}
	}
}

func (r *reach) processAppendRules(g *graph, worklistItem *derivation, rules *[]AppendRule) {
	for _, appendRule := range *rules {

		segmentNeedingOutEdge := worklistItem.segments[appendRule.AppendIdx]
		vertexNeedingOutEdge := segmentNeedingOutEdge.end
		candidateVertices := g.OutEdges(vertexNeedingOutEdge, appendRule.Label)

		for _, candidateVertex := range candidateVertices {
			segments := copyPathButReplace(worklistItem.segments, appendRule.AppendIdx,
				makePath(
					worklistItem.segments[appendRule.AppendIdx].start,
					candidateVertex,
				))
			derivation := derivation{
				name:       appendRule.HeadName,
				segments:   segments,
			}
			if recordEdge {
				derivationHash := derivation.Hash()
				edge := Edge{
					From:  vertexNeedingOutEdge,
					To:    candidateVertex,
					Label: appendRule.Label,
				}
				deriToEdge[derivationHash] = append(deriToEdge[derivationHash],edge)
				deriToDeri[[2]uint64{derivationHash,(*worklistItem).Hash()}] = true
			}
			r.addDerivation(&derivation)
		}
	}
}

func (r *reach) processInsertRules(g *graph, worklistItem *derivation, rules *[]InsertRule) {
	for _, insertRule := range *rules {
		for _, edge := range g.GetEdgesWithLabel(insertRule.Label) {
			derivation := derivation{
				name:       insertRule.HeadName,
				segments: copyPathAndInsert(worklistItem.segments, insertRule.InsertIdx,
					makePath(
						edge.From,
						edge.To,
					)),
			}
			if recordEdge {
				derivationHash := derivation.Hash()
				deriToEdge[derivationHash] = append(deriToEdge[derivationHash],edge)
				deriToDeri[[2]uint64{derivationHash,(*worklistItem).Hash()}] = true
			}
			r.addDerivation(&derivation)
		}
	}
}

func (r *reach) processConcatenateRules(worklistItem *derivation, rules *[]ConcatenateRule) {
	for _, concatenateRule := range *rules {
		for worklistItemIdxInRule, bodyName := range concatenateRule.BodyNames {
			if bodyName != worklistItem.name {
				continue
			}
			segments, derivations := r.findRHS(&concatenateRule, worklistItemIdxInRule, worklistItem)
			for i, ends := range segments {
				derivation := derivation{
					name:       concatenateRule.HeadName,
					segments: ends,
				}
				if recordEdge {
					derivationHash := derivation.Hash()
					for _, derived := range derivations[i] {
						deriToDeri[[2]uint64{derivationHash,(derived).Hash()}] = true
					}
				}
				r.addDerivation(&derivation)
			}
		}
	}
}

func (r *reach) findRHS(rule *ConcatenateRule, worklistIdx int, worklistItemSegments *derivation) ([]segments,[][]derivation) {

	if !firstCombinationCheck(rule,worklistItemSegments,worklistIdx) {
		return []segments{}, [][]derivation{}
	}

	combinations := []segments{}
	derivations := [][]derivation{}

	list := []*derivation{}
	for i := 0; i < len(rule.BodyNames); i++ {
		list = append(list, &derivation{})
	}
	list[worklistIdx]=worklistItemSegments

	var recursive func(int)
	recursive = func(idx int) {
		if idx == worklistIdx {
			idx++;
		}
		if idx == len(rule.BodyNames) {
			res := segments{}
			for _, t := range rule.TermConcatenation {
				firstSegment := list[t[0].FromBodyIdx].segments[t[0].FromIndexInBody]
				lastSegment := list[t[len(t)-1].FromBodyIdx].segments[t[len(t)-1].FromIndexInBody]
				res = append(res, makePath(firstSegment.start,lastSegment.end)) 
			}
			unList := []derivation{}
			for _, derivation := range list {
				unList = append(unList, *derivation)
			}
			derivations = append(derivations, unList)
			combinations = append(combinations, res)
			return
		}
		for _, el := range r.getFilteredDerivations(rule, worklistIdx, worklistItemSegments, idx) {
			list[idx] = el
			if !combinationCheck(rule,idx+1,list,worklistIdx) {
				continue 
			}
			recursive(idx+1)
		}
	}

	recursive(0)

	return combinations, derivations
}

func firstCombinationCheck(rule *ConcatenateRule, worklistItem *derivation, worklistIdx int) bool {
	for _, term := range rule.TermConcatenation {
		for i := 1; i < len(term); i++ {
			if term[i].FromBodyIdx == worklistIdx && term[i-1].FromBodyIdx == worklistIdx {
				if worklistItem.segments[term[i].FromIndexInBody].start !=
					worklistItem.segments[term[i-1].FromIndexInBody].end {
					return false
				}
			}
		}
	}
	return true
}

func combinationCheck(rule *ConcatenateRule, size int, list []*derivation, worklistIdx int) bool {
	for _, term := range rule.TermConcatenation {
		for i := 1; i < len(term); i++ {
			if term[i].FromBodyIdx != worklistIdx && term[i].FromBodyIdx >= size {
				i++
				continue
			}
			if term[i-1].FromBodyIdx != worklistIdx && term[i-1].FromBodyIdx >= size {
				continue
			}
			if list[term[i].FromBodyIdx].segments[term[i].FromIndexInBody].start !=
					list[term[i-1].FromBodyIdx].segments[term[i-1].FromIndexInBody].end {
				return false
			}
		}
	}

	return true
}

func (r *reach) getFilteredDerivations(rule *ConcatenateRule, worklistIdx int, worklistItemSegments *derivation, myIdx int) []*derivation {
	ans := []*derivation{}
	gotAns := false
	for _, term := range rule.TermConcatenation {
		for i := 1; i < len(term); i++ {
			if term[i-1].FromBodyIdx == worklistIdx && term[i].FromBodyIdx == myIdx {
				worklistVertex := worklistItemSegments.segments[term[i-1].FromIndexInBody].end
				startKey := derivationVertex{
					name: rule.BodyNames[myIdx],
					dimension: term[i].FromIndexInBody,
					start: true,
					vertex: worklistVertex,
				}
				if !gotAns || len(r.derivationVertexMap[startKey]) < len(ans) {
					gotAns = true
					ans = r.derivationVertexMap[startKey]
				}
			}

			if term[i-1].FromBodyIdx == myIdx && term[i].FromBodyIdx == worklistIdx {
				worklistVertex := worklistItemSegments.segments[term[i].FromIndexInBody].start
				endKey := derivationVertex{
					name: rule.BodyNames[myIdx],
					dimension: term[i-1].FromIndexInBody,
					start: false,
					vertex: worklistVertex,
				}
				if !gotAns || len(r.derivationVertexMap[endKey]) < len(ans) {
					gotAns = true
					ans = r.derivationVertexMap[endKey]
				}
			}

		}
	}
	if gotAns {
		return ans
	} else {
		return r.getDerivations(rule.BodyNames[myIdx])
	}	
}

func (r *reach) getDerivations(name string) []*derivation {
	res, ok := r.nameToDerivations[name]
	if !ok {
		return []*derivation{}
	}
	return res
}

func (r ConcatenateRule) bodyContains(ruleName string) bool {
	return slices.Contains(r.BodyNames, ruleName)
}

func copyPathButReplace(toCopy []path, idxToReplace int, replacement path) []path {
	segments := []path{}
	for i, segment := range toCopy {
		if i == idxToReplace {
			segments = append(segments, replacement)
			continue
		}
		segments = append(segments, segment)
	}
	return segments
}

func copyPathAndInsert(toCopy []path, idxToInsert int, insert path) []path {
	segments := []path{}
	for i, segment := range toCopy {
		if i == idxToInsert {
			segments = append(segments, insert)
		}
		segments = append(segments, segment)
	}
	if idxToInsert == len(toCopy) {
		segments = append(segments, insert)
	}
	return segments
}

func (r *reach) popFromWorklist() derivation {
	r.worklistIdx++
	return r.worklist[r.worklistIdx-1]
}
//If v reaches u
func (r *reach) reaches(name string, v Vertex, u Vertex) bool{
	return r.reachabilitySCC[[2]int{r.vertexSCC[v], r.vertexSCC[u]}]
}

func (r *reach) validReachability(toAdd *derivation)  bool {
	for i, _ := range toAdd.segments {
		if i > 0 && !r.reaches(toAdd.name,toAdd.segments[i-1].end,toAdd.segments[i].start) {
			return false
		}
	}
	
	return true
}

func (r *reach) addDerivation(toAdd *derivation) {
	if !r.validReachability(toAdd) {
		return 
	}
	if _, ok := r.nameToDerivations[toAdd.name]; !ok {
		r.nameToDerivations[toAdd.name] = []*derivation{}
		r.nameToDerivationsMap[toAdd.name] = make(map[uint64]bool)
	}

	if r.nameToDerivationsMap[toAdd.name][toAdd.Hash()] {
		return
	}

	for i, segment := range toAdd.segments {
		startKey := derivationVertex{
			name: toAdd.name,
			dimension: i,
			start: true,
			vertex: segment.start,
		}
		endKey := derivationVertex{
			name: toAdd.name,
			dimension: i,
			start: false,
			vertex: segment.end,
		}
		if _, ok := r.derivationVertexMap[startKey]; !ok {
			r.derivationVertexMap[startKey] = []*derivation{}
		}
		if _, ok := r.derivationVertexMap[endKey]; !ok {
			r.derivationVertexMap[endKey] = []*derivation{}
		}
		r.derivationVertexMap[startKey] = append(r.derivationVertexMap[startKey], toAdd)
		r.derivationVertexMap[endKey] = append(r.derivationVertexMap[endKey], toAdd)
	}

	r.worklist = append(r.worklist, *toAdd)

	r.nameToDerivations[toAdd.name] = append(r.nameToDerivations[toAdd.name], toAdd)
	r.nameToDerivationsMap[toAdd.name][toAdd.Hash()] = true
}

func (p path) sameEnds(p2 path) bool {
	return p.start == p2.start && p.end == p2.end
}

func isStartNonTerminal(name string) bool {
	return name == _startNonTerminal
}

func (d derivation) String() string {
	segmentStringList := []string{}
	for _, el := range d.segments {
		segmentStringList = append(segmentStringList, el.String())
	}
	return fmt.Sprintf("%s(%s)", d.name, strings.Join(segmentStringList, ", "))
}

func makePath(start Vertex, end Vertex) path {
	return path{start: start, end: end}
}

func (p path) String() string {
	return fmt.Sprintf("[%d %d]", p.start, p.end)
}
