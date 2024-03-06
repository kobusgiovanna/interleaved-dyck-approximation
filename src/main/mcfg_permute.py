# Jaco van de Pol, Aarhus University, January 2022
# 
# Idea: 
#   S(X Y Z) :- A(Y,Z,X)
#   A(X,Y,Z) :- B(Z,Y,X)                B(1,2,3) => A(3,2,1) => S(1 3 2)
#   B(1,2,3).
# =>
#   S012(X Y Z) :- A201(X,Y,Z)
#   A201(Z,X,Y) :- B(Z,Y,X)             B(1,2,3) => A(1,3,2) => S(1 3 2)
#   B(1,2,3).
# =>
#   S012(X Y Z) :- A201(X,Y,Z)
#   A201(Z,X,Y) :- B021(Z,X,Y)          B(1,3,2) => A(1,3,2) => S(1 3 2)
#   B021(1,3,2).

import sys
from mcfg_data import Rule, Head, Body
from mcfg_check import is_var
from mcfg_normal import new_nterm

def apply_perm(xs,perm):
    return [ xs[i] for i in perm ]

def get_permuted_rules(mcfg,task):
    (nterm,perm) = task
    result = []
    for rule in mcfg:
        if rule.head.nterm == nterm:
            args = [rule.head.args[i] for i in perm]
            result.append(Rule(Head(nterm,args),rule.body))
    return result

def get_permutation(lvars,rvars):
    perm = ()
    for x in lvars:
        try:
            perm = perm + (rvars.index(x),)
        except ValueError: # x is not in rvars
            continue
    return perm

def collect_vars_lhs(args):
    result = list()
    for w in args:
        result = result + [x for x in w if is_var(x)]
    return result

def permute_mcfg(mcfg):
    task = ('S',(0,))         # (Start symbol, identity permutation)
    work = [task]
    done = dict([(task,'S')]) # Start symbol in result is also 'S'
    dim  = dict([('S',1)])    # Start symbol has arity 1
    result = list()
    while work != []:
        task = work.pop(0)
        for rule in get_permuted_rules(mcfg, task):
            # compute new Head
            newA = done[task]
            new_head = Head(newA, rule.head.args)
            # compute new Body
            new_body = list()
            lvars = collect_vars_lhs(rule.head.args)
            for atom in rule.body:
                rvars = atom.args
                perm  = get_permutation(lvars, rvars)
                ntask = (atom.nterm, perm)
                if ntask not in done:
                    nterm = new_nterm(dim, len(rvars), atom.nterm)
                    if atom.nterm != nterm:
                        print(f'Map: {atom.nterm}->{nterm}', file=sys.stderr)
                    done[ntask] = nterm
                    work.append(ntask)
                else:
                    nterm = done[ntask]                
                args = [ rvars[i] for i in perm ]
                new_body.append(Body(nterm, args))
            result.append(Rule(new_head, new_body))
    return (result,dim)

if __name__ == "__main__":

    from glob import glob
    from mcfg_data  import print_mcfg
    from mcfg_check import check_mcfg, McfgError, McfgWarning
    from mcfg_parse import read_mcfg, ParseError, Reader


    def main():
        for name in glob('Test/perm*.mcfg') + glob('Test/rank*.mcfg'):
            print(f'\nTest case {name}:')
            try:
                reader = Reader(open(name))
                m = read_mcfg(reader)
                check_mcfg(m)
            except (ParseError, McfgError, IOError) as err:
                print(f'{name}: {err}')
                continue
            except (McfgWarning) as err:
                print(f'{name}: {err}')
            print_mcfg(m)
            print("\nMake non-deleting and non-permuting:")            
            (m,_) = permute_mcfg(m)
            try:
                check_mcfg(m)
            except (McfgError) as err:
                print("WARNING: transform to rank 2 is permuting!\n")
            print_mcfg(m)
            print()

    main()
