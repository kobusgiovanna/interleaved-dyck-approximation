package main

import (
	"fmt"
	"strings"
)

var _termLabels = []string{"X", "Y", "Z", "V", "W", "Q", "R", "S", "T", "U"}

// Normal form rules cf MCFL section 3

// HeadName(Label).
type BasicRule struct {
	HeadName string
	Label    Label
}

// HeadName(x_1, ..., Label x_PrependIdx, ..., x_Terms) :- BodyName(x_1, ..., x_Terms)
type PrependRule struct {
	HeadName   string
	BodyName   string
	Label      Label
	PrependIdx int
	Terms      int
}

// HeadName(x_1, ..., x_AppendIdx Label, ..., x_Terms) :- BodyName(x_1, ..., x_Terms)
type AppendRule struct {
	HeadName  string
	BodyName  string
	Label     Label
	AppendIdx int
	Terms     int
}

// HeadName(x_1, ..., x_InsertIdx, Label, x_InsertIdx+1, ..., x_Terms+1) :- BodyName(x_1, ..., x_Terms)
type InsertRule struct {
	HeadName      string
	BodyName      string
	Label         Label
	InsertIdx     int
	OriginalTerms int
}

// HeadName(s_1, ..., s_Terms) :- BodyName_1(...), ... (where termConcatenation defines construction of s_i)
type ConcatenateRule struct {
	HeadName          string
	BodyNames         []string
	TermConcatenation []TermConcatenator
}

type Rule struct {
	Type            int
	BasicRule       BasicRule
	PrependRule     PrependRule
	AppendRule      AppendRule
	InsertRule      InsertRule
	ConcatenateRule ConcatenateRule
}

func (r Rule) String() string {
	if r.Type == 1 {
		return "BASIC" //r.BasicRule.String()
	}
	if r.Type == 2 {
		return r.PrependRule.String()
	}
	if r.Type == 3 {
		return r.AppendRule.String()
	}
	if r.Type == 4 {
		return "BASIC" //r.InsertRule.String()
	}
	if r.Type == 5 {
		return fmt.Sprintf("%s%s", "CONCAT: ", r.ConcatenateRule.String())
	}
	return "ERROR"
}

type TermConcatenator []TermIdentifier

type TermIdentifier struct {
	FromBodyIdx     int
	FromIndexInBody int
}

func (t TermIdentifier) String() string {
	return termName(t.FromIndexInBody, t.FromBodyIdx)
}

func (t TermConcatenator) String() string {
	terms := []string{}
	for _, el := range t {
		terms = append(terms, el.String())
	}
	return strings.Join(terms, " ")
}

func (r BasicRule) String() string {
	return fmt.Sprintf("%s(%s).", r.HeadName, r.Label)
}

func (r PrependRule) String() string {
	headContents := termNames(r.Terms, 0, r.HeadName)
	headContents[r.PrependIdx] = fmt.Sprintf("%s %s", r.Label, headContents[r.PrependIdx])

	bodyContents := termNames(r.Terms, 0, r.BodyName)

	return fmt.Sprintf("%s :- %s.", termToString(r.HeadName, headContents), termToString(r.BodyName, bodyContents))
}

func (r AppendRule) String() string {
	headContents := termNames(r.Terms, 0, r.HeadName)
	headContents[r.AppendIdx] = fmt.Sprintf("%s %s", headContents[r.AppendIdx], r.Label)

	bodyContents := termNames(r.Terms, 0, r.BodyName)

	return fmt.Sprintf("%s :- %s.", termToString(r.HeadName, headContents), termToString(r.BodyName, bodyContents))
}

func (r InsertRule) String() string {
	headContents := termNames(r.OriginalTerms, 0, r.HeadName)
	if r.InsertIdx == r.OriginalTerms {
		headContents[r.InsertIdx-1] = fmt.Sprintf("%s, %s", headContents[r.InsertIdx-1], r.Label)
	} else {
		headContents[r.InsertIdx] = fmt.Sprintf("%s, %s", r.Label, headContents[r.InsertIdx])
	}

	bodyContents := termNames(r.OriginalTerms, 0, r.BodyName)

	return fmt.Sprintf("%s :- %s.", termToString(r.HeadName, headContents), termToString(r.BodyName, bodyContents))
}

func (r ConcatenateRule) String() string {


	bodyIdxToLength := make([]int, len(r.BodyNames))
	for _, termConcatenator := range r.TermConcatenation {
		for _, termIdentifier := range termConcatenator {
			bodyIdxToLength[termIdentifier.FromBodyIdx] += 1
		}
	}
	bodyNaming := [][]string{}
	bodyContents := []string{}
	for bodyIdx, bodyName := range r.BodyNames {
		bodyList := termNames(bodyIdxToLength[bodyIdx], bodyIdx, bodyName)
		bodyContents = append(bodyContents, termToString(bodyName, bodyList))
		bodyNaming = append(bodyNaming, bodyList)
	}

	headContents := []string{}
	for _, el := range r.TermConcatenation {
		s := []string{}
		for _, termIdentifier := range el {
			s2 := bodyNaming[termIdentifier.FromBodyIdx][termIdentifier.FromIndexInBody]
			s = append(s, s2)
		}
		headContents = append(headContents, strings.Join(s, " "))
	}
	headContent := strings.Join(headContents, ", ")

	return fmt.Sprintf("%s(%s) :- %s.", r.HeadName, headContent, strings.Join(bodyContents, ", "))
}

func termName(idx int, labelIdx int) string {
	//TODO possible idxoutofbounds exception
	return fmt.Sprintf("%s%d", _termLabels[labelIdx], idx)
}

func termNames(termCount int, termLabelIdx int, headName string) []string {
	res := []string{}
	for i := 0; i < termCount; i++ {
		res = append(res, termName(i, termLabelIdx))
	}
	return res
}

func termToString(termName string, termBody []string) string {
	return fmt.Sprintf("%s(%s)", termName, strings.Join(termBody, ", "))
}
