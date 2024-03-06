# Jaco van de Pol, Aarhus University, January 2022

# Define "raw" MCFG data structure (can be semantically incorrect)
# Provides pretty-printing and computing dimension and rank

import sys


class Atom:
    def __init__(self, nterm, args):
        self.nterm = nterm
        self.args = args  # list of words or variables


class Head(Atom):
    # args is a list of words
    def __str__(self):
        def print_word(arg):
            return " ".join(x for x in arg)

        args = ", ".join(print_word(arg) for arg in self.args)
        return f"{self.nterm}({args})"


class Body(Atom):
    # args is a list of variables
    def __str__(self):
        args = ", ".join(arg for arg in self.args)
        return f"{self.nterm}({args})"


class Rule:
    def __init__(self, head, body):
        self.head = head  # Head
        self.body = body  # list[Body]

    def __str__(self):
        body = ", ".join(f"{rhs}" for rhs in self.body)
        if len(self.body) == 0:
            return f"{self.head}."
        else:
            return f"{self.head} :- {body}."


def dimension(mcfg):
    if len(mcfg) == 0:
        return 0
    return max(len(rule.head.args) for rule in mcfg)


def rank(mcfg):
    if len(mcfg) == 0:
        return 0
    return max(len(rule.body) for rule in mcfg)


def deg(rule):
    if len(rule.body) == 0:
        deg_body = 0
    else:
        deg_body = sum(len(atom.args) for atom in rule.body)
    return len(rule.head.args) + deg_body


def degree(mcfg):
    if len(mcfg) == 0:
        return 0
    return max(deg(rule) for rule in mcfg)


def print_mcfg(mcfg, file=None):
    print("\n".join(f"{rule}" for rule in mcfg), file=file)
    print(
        f"This is a {dimension(mcfg)}-MCFG({rank(mcfg)}) of degree {degree(mcfg)}",
        file=sys.stderr,
    )


def is_var(x):
    return "A" <= x[0] <= "Z"


def new_nterm(dim, arity, nterm="A"):
    if nterm not in dim:
        dim[nterm] = arity
        return nterm
    # get base name (without trailing numbers)
    while "0" <= nterm[len(nterm) - 1] <= "9":
        nterm = nterm[:-1]
    i = 0
    # search smallest fresh index
    while f"{nterm}{i}" in dim:
        i += 1
    dim[f"{nterm}{i}"] = arity
    return f"{nterm}{i}"


if __name__ == "__main__":
    from glob import glob

    from mcfg_parse import ParseError, Reader, enum_rule

    def main():
        for file in glob("Examples/*.mcfg") + glob("Test/*.mcfg"):
            print(f"File: {file}")
            try:
                reader = Reader(open(f"{file}"))
                for rule in enum_rule(reader):
                    print(rule)
            except (ParseError, IOError) as err:
                print(f"{file}: {err}")
            print()

    main()
