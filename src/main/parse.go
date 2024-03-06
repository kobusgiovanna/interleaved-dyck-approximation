package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"golang.org/x/exp/slices"
)

const _epsilonParseLabel = "eps"

func ParseNormalForm(reader io.Reader) (MCFG, error) {

	normalized, err := normalize(reader)
	if err != nil {
		return MCFG{}, errors.New("error message: " + normalized)
	}

	scanner := bufio.NewScanner(strings.NewReader(normalized))
	mcfg := makeEmptyMCFG()

	for scanner.Scan() {
		line := scanner.Text()
		parseLine(mcfg, line)
	}

	return *mcfg, nil
}

func normalize(reader io.Reader) (string, error) {
	cmd := exec.Command("python3", "mcfg.py", "-n")
	cmd.Stdin = reader

	outErr := bytes.NewBufferString("")
	outStd := bytes.NewBufferString("")
	cmd.Stderr = outErr
	cmd.Stdout = outStd
	err := cmd.Run()

	if err != nil {
		return outErr.String(), err
	}

	return outStd.String(), nil
}

func parseLine(mcfg *MCFG, line string) {
	if !strings.Contains(line, ".") {
		return
	}
	if strings.Contains(line, "-") && line[0:1] == "-" {
		return
	}
	trimmed := strings.Split(line, ".")[0]

	numSymbols := strings.Count(trimmed, "(")

	if numSymbols == 1 {
		parseBasicRule(mcfg, trimmed)
		return
	}

	tokens := strings.Split(trimmed, " :- ")
	lhs := tokens[0]
	rhs := tokens[1]

	if numSymbols == 2 {
		parseTwoSymbolRule(mcfg, lhs, rhs)
		return
	}
	parseConcatenateRule(mcfg, lhs, rhs)
}

func parseTwoSymbolRule(mcfg *MCFG, lhs string, rhs string) {
	_, lhsTokens := parsePredicate(lhs)
	_, rhsTokens := parsePredicate(rhs)

	if len(lhsTokens) != len(rhsTokens) {
		parseInsertRule(mcfg, lhs, rhs)
		return
	}
	parsePrependAppendRule(mcfg, lhs, rhs)
}

// "A(X0, 1) :- A1(X0)"
func parseInsertRule(mcfg *MCFG, lhs string, rhs string) {
	lhsName, lhsNestedTokens := parsePredicate(lhs)
	rhsName, rhsNestedTokens := parsePredicate(rhs)

	flattenedRhsTokens := flatten(rhsNestedTokens)

	label := ""
	idx := -1

	for i, tokens := range lhsNestedTokens {
		token := tokens[0]
		if slices.Contains(flattenedRhsTokens, token) {
			continue
		}
		idx = i
		label = token
	}

	if idx == -1 {
		parseConcatenateRule(mcfg, lhs, rhs)
		return
	}

	if label == _epsilonParseLabel {
		label = _epsilonLabel
	}

	mcfg.InsertRules = append(mcfg.InsertRules, InsertRule{
		HeadName:      lhsName,
		BodyName:      rhsName,
		Label:         Label(label),
		InsertIdx:     idx,
		OriginalTerms: len(rhsNestedTokens),
	})
}

func parsePrependAppendRule(mcfg *MCFG, lhs string, rhs string) {
	lhsName, lhsNestedTokens := parsePredicate(lhs)
	rhsName, rhsNestedTokens := parsePredicate(rhs)

	flattenedRhsTokens := flatten(rhsNestedTokens)

	for i, tokens := range lhsNestedTokens {
		if len(tokens) == 1 {
			continue
		}
		leftmostTokenIsNonterminal := slices.Contains(flattenedRhsTokens, tokens[0])
		isAppendRule := leftmostTokenIsNonterminal
		if isAppendRule {
			mcfg.AppendRules = append(mcfg.AppendRules, AppendRule{
				HeadName:  lhsName,
				BodyName:  rhsName,
				Label:     Label(tokens[1]),
				AppendIdx: i,
				Terms:     len(rhsNestedTokens),
			})
			continue
		}
		mcfg.PrependRules = append(mcfg.PrependRules, PrependRule{
			HeadName:   lhsName,
			BodyName:   rhsName,
			Label:      Label(tokens[0]),
			PrependIdx: i,
			Terms:      len(rhsNestedTokens),
		})
	}
}

// S(X1 Y1 X2 Y2) :- P(X1, X2), Q(Y1, Y2).
func parseConcatenateRule(mcfg *MCFG, lhs string, rhs string) {
	headName, lhsNestedTokens := parsePredicate(lhs)

	bodyNames := []string{}
	rhsNestedTokens := [][]string{}
	tokens := strings.Split(rhs, ")")
	for _, token := range tokens {
		if token == "" {
			continue
		}
		trimmed, _ := strings.CutPrefix(token, ", ")

		split := strings.Split(trimmed, "(")
		name := split[0]
		if name == "Eps" {
			continue
		}
		body := split[1]
		rhsNestedTokens = append(rhsNestedTokens, strings.Split(body, ", "))
		bodyNames = append(bodyNames, name)
	}

	termConcatenation := []TermConcatenator{}
	for _, tokens := range lhsNestedTokens {
		termConcatenator := TermConcatenator{}
		for _, token := range tokens {
			if token == "E" {
				continue
			}
			i, j := findNestedIndex(token, rhsNestedTokens)
			termConcatenator = append(termConcatenator, TermIdentifier{
				FromBodyIdx:     i,
				FromIndexInBody: j,
			})
		}
		termConcatenation = append(termConcatenation, termConcatenator)
	}

	mcfg.ConcatenateRules = append(mcfg.ConcatenateRules, ConcatenateRule{
		HeadName:          headName,
		BodyNames:         bodyNames,
		TermConcatenation: termConcatenation,
	})
}

func findNestedIndex(token string, nestedSlice [][]string) (int, int) {
	for i, slice := range nestedSlice {
		for j, el := range slice {
			if token != el {
				continue
			}
			return i, j
		}
	}
	return -1, -1
}

// "A(0)"
func parseBasicRule(mcfg *MCFG, line string) {
	name, body := parsePredicate(line)

	bodyLabel := body[0][0]

	if bodyLabel == _epsilonParseLabel {
		bodyLabel = _epsilonLabel
	}

	mcfg.BasicRules = append(mcfg.BasicRules, BasicRule{
		HeadName: name,
		Label:    Label(bodyLabel),
	})
}

func parsePredicate(pred string) (string, [][]string) {
	removeCloseParenthesis := strings.Replace(pred, ")", "", 1)
	tokens := strings.Split(removeCloseParenthesis, "(")
	if len(tokens) != 2 {
		fmt.Println("ERROR ", pred)
	}

	name := tokens[0]
	bodyString := tokens[1]

	return name, parseBody(bodyString)
}

func parseBody(body string) [][]string {
	tokens := strings.Split(body, ", ")
	res := make([][]string, len(tokens))
	for i, token := range tokens {
		res[i] = strings.Split(token, " ")
	}
	return res
}

func flatten(input [][]string) []string {
	flattened := []string{}
	for _, tokens := range input {
		flattened = append(flattened, tokens...)
	}
	return flattened
}

func makeEmptyMCFG() *MCFG {
	return &MCFG{
		BasicRules:       []BasicRule{},
		PrependRules:     []PrependRule{},
		AppendRules:      []AppendRule{},
		InsertRules:      []InsertRule{},
		ConcatenateRules: []ConcatenateRule{},
	}
}
