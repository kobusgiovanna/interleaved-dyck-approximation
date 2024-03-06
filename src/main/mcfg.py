#!/usr/bin/python3

# Jaco van de Pol, Aarhus University, January 2022

import sys
import os
import getopt
from mcfg_data import print_mcfg
from mcfg_parse import Reader, read_mcfg, ParseError
from mcfg_check import check_and_exit
from mcfg_rank import rank2, chomsky
from mcfg_normal import normalize
from mcfg_permute import permute_mcfg


def short():
    print(
        "Usage: mcfg [options] [- | <input-file>[.mcfg]] [<output-file>[.mcfg]]",
        file=sys.stderr,
    )
    print("Options: [-h/--help , -n/--norm, -p/--perm, -r/--rank]", file=sys.stderr)


def usage():
    print("Transform Multiple Context-Free Grammars", file=sys.stderr)
    short()
    print(
        """\
    [-n/--norm] : transform the mcfg to normal form
    [-p/--perm] : transform the mcfg to be non-deleting, non-permuting
    [-r/--rank] : transform the mcfg to rank at most 2 (not yet implemented)
    [<input-file>[.mcfg]] : file to read the input mcfg from (default/-: stdin)
    [<output-file>[.mcfg]]: file to output the final mcfg to (default: stdout)\
    """,
        file=sys.stderr,
    )


def open_mcfg_file(name, mode="r"):
    try:
        (_, ext) = os.path.splitext(name)
        if ext != ".mcfg":
            name = name + ".mcfg"
        return (name, open(name, mode=mode))
    except IOError as err:
        print(f"file {name}: {err}", file=sys.stderr)
        sys.exit(2)


def main():
    chom = False
    norm = False
    perm = False
    rank = False
    inname = "stdin"
    outname = "stdout"
    outfile = sys.stdout
    infile = sys.stdin

    try:
        args = sys.argv[1:]
        opts, args = getopt.gnu_getopt(
            args, "hcnpr", ["help", "chomsky", "norm", "perm", "rank"]
        )
    except getopt.GetoptError:
        short()
        sys.exit(2)
    for opt, arg in opts:
        if opt in ["-h", "--help"]:
            usage()
            sys.exit()
        elif opt in ["-c", "--chomsky"]:
            chom = True
        elif opt in ["-n", "--norm"]:
            norm = True
        elif opt in ["-p", "--perm"]:
            perm = True
        elif opt in ["-r", "--rank"]:
            rank = True
    if len(args) >= 1 and args[0] != "-":
        (inname, infile) = open_mcfg_file(args[0])
    if len(args) >= 2:
        (outname, outfile) = open_mcfg_file(args[1], mode="w")
    if len(args) > 2:
        short()
        sys.exit(2)

    print("\nmcfg transformer:", file=sys.stderr)
    print(f"Reading from {inname}", file=sys.stderr)
    print(f"Writing into {outname}", file=sys.stderr)

    print("=== Parsing ===", file=sys.stderr)
    try:
        reader = Reader(infile)
        m = read_mcfg(reader)
    except ParseError as err:
        print(f"{inname}: {err}", file=sys.stderr)
        sys.exit(2)

    print("=== Checking static semantics ===", file=sys.stderr)
    dim = check_and_exit(m)

    if chom:
        print("=== Applying Chomsky Transform ===", file=sys.stderr)
        m = chomsky(m, dim)
        check_and_exit(m)

    if rank:
        print("=== Reducing to rank 2 ===", file=sys.stderr)
        m = rank2(m, dim)
        check_and_exit(m)

    if perm:
        print("=== Making non-permuting ===", file=sys.stderr)
        (m, dim) = permute_mcfg(m)
        check_and_exit(m)

    if norm:
        print("=== Normalizing ===", file=sys.stderr)
        m = normalize(m, dim)
        check_and_exit(m)

    print_mcfg(m, file=outfile)


main()
