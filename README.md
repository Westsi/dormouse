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
- Codegen for x86_64, more architectures being worked on, starting with ARM

## Command Line Parameters
- `-d` - debug print
- `-a` - target architecture. supports x86_64, with aarch64 in the works
- `-v` - verbose
- `-o` - file name of output file which will be placed in `out/ARCHITECTURE`
