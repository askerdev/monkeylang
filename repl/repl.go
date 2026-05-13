package repl

import (
	"io"

	"github.com/askerdev/monkeylang/evaluator"
	"github.com/askerdev/monkeylang/lexer"
	"github.com/askerdev/monkeylang/object"
	"github.com/askerdev/monkeylang/parser"
	"github.com/chzyer/readline"
)

const PROMPT = ">> "

func Start(in io.ReadCloser, out io.Writer) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
		Stdin:  in,
		Stdout: out,
		Stderr: out,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	env := object.NewEnvironment()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func PrintParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
