// Command miramar is the CLI entry point. See ROADMAP.md for what's implemented.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styleFail(err.Error()))
		os.Exit(1)
	}
}
