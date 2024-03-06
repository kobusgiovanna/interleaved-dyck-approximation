# Jaco van de Pol, Aarhus University, January 2022

# Parse "raw" MCFG rules from a file.

# Reserved symbols: ( ) , . :-
# Whitespace: ' ', \t, \n, % (comment)
# Parser does not distinguish names:
#   Non-terminals: should contain at least a start symbol S.
#   Variables: should start with Capitals
#   Terminals: all words that start with another symbol

from mcfg_data import Head, Body, Rule


class ParseError(Exception):
    pass


class LineNumber:
    def __init__(self):
        self.lno = 1  # line number used in error messages

    def inc(self):
        self.lno += 1


special = "%,.(): \n\t"

"""Iterate all tokens in the file, specific for this parser.
"""


def read_tokens(file, lno):
    c = file.read(1)
    while True:
        if c == "":
            yield None
        elif c in " \t\n":  # whitespace
            if c == "\n":  # count line
                lno.inc()
        elif c == "%":  # comment, count line
            while c not in ["\n", ""]:
                c = file.read(1)
            lno.inc()
        elif c == ",":
            yield ","
        elif c == ".":
            yield "."
        elif c == "(":
            yield "("
        elif c == ")":
            yield ")"
        elif c == ":":  # start of :-
            c = file.read(1)
            if c == "-":
                yield ":-"
            else:
                raise ParseError(f'":" should only appear in ":-"')
        elif c == "-":
            raise ParseError(f'"-" should only appear in ":-"')
        else:  # read an identifier
            word = [c]
            c = file.read(1)
            while c not in special:
                word.append(c)
                c = file.read(1)
            yield ("".join(word))
            continue  # we read one character too much already
        c = file.read(1)


"""This class can be used to read tokens.
   It supports lookahead, and pushing tokens back.
   Finally, it keeps track of line numbers for reporting.
"""


class Reader:
    def __init__(self, file):
        self.lineNumber = LineNumber()
        self.reader = read_tokens(file, self.lineNumber)
        self.queue = list()  # used to push tokens / store look-ahead

    """Return and consume next token"""

    def next(self):
        if len(self.queue) > 0:
            return self.queue.pop(0)
        return next(self.reader)

    """Return next token, but don't consume"""

    def lookahead(self):
        if len(self.queue) > 0:
            return self.queue[0]
        token = next(self.reader)
        self.queue.append(token)
        return token

    """Push a token back; currently not used"""

    def push(self, c):
        self.queue.append(c)

    """Return the current line number"""

    def line_number(self):
        return self.lineNumber.lno


"""Read an expected token"""


def read_special(reader, expected):
    token = reader.next()
    if token != expected:
        if token == None:
            token = "End-of-File"
        raise ParseError(f'"{expected}" expected instead of "{token}"')


"""Read a single name"""


def read_name(reader):
    token = reader.next()
    if token[0] in special:
        raise ParseError(f'some "name" is expected instead of "{token}"')
    return token


"""Read a word of names, separated by whitespace"""


def read_word(reader):
    def enum_word():
        yield read_name(reader)
        while reader.lookahead() not in special:
            yield read_name(reader)

    return list(enum_word())


"""Read a list of expressions, separated by comma's"""


def read_list(expr, reader):
    def enum_list():
        yield expr(reader)
        while reader.lookahead() == ",":
            reader.next()
            yield expr(reader)

    return list(enum_list())


"""Read an atom 'A(expr,...,expr)' """


def read_atom(expr, reader):
    A = read_name(reader)
    read_special(reader, "(")
    vars = read_list(expr, reader)
    read_special(reader, ")")
    return (A, vars)


"""Read a Body 'A(name,...,name)' """


def read_body(reader):
    return Body(*read_atom(read_name, reader))


"""Read a Head 'A(word,...,word)' """


def read_head(reader):
    return Head(*read_atom(read_word, reader))


"""Enumerate each Rule 'Head [:- Body,...,Body].' """


def enum_rule(reader):
    try:
        while reader.lookahead() != None:
            head = read_head(reader)
            if reader.lookahead() == ".":
                body = []
            else:
                read_special(reader, ":-")
                body = read_list(read_body, reader)
            read_special(reader, ".")
            yield Rule(head, body)
        return []  # EOF reached
    except ParseError as err:
        raise ParseError(f"Parse error in Line {reader.line_number()}: {err}")


"""Read an MCFG"""


def read_mcfg(reader):
    return list(enum_rule(reader))


if __name__ == "__main__":
    from glob import glob

    def main():
        for file in glob("Examples/*.mcfg") + ["blooper.mfcg"] + glob("Test/*.mcfg"):
            print(f"File: {file}")
            try:
                reader = Reader(open(f"{file}"))
                for rule in enum_rule(reader):
                    print(rule)
            except (ParseError, IOError) as err:
                print(f"{file}: {err}")
            print()

    main()
