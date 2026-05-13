package main

import (
	"fmt"
	"io"
	"os"
	"os/user"

	"github.com/askerdev/monkeylang/evaluator"
	"github.com/askerdev/monkeylang/lexer"
	"github.com/askerdev/monkeylang/object"
	"github.com/askerdev/monkeylang/parser"
	"github.com/askerdev/monkeylang/repl"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "monkeylang",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			panic("unknown length of args, must be <= 1")
		} else if len(args) == 1 {
			file, err := os.ReadFile(args[0])
			if err != nil {
				panic(err)
			}

			env := object.NewEnvironment()
			l := lexer.New(string(file))
			p := parser.New(l)

			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				repl.PrintParserErrors(os.Stdout, p.Errors())
				return
			}

			evaluated := evaluator.Eval(program, env)
			if evaluated != nil {
				io.WriteString(os.Stdout, evaluated.Inspect())
				io.WriteString(os.Stdout, "\n")
			}
		} else {
			user, err := user.Current()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
			fmt.Printf("Feel free to type in commands\n")
			repl.Start(os.Stdin, os.Stdout)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
