package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseDotFile(filename string) *graph {
	return parseDotFile(filename, false)
}

// formats labels from i.e. "ob--XX" to "A".
// This is used for i.e. the antlr benchmark
func ParseDotFileAndFormatLabels(filename string) *graph {
	return parseDotFile(filename, true)
}

func parseDotFile(filename string, formatLabels bool) *graph {
	readFile, err := os.Open(filename)

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	g := MakeGraph()
	for fileScanner.Scan() {
		line := fileScanner.Text()
		parseDotLine(line, g, formatLabels)
	}

	readFile.Close()

	return g
}

func parseDotLine(line string, g *graph, formatLabels bool) {

	//need to parse lines of the form:
	//2128493581->1164059400[label="ob--43"]

	if !strings.Contains(line, "->") {
		return
	}
	//first split on ->
	//then split on [label="
	tokens1 := strings.Split(line, "->")

	startVertex, _ := strconv.Atoi(tokens1[0])
	tokens1 = strings.Split(tokens1[1], "[label=\"")
	endVertex, _ := strconv.Atoi(tokens1[0])

	var label Label
	if formatLabels {
		tokens1 = strings.Split(tokens1[1], "--")
		label = Label(parseLabel(tokens1[0]))
	} else {
		label = Label(strings.Split(tokens1[1], "\"")[0])
	}

	//check if edge already exists
	//for antlr benchmark, this cuts away about 10.000 edges (of about 70.000)
	for _, vtx := range g.outEdges[Vertex(startVertex)][label] {
		if vtx == Vertex(endVertex) {
			return
		}
	}

	g.AddEdge(Vertex(startVertex), Vertex(endVertex), label)
}

func parseLabel(label string) string {
	if label == "ob" {
		return "a"
	}
	if label == "op" {
		return "b"
	}
	if label == "cb" {
		return "A"
	}
	if label == "cp" {
		return "B"
	}
	return ""
}
