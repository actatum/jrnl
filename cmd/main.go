// Package main provides the entrypoint to the TUI journal.
package main

import (
	"fmt"
	"os"

	"github.com/actatum/jrnl/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
