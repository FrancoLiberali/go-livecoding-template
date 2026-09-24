// Command cli is a scratch executable for plain programs — no HTTP, no gRPC.
// Use it for algorithm-style interview problems: put the logic in run() (which
// takes its args and output writer as parameters) so it stays unit-testable,
// and keep pure helpers like greet() separate from I/O.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run holds the program logic, decoupled from os.* so tests can drive it with
// arbitrary args and capture the output.
func run(args []string, out io.Writer) error {
	if _, err := fmt.Fprintln(out, greet(args)); err != nil {
		return err
	}

	return nil
}

// greet builds a greeting from the CLI args — replace with your solution.
func greet(args []string) string {
	who := "world"
	if len(args) > 0 {
		who = strings.Join(args, " ")
	}

	return "hello, " + who
}
