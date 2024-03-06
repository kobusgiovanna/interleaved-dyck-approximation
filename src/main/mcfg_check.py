# Jaco van de Pol, Aarhus University, January 2022

# Check the correctness conditions for an MCFG (static semantics)

# check: each non-terminal has the same dimension in all rules
# check: all arguments in the body are unique variables
# check: all variables in the head occur at most once
# check: all variables in the head occur in the body
# check: each non-terminal that occurs in some body is defined in some head
# check: there exists a start symbol S, and it has dimension 1

# warning: all variables in the head occur in the same order as in the body
#        (non-deleting and non-ordering)

import sys

from mcfg_data import print_mcfg, is_var


class McfgError(Exception):
    pass


class McfgWarning(Exception):
    pass


""" Check if all definitions agree on the arity of the non-terminals
    This also updates the dimension dictionary "dim"
"""


def check_head(head, dim):
    A = head.nterm
    d = len(head.args)
    if A in dim and dim[A] != d:
        raise (McfgError(f"head {A} cannot have dimension {dim[A]} and {d}"))
    else:
        dim[A] = d


""" Check if all non-terminals in the rhs have been defined with the proper arity"""


def check_body(rule, dim):
    for rhs in rule.body:
        A = rhs.nterm
        d = len(rhs.args)
        if A not in dim:
            raise (McfgError(f"{A} in body has no defining head"))
        if dim[A] != d:
            raise (
                McfgError(f"{A} in body has dimension {d}, but the head used {dim[A]}")
            )


""" Check if the rhs consists of unique variables only"""


def check_vars_rhs(rhss):
    vars = set()
    for rhs in rhss:
        for arg in rhs.args:
            if not is_var(arg):
                raise (
                    McfgError(
                        f"body contains non-variable {arg} (all variables must be capitalized)"
                    )
                )
            elif arg in vars:
                raise (McfgError(f"body is non-linear"))
            else:
                vars.add(arg)
    return vars


""" Check if the left-hand side uses all variabels in the rhs precisely once."""


def check_vars_lhs(lhs, all_vars):
    seen_vars = set()
    for w in lhs.args:
        for x in w:
            if is_var(x):
                if not x in all_vars:
                    raise (McfgError(f"head contains free variable {x}"))
                elif x in seen_vars:
                    raise (McfgError(f"head is non-linear"))
                else:
                    seen_vars.add(x)
    if len(all_vars) != len(seen_vars):
        raise (McfgWarning(f"rule is deleting"))


""" Check if rule is permuting, assuming it is left-linear"""


def check_permuting(rule):
    for atom in rule.body:
        vars = atom.args[:]
        for w in rule.head.args:
            for x in w:
                if len(vars) > 0 and x == vars[0]:
                    vars.pop(0)
        if len(vars) != 0:
            raise (McfgWarning(f"rule is permuting"))


def check_var(rule):
    X = check_vars_rhs(rule.body)
    check_vars_lhs(rule.head, X)
    check_permuting(rule)


""" Check if the start rule is defined and has arity 1"""


def check_start(dim):
    if "S" not in dim:
        raise (McfgError("There should be a start symbol 'S'"))
    else:
        if dim["S"] != 1:
            raise (McfgError(f'Start symbol "S" has arity {dim["S"]}, not 1'))


def check_mcfg(mcfg):
    dim = dict()
    error = 0
    warning = 0
    for rule in mcfg:
        try:
            check_head(rule.head, dim)
        except McfgError as err:
            error += 1
            print(f"{rule} <-- {err}", file=sys.stderr)
    for rule in mcfg:
        try:
            check_body(rule, dim)
        except McfgError as err:
            error += 1
            print(f"{rule} <-- {err}", file=sys.stderr)
        try:
            check_var(rule)
        except McfgError as err:
            error += 1
            print(f"{rule} <-- {err}", file=sys.stderr)
        except McfgWarning as err:
            warning += 1
            print(f"Warning: {rule} <-- {err}", file=sys.stderr)
    try:
        check_start(dim)
    except McfgError as err:
        error += 1
        print(f"{err}", file=sys.stderr)
    if error == 0:
        if warning == 0:
            return dim
        else:
            raise (McfgWarning(f"Detected {warning} warnings"))
    else:
        raise (McfgError(f"Detected {error} semantic errors"))


def check_and_exit(mcfg):
    try:
        return check_mcfg(mcfg)
    except McfgError as err:
        print(f"ERROR: MCFG is semantically not valid", file=sys.stderr)
        sys.exit(2)
    except McfgWarning:
        print(f"WARNING: MCFG is deleting or permuting", file=sys.stderr)


if __name__ == "__main__":
    from glob import glob
    from mcfg_data import dimension, rank
    from mcfg_parse import read_mcfg, Reader, ParseError

    def check_file(reader, file):
        try:
            mcfg = read_mcfg(reader)
            print_mcfg(mcfg)
            check_mcfg(mcfg)
            print(f"{file} is a correct {dimension(mcfg)}-MCFG({rank(mcfg)})")
        except (ParseError, McfgError, McfgWarning) as err:
            print(f"{file}: {err}")

    def main():
        for file in (
            glob("Examples/*.mcfg") + glob("Test/check*.mcfg") + glob("Test/perm*.mcfg")
        ):
            print(f"\nFile: {file}")
            try:
                reader = Reader(open(file))
                check_file(reader, file)
            except IOError as err:
                print(f"{file}: {err}")

    main()
