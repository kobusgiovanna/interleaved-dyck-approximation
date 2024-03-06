package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var directoryInput = "taint"
var directoryOutput = "taint-out"

var curr_grammar = "classic"

func main() {

	osInput := os.Args[1]

	fileStructure := strings.Split(osInput, "/")

	directoryInput = fileStructure[0]
	directoryOutput = directoryInput + "-out"
	current_benchmark := fileStructure[1]

	os.MkdirAll(directoryOutput, os.ModePerm)

	fileInfos, _ := ioutil.ReadDir(directoryInput)

	for _, fileInfo := range fileInfos {

		if fileInfo.Name() != current_benchmark {
			continue
		}

		//read graph
		fileName := directoryInput + "/" + fileInfo.Name()
		fmt.Println("Running:", fileName)
		g := ParseDotFile(directoryInput + "/" + fileInfo.Name())

		//removing vertices and edges from valueflow that are not reachable 
		//through a path that [s]
		if directoryInput == "valueflow" {
			g = g.removeValueflowUnreachable()
		}


		//regularization, for valueflow bracket condition is included in automaton
		regularizationPaths := getAutomatonReachability(g)
		outputFileName := directoryOutput + "/" + fileInfo.Name()
		outputFile, _ := os.Create(outputFileName)
		defer outputFile.Close()
		outputWord := "Regularization: " + strconv.Itoa(len(regularizationPaths))
		outputFile.Write([]byte(outputWord + "\n"))

		//intersection of grammars
		intersectionPaths := getIntersectionReachability(g)
		outputWord = "Intersection: " + strconv.Itoa(len(intersectionPaths))
		outputFile.Write([]byte(outputWord + "\n"))

		//remove useless edges (not in path among reachable pair)
		//from now on these edges cannot influence the answer
		g = g.removeNotPathPath(intersectionPaths)
		_, _, g = parseDyckComponent(g)


		//underapproximation through D(\Sigma_{\alpha}\cup\Sigma_{\beta})
		reachablePaths := getUnderApprox(g)
		outputWord = "Underapproximation: " + strconv.Itoa(len(reachablePaths))
		outputFile.Write([]byte(outputWord + "\n"))

		clearMaps()
		curr_grammar = "classic"
		classicMRPaths := getMROverApprox(g, reachablePaths)
		outputWord = "Mutual refinement: " + strconv.Itoa(len(classicMRPaths))
		outputFile.Write([]byte(outputWord + "\n"))

		//reduce graph further
		g = g.removeNotPathPath(classicMRPaths)
		_, _, g = parseDyckComponent(g)

		clearMaps()
		curr_grammar = "augmented"
		augmentedMRPaths := getMROverApprox(g, reachablePaths)
		outputWord = "Stronger Grammar: " + strconv.Itoa(len(augmentedMRPaths))
		outputFile.Write([]byte(outputWord + "\n"))

		//reduce graph further
		g = g.removeNotPathPath(augmentedMRPaths)
		_, _, g = parseDyckComponent(g)

		filteredOverPaths := getOnDemandMR(g, reachablePaths, augmentedMRPaths, fileInfo.Name())

		outputWord = "On-Demand: " + strconv.Itoa(len(filteredOverPaths))
		outputFile.Write([]byte(outputWord + "\n"))

	}
}

func getAutomatonReachability(g *graph) []path {
	alphaPaths := []path{}
	gComps := g.splitComponents()
	for _, gComp := range gComps {
		if len(gComp.edgeList) == len(gComp.vertices) {
			continue
		}
		parList, braList, comp := parseDyckComponentNaive(gComp)
		comp = comp.multiplyByAutomaton(braList)
		alphaGrammar, _ := dyck_alpha_grammar(parList, braList)
		recordEdge = false
		compPaths, _ := AllPairsReachability(comp, &alphaGrammar, false, [][]Vertex{}, parList, braList)
		recordEdge = true
		parsedCompPaths := filterAutomatonPaths(compPaths, braList)
		alphaPaths = append(alphaPaths,parsedCompPaths...)
	}
	return alphaPaths
}

func getIntersectionReachability(g *graph) []path {

	alphaPaths := []path{}
	betaPaths := make(map[path]bool)
	bracketPaths := make(map[path]bool)
	recordEdge = false

	gComps := g.splitComponents()
	for _, gComp := range gComps {
		//empty graph (ignoring trivial paths)
		if len(gComp.edgeList) == len(gComp.vertices) {
			continue
		}
		//find paths that respect alphaGrammar
		parList, braList, comp := parseDyckComponentNaive(gComp)
		alphaGrammar, _ := dyck_alpha_grammar(parList, braList)
		alphaPathsComp, _ := AllPairsReachability(comp, &alphaGrammar, false, [][]Vertex{}, parList, braList)
		alphaPaths = append(alphaPaths,alphaPathsComp...)
		
		//reduce the graph with information from alpha paths
		comp = comp.removeNotPathPath(alphaPathsComp)
		parList, braList, comp = parseDyckComponentNaive(comp)

		//find paths that respect betaGrammar
		betaGrammar, _ := dyck_beta_grammar(parList, braList)
		betaPathsComp, _ := AllPairsReachability(comp, &betaGrammar, false, [][]Vertex{}, parList, braList)
		for _, path := range betaPathsComp {
			betaPaths[path]=true
		}

		if directoryInput == "valueflow" {
			comp = comp.removeNotPathPath(betaPathsComp)
			parList, braList, comp = parseDyckComponentNaive(comp)

			bracketPathsComp := g.filterBracketPaths(betaPathsComp)
			for _, path := range bracketPathsComp {
				bracketPaths[path]=true
			}
		}

	}

	recordEdge = true

	overPaths := []path{}
	for _, alphaPath := range alphaPaths {
		if betaPaths[alphaPath] && alphaPath.start != alphaPath.end {
			if directoryInput != "valueflow" || bracketPaths[alphaPath] {
				overPaths = append(overPaths, alphaPath)
			}
		}
	}

	return overPaths
}

func getUnderApprox(g *graph) []path {

	_, _, gCopy := parseDyckComponentNaive(g)

	if directoryInput == "valueflow" {
		gCopy = gCopy.valueflowTransformation()
	}

	reachablePaths := []path{}
	gComps := gCopy.splitComponents()
	for _, gComp := range gComps {

		//empty graph (ignoring trivial paths)
		if len(gComp.edgeList) == len(gComp.vertices) {
			continue
		}

		//find paths that respect interleaved_dyck
		parList, braList, comp := parseDyckComponent(gComp)

		grammar, _ := interleaved_dyck(parList, braList)
		recordEdge = false
		compPaths, _ := AllPairsReachability(comp, &grammar, false, [][]Vertex{}, parList, braList)
		recordEdge = true

		reachablePaths = append(reachablePaths,compPaths...)

	}

	filteredReachable := []path{}
	for _, path := range reachablePaths {
		if path.start != path.end {
			filteredReachable = append(filteredReachable,path)
		}
	}

	if directoryInput == "valueflow" {
		return filterValueflowPaths(filteredReachable)
	}
	return filteredReachable
}

func getMROverApprox(g *graph, underApprox []path) []path {

	//merge mutually reachable vertices
	condensedGraph, parent := condensateFromUnderApprox(g, underApprox)

    MRCondensedOverPaths := mutualRefinement(condensedGraph, false, makePath(Vertex(0), Vertex(0)))

    afterTrans := make(map[Vertex][]Vertex)
    for chi, par := range parent {
    	afterTrans[par] = append(afterTrans[par],chi)
    }

    MROverPaths := []path{}
    for _, path := range MRCondensedOverPaths {
    	if path.start == path.end {
    		continue
    	}
    	for _, ini := range afterTrans[path.start] {
    		for _, fin := range afterTrans[path.end] {
    			if ini == fin {
    				continue
    			}
				MROverPaths = append(MROverPaths, makePath(ini,fin))
    		}
    	}
    }

    for v, _ := range condensedGraph.vertices {
    	for _, ini := range afterTrans[v] {
    		for _, fin := range afterTrans[v] {
    			if ini == fin {
    				continue
    			}
				MROverPaths = append(MROverPaths, makePath(ini,fin))
    		}
    	}
    }

	return MROverPaths

}

//make union find for guys that are reachable
//only run mutual refinement path if both of them are their own parents
//store the answer
//afterwards run again

func getOnDemandMR(g *graph, underApprox []path, overApprox []path, fileName string) []path{

	condensedGraph, parent := condensateFromUnderApprox(g, underApprox)

	underMap := make(map[path]bool)
	for _, pair := range underApprox {
		underMap[pair] = true
	}

	unknownPaths := []path{}
	for _, path := range overApprox {
		if !underMap[path] {
			unknownPaths = append(unknownPaths, path)
		}
	}

	//process paths that exist in condensed graph first
	//then the others can have their answer derived
	uRoot := make(map[Vertex][]path)
	uDerived := []path{}
	for _, uPath := range unknownPaths {
		if uPath.start == findPMR(uPath.start, &parent) && uPath.end == findPMR(uPath.end, &parent) {
			uRoot[uPath.start] = append(uRoot[uPath.start], uPath)
		} else {
			uDerived = append(uDerived, uPath)
		}
	}
	unknownPaths = []path{}
	for _, pathList := range uRoot {
		unknownPaths = append(unknownPaths, pathList...)
	}
	unknownPaths = append(unknownPaths, uDerived...)
	memory := make(map[path]bool)

	filteredOverPaths := underApprox

	//add the good paths
	for i, currPath := range unknownPaths {
		if i%100 == 0 {
			clearMaps()
		}

	    fv := findPMR(currPath.start, &parent)
		lv := findPMR(currPath.end, &parent)

		if currPath.start != fv || currPath.end != lv {
			if memory[makePath(fv,lv)] {
				filteredOverPaths = append(filteredOverPaths, currPath)
			}
			continue
		}

		if len(mutualRefinement(condensedGraph, true, makePath(fv,lv))) > 0 {
			memory[currPath] = true
			filteredOverPaths = append(filteredOverPaths, currPath)
		} 
	}

	return filteredOverPaths
}


func mutualRefinement(g *graph, onePath bool, myPath path) []path {

	if onePath && (!g.vertices[myPath.start] || !g.vertices[myPath.end]) {
		return []path{}
	}

	paths := []path{}
	components := g.splitComponents()

	for _, comp := range components {
		if onePath {
			if !comp.vertices[myPath.start] || !comp.vertices[myPath.end] {
				continue
			}
			comp = comp.removeNotPathPath([]path{myPath})
		}
		//if onepath then component must contain mypath
		parList, braList, parsedComp := parseDyckComponent(comp)
		oldEdgeNum := len(parsedComp.GetEdges())
		alphaPaths := getAlphaPaths(parsedComp, parList, braList)
		if onePath {
			found := false
			for _, path := range alphaPaths {
				if path == myPath {
					found = true
				}
			}
			if !found {
				return []path{}
			}
			alphaPaths = []path{myPath}
		}
		alphaEdges := usedEdges(&alphaPaths)
		parsedComp = getGraphFromEdgeMap(alphaEdges)
		parList, braList, parsedComp = parseDyckComponent(parsedComp)

		betaPaths := getBetaPaths(parsedComp, parList, braList)
		if onePath {
			found := false
			for _, path := range betaPaths {
				if path == myPath {
					found = true
				}
			}
			if !found {
				return []path{}
			}
			betaPaths = []path{myPath}
		}
		betaEdges := usedEdges(&betaPaths)
		parsedComp = getGraphFromEdgeMap(betaEdges)

		if directoryInput == "valueflow" {
			betaPaths = parsedComp.filterBracketPaths(betaPaths)
			if onePath && len(betaPaths) == 0 {
				return []path{}
			}
			parsedComp = parsedComp.removeValueflowUnreachable()
		}

		parList, braList, parsedComp = parseDyckComponent(parsedComp)
		currEdgeNum := len(parsedComp.GetEdges())

		if currEdgeNum == 0 || oldEdgeNum == currEdgeNum {
			//it has converged
			betaPathMap := make(map[path]bool)
			for _, betaPath := range betaPaths {
				betaPathMap[betaPath] = true
			}
			for _, alphaPath := range alphaPaths {
				if onePath && alphaPath != myPath {
					continue
				}
				if alphaPath.start != alphaPath.end && betaPathMap[alphaPath] {
					paths = append(paths, alphaPath)
				}
			}
		} else {
			newPaths := mutualRefinement(parsedComp, onePath, myPath)
			for _, path := range newPaths {
				if onePath && path != myPath {
					continue
				}
				paths = append(paths, path)
			}
			if onePath && len(newPaths)==0 {
				return []path{}
			}
		}

	}

	return paths
}


