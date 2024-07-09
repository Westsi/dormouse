# Dormouse
A C-Like statically typed programming language.

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/Westsi/dormouse/ci.yml?style=for-the-badge&logo=github)


## Why another one?
- For me to learn

## Running
Download repo, install Go. `go run . PATH_TO_FILE`. As easy as it gets.

## How it works
1. Lexing, see `lex`
2. Pratt parser, see `parse`
3. Generate assembly for the system, see `codegen`
4. Compile assembly with default system assembler, see `codegen`.

## Features
- Parsing and lexing for a good chunk of the syntax.
- Codegen for x86_64 and aarch64.

## Command Line Parameters
- `-d` - debug print
- `-a` - target architecture. supports x86_64, and aarch64 without a couple features of x86_64.
- `-v` - verbose
- `-o` - file name of output file which will be placed in `out/ARCHITECTURE`

## SSA
To see output, view `out/ssa`. An example program is shown below.

Before optimizations:
```
func main(int v1, int v2) int {
    int v3 = 7 + v1
    int v4 = 8 + v2
    int v5 = v3 + v4
    return v5
}
```

After optimizations:
```
func main(int v1, int v2) int {
    v5 = 15 + v1 + v2
    return v5
}
```
