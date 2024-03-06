# Jaco van de Pol, Aarhus University, January 2022

# Apply transformation to normal form [Pavlogiannis, van de Pol]

from sys import stderr

from mcfg_data import Body, Head, Rule, is_var, new_nterm


# STEP 1
def step1(mcfg,dim):
    result = list()
    for rule in mcfg:
        if len(rule.body)>0:
            result.append(rule)
        elif len(rule.head.args) <= 1:
            result.append(rule)
        else:
            A = rule.head.nterm
            (w1,*ws) = rule.head.args
            A1 = new_nterm(dim,1,A)
            print(f'Transform step 1: {A}->{A1}', file=stderr)
            result.append(Rule(Head(A1, [w1]), []))
            result.append(Rule(Head(A , [['X']] + ws), [Body(A1, ['X'])]))
    return result


# STEP 2

def step2(mcfg,dim):
    result = list()
    for rule in mcfg:
        if len(rule.body) > 0:
            result.append(rule)
        elif len(rule.head.args) > 1:
            result.append(rule)
        elif len(rule.head.args[0]) == 1:
            result.append(rule)
        else:
            A = rule.head.nterm
            (a,*w1) = rule.head.args[0]
            A1 = new_nterm(dim,1,A)
            print(f'Transform step 2: {A}->{A1}', file=stderr)
            result.append(Rule(Head(A1, [[a]]), []))
            result.append(Rule(Head(A , [['X'] + w1]), [Body(A1, ['X'])]))
    return result

# STEP 3

'''Return (i,j,k), the first word with a maximal non-variable part [j:k] '''
def find_terminal_segment(lhss):
    for i in range(0,len(lhss)):
        w = lhss[i]
        for j in range(0,len(w)):
            if not is_var(w[j]):
                for k in range(j+1,len(w)):
                    if is_var(w[k]):
                        return (i,j,k) 
                return (i,j,len(w)) # segment is terminal until the end
    return (len(lhss),0,0) # not found

def find_rhs(x,rhss):
    for l in range(0,len(rhss)):
        args = rhss[l].args
        for m in range(0,len(args)):
            if args[m] == x:
                return (l,m)
    print(f'Find: {x} in {rhss}')
    assert(False) # the variable x must occur in the right hand side

def step3(mcfg,dim):
    result = []
    mcfg   = mcfg[:]
    while len(mcfg) > 0:
        rule = mcfg.pop()
        head = rule.head
        body = rule.body
        if len(body) <= 1:
            result.append(rule)
            continue
        # else:
        A0x = head.args
        k0  = len(A0x)
        (i,j,k) = find_terminal_segment(A0x)
        if i == k0: # only variables
            result.append(rule)
            continue
        # else: i-th argument contains a maximal terminal word w=si[j:k]
        A0 = head.nterm
        si = A0x[i]
        w  = si[j:k]
        if j == 0 and k == len(si): # si = w
            A1 = new_nterm(dim,k0-1,A0)
            print(f'Transform step 3a: {A0}->{A1}', file=stderr)
            X1 = [ f'Z{n}' for n in range(0,k0)]
            X2 = [ [Zn] for Zn in X1 ]
            s1 = A0x[0:i]
            s2 = A0x[i+1:k0]
            head1 = Head(A0, X2[0:i] + [w] + X2[i+1:k0])
            body1 = Body(A1, X1[0:i] + X1[i+1:k0])
            result.append(Rule(head1, [body1]))
            head2 = Head(A1, s1+s2)
            mcfg.append(Rule(head2, body))
        elif j == 0: # si = w x s2
            x     = si[k]
            s2    = si[k+1:]
            (l,m) = find_rhs(x, body)
            atom  = body[l]
            Alx   = atom.args
            Al1   = new_nterm(dim,len(Alx),A0)
            print(f'Transform step 3b: {A0}->{Al1}', file=stderr)
            # add rule 1:
            newargs1    = Alx[:]
            newargs1[m] = w + [Alx[m]]
            head1       = Head(Al1, newargs1)
            result.append(Rule(head1, [atom]))
            # add rule 2:
            newargs2    = A0x[:]
            newargs2[i] = [x] + s2
            head2       = Head(A0,newargs2)
            body2       = body[:]
            body2[l]    = Body(Al1, body[l].args)
            mcfg.append(Rule(head2, body2))
        else: # j>0:  si = s1 x w s2
            x      = si[j-1]
            s1     = si[:j-1]
            s2     = si[k:]
            (el,m) = find_rhs(x, body)
            atom   = body[el]
            Alx    = atom.args
            Al1    = new_nterm(dim,len(Alx),A0)
            print(f'Transform step 3c: {A0}->{Al1}', file=stderr)
            # add rule 1:
            head1 = Head(Al1, Alx[:m] + [[Alx[m]] + w] + Alx[m+1:])
            result.append(Rule(head1,[atom]))
            # add rule 2:
            head2 = Head(A0, A0x[:i] + [s1+[x]+s2] + A0x[i+1:])
            body2 = body[:el] + [Body(Al1, Alx)] + body[el+1:]
            mcfg.append(Rule(head2, body2))
    return result

# STEP 4

def find_all_terminal(lhss):
    (i,j,k) = find_terminal_segment(lhss)
    l = len(lhss)
    if i < l and j == 0 and k == len(lhss[i]):
        return i
    return l # not found

def find_non_var(lhss):
    for i in range(0,len(lhss)):
        if not (len(lhss[i]) == 1 and is_var(lhss[i][0])):
            return i
    return len(lhss) # not found

def step4(mcfg,dim):
    result = []
    mcfg   = mcfg[:]
    while len(mcfg)>0:
        rule = mcfg.pop()
        head = rule.head
        body = rule.body
        if len(body) != 1:
            result.append(rule)
            continue
        # else:
        A0x = head.args
        k0  = len(A0x)
        i   = find_all_terminal(A0x)
        if i == k0: # no terminal argument found
            result.append(rule)
            continue
        # else:
        j = find_non_var(A0x[:i] + A0x[i+1:])
        if j == k0-1: # no other non-var argument found
            result.append(rule)
            continue
        # else: 
        A0 = head.nterm
        A1 = new_nterm(dim,k0-1,A0)
        print(f"Transform step 4: {A0}->{A1}", file=stderr)
        # add rule 1:
        X1 = [ f'Z{n}' for n in range(0,k0-1)]
        X2 = [ [Zn] for Zn in X1]
        head1 = Head(A0, X2[0:i] + A0x[i] + X2[i:k0-1])
        body1 = Body(A1, X1)
        result.append(Rule(head1, [body1]))
        # add rule 2:
        head2 = Head(A1, A0x[0:i] + A0x[i+1:k0])
        mcfg.append(Rule(head2, body))
    return result

# STEP 5

'''find an atom of length > min, with two variables'''
def find_two_vars(lhss,min):
    for i in range(0,len(lhss)):
        atom=lhss[i]
        if len(atom) > min:
            for j in range(0,len(atom)):
                if is_var(atom[j]):
                    for k in range(j+1,len(atom)):
                        if is_var(atom[k]):
                            return (i,j,k)
    return (len(lhss),0,0) # not found

def step5(mcfg,dim):
    result = []
    mcfg   = mcfg[:]
    while len(mcfg)>0:
        rule = mcfg.pop()
        head = rule.head
        body = rule.body
        if len(body)!=1:
            result.append(rule)
            continue
        # else:
        A0x = head.args
        k0  = len(A0x)
        (i,_,k) = find_two_vars(A0x,2)
        if i==k0: # no argument with two vars found
            result.append(rule)
            continue
        # else:
        si  = A0x[i]
        si1 = si[:k]
        si2 = si[k:]
        A0  = head.nterm
        A1  = new_nterm(dim,k0+1,A0)
        print(f"Transform step 5: {A0}->{A1}", file=stderr)
        # add rule 1
        X1 = [ f'Z{n}' for n in range(0,k0+1)]
        X2 = [ [Zn] for Zn in X1]
        head1 = Head(A0, X2[0:i] + [ [X1[i],X1[i+1]] ] + X2[i+2:k0+1])
        body1 = Body(A1, X1[0:k0+1])
        result.append(Rule(head1, [body1]))
        # add rule 2
        head2 = Head(A1,A0x[0:i] + [si1,si2] + A0x[i+1:k0])
        mcfg.append(Rule(head2, body))
    return result

# STEP 6

def find_too_long(lhss):
    for i in range(0,len(lhss)):
        atom = lhss[i]
        if len(atom) > 2:
            return i
        elif len(atom) == 2: # check if both terminal
            if not (is_var(atom[0]) or is_var(atom[1])):
                return i
    return len(lhss) # not found

def step6(mcfg,dim):
    result = []
    mcfg   = mcfg[:]
    while len(mcfg)>0:
        rule = mcfg.pop()
        head = rule.head
        body = rule.body
        if len(body)!=1:
            result.append(rule)
            continue
        # else:
        A0x = head.args
        k0  = len(A0x)
        i   = find_too_long(A0x)
        if i == k0: # no argument with two vars found
            result.append(rule)
            continue
        # else:
        si = A0x[i]
        s1 = si[0]
        s2 = si[1:]
        A0 = head.nterm
        A1 = new_nterm(dim,k0+1,A0)
        X1 = [ f'Z{n}' for n in range(0,k0)]
        X2 = [ [Zn] for Zn in X1]
        if is_var(s1): # s2 contains only terminals
            print(f"Transform step 6a: {A0}->{A1}", file=stderr)
            # add rule 1
            head1 = Head(A0, X2[0:i] + [ [X1[i]]+s2 ] + X2[i+1:k0])
            body1 = Body(A1, X1[0:k0])
            result.append(Rule(head1, [body1]))
            # add rule 2
            head2 = Head(A1,A0x[0:i] + [[s1]] + A0x[i+1:k0])
            result.append(Rule(head2, body))
        else:
            print(f"Transform step 6b: {A0}->{A1}", file=stderr)
            # add rule 1
            head1 = Head(A0, X2[0:i] + [ [s1,X1[i]] ] + X2[i+1:k0])
            body1 = Body(A1, X1[0:k0])
            result.append(Rule(head1, [body1]))
            # add rule 2
            head2 = Head(A1,A0x[0:i] + [s2] + A0x[i+1:k0])
            mcfg.append(Rule(head2, body))
    return result

# STEP 7

'''Create n new variables, fresh w.r.t. vars'''
def fresh_vars(n,vars):
    i = 0
    j = 0
    result = list()
    while j < n:
        X = f'Z{i}'
        i += 1
        if X not in vars:
            result.append(X)
            j += 1
    return result

def step7(mcfg,dim):
    result = []
    mcfg   = mcfg[:]
    while len(mcfg)>0:
        rule = mcfg.pop()
        head = rule.head
        body = rule.body
        if len(body)!=1:
            result.append(rule)
            continue
        # else:
        A0x = head.args
        k0  = len(A0x)
        i   = find_non_var(A0x)
        if i == k0: # no non-var argument found
            result.append(rule)
            continue
        # else:
        j = find_non_var(A0x[:i]+A0x[i+1:])
        if j == k0-1: # no other non-var argument found
            result.append(rule)
            continue
        # else: 
        si = A0x[i]
        xs = [ x for x in si if is_var(x) ]
        el = len(xs)
        A0 = head.nterm
        A1 = new_nterm(dim,k0+el-1,A0)
        print(f"Transform step 7: {A0}->{A1}", file=stderr)
        # add rule 1:
        X1 = fresh_vars(k0+el-1, xs)
        X2 = [ [Xn] for Xn in X1 ] 
        head1 = Head(A0, X2[0:i] + [si] + X2[i+el:k0+el])
        body1 = Body(A1, X1[0:i] +  xs  + X1[i+el:k0+el])
        # add rule 2:
        X = [ [x] for x in xs]
        result.append(Rule(head1,[body1]))
        head2 = Head(A1, A0x[0:i] + X + A0x[i+1:k0])
        mcfg.append(Rule(head2,body))
    return result

def normalize(mcfg,dim):
    mcfg = step1(mcfg,dim)
    mcfg = step2(mcfg,dim)
    mcfg = step3(mcfg,dim)
    mcfg = step4(mcfg,dim)
    mcfg = step5(mcfg,dim)
    mcfg = step6(mcfg,dim)
    mcfg = step7(mcfg,dim)
    return mcfg
    
if __name__ == "__main__":

    from glob import glob

    from mcfg_check import McfgError, check_mcfg
    from mcfg_data import degree, dimension, print_mcfg, rank
    from mcfg_parse import ParseError, Reader, read_mcfg

    def perform(step,mcfg,dim,text):
            print(f'\n{text}')
            result = step(mcfg,dim)
            if len(result) != len(mcfg):
                print_mcfg(result)
                check_mcfg(result)
            return result

    def main():
        for name in glob('Examples/*.mcfg') + glob('Test/norm*.mcfg'):
            print(f'\nTest case {name}:')
            try:
                reader = Reader(open(name))
                m = read_mcfg(reader)
                dim = check_mcfg(m)
            except (ParseError, McfgError, IOError) as err:
                print(f'{name}: {err}')
                continue
            print_mcfg(m)
            (d0,r0,deg0) = (dimension(m),rank(m),degree(m))
            m = perform(step1, m, dim, "Step 1:")            
            m = perform(step2, m, dim, "Step 2:")            
            m = perform(step3, m, dim, "Step 3:")            
            m = perform(step4, m, dim, "Step 4:")
            m = perform(step5, m, dim, "Step 5:")
            m = perform(step6, m, dim, "Step 6:")
            m = perform(step7, m, dim, "Step 7:")
            (d1,r1,deg1) = (dimension(m),rank(m),degree(m))
            if (d0,r0,deg0) == (d1,r1,deg1):
                print("Dimension, rank and degree are unchanged.")
            else:
                print("Dimension, rank or degree have changed!")
            print()
    main()
