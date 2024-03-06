package main

import (
	"fmt"
	"strings"
)

type MCFG struct {
	BasicRules       []BasicRule
	PrependRules     []PrependRule
	AppendRules      []AppendRule
	InsertRules      []InsertRule
	ConcatenateRules []ConcatenateRule
}

func (m MCFG) Dimension() int {
	res := 1
	for _, rule := range m.PrependRules {
		if rule.Terms > res {
			res = rule.Terms
		}
	}
	for _, rule := range m.AppendRules {
		if rule.Terms > res {
			res = rule.Terms
		}
	}
	for _, rule := range m.InsertRules {
		if rule.OriginalTerms+1 > res {
			res = rule.OriginalTerms + 1
		}
	}
	for _, rule := range m.ConcatenateRules {
		if len(rule.TermConcatenation) > res {
			res = len(rule.TermConcatenation)
		}
	}
	return res
}

func (m MCFG) Rank() int {
	res := 0
	if len(m.PrependRules) != 0 {
		res = 1
	}
	if len(m.AppendRules) != 0 {
		res = 1
	}
	if len(m.InsertRules) != 0 {
		res = 1
	}
	for _, rule := range m.ConcatenateRules {
		if len(rule.BodyNames) > res {
			res = len(rule.BodyNames)
		}
	}
	return res
}

func (m MCFG) String() string {
	ruleStrings := []string{
		fmt.Sprintf("%d-dimensional MCFG of rank %d", m.Dimension(), m.Rank()),
	}
	for _, rule := range m.BasicRules {
		ruleStrings = append(ruleStrings, rule.String())
	}
	for _, rule := range m.PrependRules {
		ruleStrings = append(ruleStrings, rule.String())
	}
	for _, rule := range m.AppendRules {
		ruleStrings = append(ruleStrings, rule.String())
	}
	for _, rule := range m.InsertRules {
		ruleStrings = append(ruleStrings, rule.String())
	}
	for _, rule := range m.ConcatenateRules {
		ruleStrings = append(ruleStrings, rule.String())
	}
	return strings.Join(ruleStrings, "\n")
}
