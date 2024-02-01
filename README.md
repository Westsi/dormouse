# Dormouse
A C-Like statically typed programming language.

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
