# Jaco van de Pol, Aarhus University, January 2022

from mcfg_data import Body, Head, Rule, new_nterm, is_var
from mcfg_normal import fresh_vars, find_rhs

'''Reduces the rank of all rules to maximum 2
   d-MCFG(k+2) is a subset of d(k-1)-MCFG(2)
'''
def rank2(mcfg,dim):
    result = list()
    for rule in mcfg:
        body  = rule.body[:]
        nterm = rule.head.nterm
        while len(body) > 2:
            b2 = body.pop()
            b1 = body.pop()
            X1 = b1.args
            X2 = b2.args
            m1 = len(X1)
            m2 = len(X2)
            A0 = new_nterm(dim, m1 + m2, nterm)
            Y  = [ [x] for x in X1 + X2]
            body.append(Body(A0, X1 + X2))
            newhead = Head(A0, Y)
            newbody = [b1,b2]
            result.append(Rule(newhead,newbody))
        result.append(Rule(rule.head,body))
    return result

def find_var(word):
    for i in range(0,len(word)):
        if is_var(word[i]):
            return i
    return len(word) # not found

'''Reduces rules of dimension 1 to rank 2, keeping dimension 1
   Similar to Chomsky Normal Form: 1-MCFG(k+2) is a subset of 1-MCFG(2)
'''
def chomsky(mcfg,dim):
    result = list()
    mcfg = mcfg[:]
    while len(mcfg) != 0:
        rule = mcfg.pop()
        A = rule.head.nterm
        args = rule.head.args
        body = rule.body
        if dim[A] > 1 or len(body) <= 2:
            result.append(rule)
            continue
        if  max(dim[atom.nterm] for atom in body) > 1:
            result.append(rule)
            continue
        # else
        w = args[0]
        i = find_var(w)
        if i == len(w):
            result.append(rule)
            continue
        # else  A(w1 X s2) -> ..., Aj(X), ...
        X = w[i]
        (j,_) = find_rhs(X, body)
        Aj = new_nterm(dim, 1, A)
        Z = fresh_vars(1,[X])[0]
        head1 = Head(A, [w[:i+1] + [Z]])
        body1 = [body[j], Body(Aj,[Z])]
        result.append(Rule(head1,body1))
        head2 = Head(Aj,[w[i+1:]])
        body2 = body[0:j] + body[j+1:]
        mcfg.append(Rule(head2, body2))

    return result

if __name__ == "__main__":

    from glob import glob
    from mcfg_data  import print_mcfg
    from mcfg_check import check_mcfg, McfgError, McfgWarning
    from mcfg_parse import read_mcfg, ParseError, Reader
    from mcfg_permute import permute_mcfg

    def main():
        for name in glob('Examples/*.mcfg') + glob('Test/norm*.mcfg') + glob('Test/rank*.mcfg'):
            print(f'\nTest case {name}:')
            try:
                reader = Reader(open(name))
                m = read_mcfg(reader)
                dim = check_mcfg(m)
            except (ParseError, McfgError, IOError) as err:
                print(f'{name}: {err}')
                continue
            print_mcfg(m)
            print("\nApply Chomsky Transform:")
            m = chomsky(m,dim)         
            print_mcfg(m)
            print("\nTransform to rank <= 2:")            
            m = rank2(m,dim)
            print_mcfg(m)
            try:
                check_mcfg(m)
            except (McfgError) as err:
                print(f'{name}: {err}')
                continue
            except (McfgWarning) as err:
                print("WARNING: transform to rank 2 is permuting!\n")
                (m,dim) = permute_mcfg(m)
                print_mcfg(m)
                check_mcfg(m)
            print()

    main()
