package main

import (
	"strconv"
	"strings"
	"slices"
)

func mk_rule(params ...string) string {
	var res = params[0]
	for i := 1; i < len(params); i++ {
		if i == 1 {
			res += " :- "
		} else {
			res += ", "
		}
		res += params[i]
	}
	res += "."
	return res
}

func mk_stat(params ...string) string {
	var res = params[0] + "("
	for i := 1; i < len(params); i++ {
		if i > 1 {
			res += ", "
		}
		res += params[i]
	}
	res += ")"
	return res
}

func mk_base(funname string, k int, varname string) string {
	arr := make([]string, k+1)
	arr[0] = funname + strconv.Itoa(k)
	for i := 1; i <= k; i++ {
		arr[i] += varname + strconv.Itoa(i-1)
	}
	return mk_stat(arr...)
}

func writeLine(s string, res *string) {
	*res += s + "\n"
}

func mk_parenthesis(labelsP []int) string {
	var d = ""
	var res = &d
	for _, val := range labelsP {
        var si = strconv.Itoa(val)
        writeLine(mk_rule(mk_stat("Po"+si,"op--"+si)),res)
        writeLine(mk_rule("P(X0, cp--"+si+")","Po"+si+"(X0)"),res)
    }
    return d
}

func mk_dummy_parenthesis(name string, labelsP []int) string {
    var d = "" 
    var res = &d 
    for _, val := range labelsP {
        si := strconv.Itoa(val)
        writeLine(name + "(op--"+si+").", res)
        writeLine(name + "(cp--"+si+").", res)
    }
    return d
}

func mk_brackets(labelsB []int) string {
	var d = ""
	var res = &d
	for _, val := range labelsB  {
        var si = strconv.Itoa(val)
        writeLine(mk_rule(mk_stat("Bo"+si,"ob--"+si)),res)
        writeLine(mk_rule("B(X0, cb--"+si+")","Bo"+si+"(X0)"),res)
    }
    return d
}

func mk_dummy_brackets(name string, labelsB []int) string {
	var d = ""
	var res = &d
	for _, val := range labelsB  {
        var si = strconv.Itoa(val)
        writeLine(name + "(ob--"+si+").",res)
        writeLine(name + "(cb--"+si+").",res)
    }
    return d
}

func dyck_alpha_grammar(labelsP []int, labelsB []int) (MCFG, error) {
	var d = ""
	var res = &d
	writeLine(mk_parenthesis(labelsP),res)
	writeLine(mk_dummy_brackets("S",labelsB),res)
	writeLine("S(eps).",res)
	writeLine("S(normal).",res)
	writeLine("S(X0 Y0) :- S(X0), S(Y0).",res)
	if len(labelsP) > 0 {
		writeLine("S(Y0 X0 Y1) :- S(X0), P(Y0, Y1).",res)
	}
	return ParseNormalForm(strings.NewReader(*res))
}

func int_to_par(bitset int, k int) string {
	s := ""
	//0 is first
	for i := 0; i < k; i++ {
		if bitset>>i % 2 == 0 {
			s += "p"
		} else {
			s += "i"
		}
	}
	return s
}

func dyck_alpha_grammar_k_parity_se(labelsP []int, labelsB []int, k int) (MCFG, error) {

	var d = ""
	var res = &d

	slices.Sort(labelsB)

	labelGroup := make(map[int]int)

	for i, label := range labelsB {
		labelGroup[label] = i%k
	}

	writeLine(mk_parenthesis(labelsP),res)
	writeLine("Se(eps).",res)
	writeLine("Se(normal).",res)
	writeLine("Eps(eps).",res)
	//writeLine(mk_dummy_brackets("Si",labelsB),res)
	emptyNum := 0
	for _, val := range labelsB {
		var si = strconv.Itoa(val)
		emptyNum += 1<<labelGroup[val]
		p := int_to_par(emptyNum, k)
		writeLine("S"+p+"c(cb--"+si+").",res)
		writeLine("S"+p+"o(ob--"+si+").",res)
		emptyNum -= 1<<labelGroup[val]
	}
	if len(labelsP) > 0 {
		writeLine("Se(Y0 X0 Y1) :- Se(X0), P(Y0, Y1).",res)
	}
	for intP := 0;intP < 1<<k ; intP++ {
		p := int_to_par(intP, k)
		for _, c := range []string{"","c"} {
			for _, o := range []string{"", "o"} {
				if len(labelsP) > 0 {
					writeLine("S"+p+c+o+"(Y0 X0 Y1) :- S"+p+c+o+"(X0), P(Y0, Y1).",res)
				}
			}
		}
	}
	writeLine("Se(X0 Y0) :- Se(X0), Se(Y0).",res)
	for intP1 := 0;intP1 < 1<<k ; intP1++ {
		p1 := int_to_par(intP1, k)
		for _, c1 := range []string{"","c"} {
			for _, o1 := range []string{"", "o"} {
				writeLine("S"+p1+c1+o1+"(X0 Y0) :- S"+p1+c1+o1+"(X0), Se(Y0).",res)
				writeLine("S"+p1+c1+o1+"(X0 Y0) :- Se(X0), S"+p1+c1+o1+"(Y0).",res)
				for intP2 := 0;intP2 < 1<<k ; intP2++ {
					p2 := int_to_par(intP2, k)
					for _, c2 := range []string{"","c"} {
						for _, o2 := range []string{"", "o"} {
							intP3 := intP1 ^ intP2
							p3 := int_to_par(intP3, k)
							writeLine("S"+p3+c1+o2+"(X0 Y0) :- S"+p1+c1+o1+"(X0), S"+p2+c2+o2+"(Y0).",res)
						}
					}
				}
			}
		}
	}

	p := int_to_par(0, k)

	writeLine("S(X0 E) :- Se(X0), Eps(E).",res)
	writeLine("S(X0 E) :- S" + p + "(X0), Eps(E).",res)

	//fmt.Println(*res)

	return ParseNormalForm(strings.NewReader(*res))
}

func dyck_beta_grammar(labelsP []int, labelsB []int) (MCFG, error) {
    var d = ""
    var res = &d
    writeLine(mk_brackets(labelsB), res)
    writeLine(mk_dummy_parenthesis("S",labelsP), res)
    writeLine("S(eps).", res)
    writeLine("S(normal).", res)
    writeLine("S(X0 Y0) :- S(X0), S(Y0).", res)
    if len(labelsB) > 0 {
        writeLine("S(Y0 X0 Y1) :- S(X0), B(Y0, Y1).", res)
    }
    return ParseNormalForm(strings.NewReader(*res))
}

func dyck_beta_grammar_k_parity_se(labelsP []int, labelsB []int, k int) (MCFG, error) {

	var d = ""
	var res = &d

	slices.Sort(labelsP)

	labelGroup := make(map[int]int)

	for i, label := range labelsP {
		labelGroup[label] = i%k
	}

	writeLine(mk_brackets(labelsB),res)
	writeLine("Se(eps).",res)
	writeLine("Se(normal).",res)
	writeLine("Eps(eps).",res)
	//writeLine(mk_dummy_brackets("Si",labelsB),res)
	emptyNum := 0
	for _, val := range labelsP {
		var si = strconv.Itoa(val)
		emptyNum += 1<<labelGroup[val]
		p := int_to_par(emptyNum, k)
		writeLine("S"+p+"c(cp--"+si+").",res)
		writeLine("S"+p+"o(op--"+si+").",res)
		emptyNum -= 1<<labelGroup[val]
	}
	if len(labelsB) > 0 {
		writeLine("Se(Y0 X0 Y1) :- Se(X0), B(Y0, Y1).",res)
	}
	for intP := 0;intP < 1<<k ; intP++ {
		p := int_to_par(intP, k)
		for _, c := range []string{"","c"} {
			for _, o := range []string{"", "o"} {
				if len(labelsB) > 0 {
					writeLine("S"+p+c+o+"(Y0 X0 Y1) :- S"+p+c+o+"(X0), B(Y0, Y1).",res)
				}
			}
		}
	}
	writeLine("Se(X0 Y0) :- Se(X0), Se(Y0).",res)
	for intP1 := 0;intP1 < 1<<k ; intP1++ {
		p1 := int_to_par(intP1, k)
		for _, c1 := range []string{"","c"} {
			for _, o1 := range []string{"", "o"} {
				writeLine("S"+p1+c1+o1+"(X0 Y0) :- S"+p1+c1+o1+"(X0), Se(Y0).",res)
				writeLine("S"+p1+c1+o1+"(X0 Y0) :- Se(X0), S"+p1+c1+o1+"(Y0).",res)
				for intP2 := 0;intP2 < 1<<k ; intP2++ {
					p2 := int_to_par(intP2, k)
					for _, c2 := range []string{"","c"} {
						for _, o2 := range []string{"", "o"} {
							intP3 := intP1 ^ intP2
							p3 := int_to_par(intP3, k)
							writeLine("S"+p3+c1+o2+"(X0 Y0) :- S"+p1+c1+o1+"(X0), S"+p2+c2+o2+"(Y0).",res)
						}
					}
				}
			}
		}
	}

	p := int_to_par(0, k)

	writeLine("S(X0 E) :- Se(X0), Eps(E).",res)
	writeLine("S(X0 E) :- S" + p + "(X0), Eps(E).",res)

	//fmt.Println(*res)

	return ParseNormalForm(strings.NewReader(*res))
}

func interleaved_dyck(labelsP []int, labelsB []int) (MCFG, error) {
	//fmt.Println(K,len(labelsP),len(labelsB))
	var d = ""
	var res = &d

	writeLine(`Eps(eps).`,res)

	writeLine(mk_parenthesis(labelsP),res)
	writeLine(mk_brackets(labelsB),res)

	writeLine("S(eps).",res)
	writeLine("S(normal).",res)
	writeLine("S(X0 Y0) :- S(X0), S(Y0).",res)
	if len(labelsP) > 0 {
		writeLine("S(Y0 X0 Y1) :- S(X0), P(Y0, Y1).",res)
	}
	if len(labelsB) > 0 {
		writeLine("S(Y0 X0 Y1) :- S(X0), B(Y0, Y1).",res)
	}

	return ParseNormalForm(strings.NewReader(*res))

}

func bracket_grammar(labelsP []int, labelsB []int) (MCFG, error) {
	//fmt.Println(K,len(labelsP),len(labelsB))
	var d = ""
	var res = &d

	writeLine("Eps(eps).",res)
	writeLine("Sn(eps).",res)
	writeLine("Sn(normal).",res)

	for _, val := range labelsP {
		var si = strconv.Itoa(val)
		writeLine("Sn(cp--"+si+").",res)
		writeLine("Sn(op--"+si+").",res)
	}

	for _, val := range labelsB {
		var si = strconv.Itoa(val)
		writeLine("Sn(cb--"+si+").",res)
		writeLine("Sn(ob--"+si+").",res)
	}

	writeLine("Sn(X0 Y0) :- Sn(X0), Sn(Y0).",res)
	writeLine("S0(X0 cb--0) :- Sn(X0).",res)
	writeLine("S(ob--0 X0) :- S0(X0).",res)

	return ParseNormalForm(strings.NewReader(*res))

}

//implement choose grammar function

