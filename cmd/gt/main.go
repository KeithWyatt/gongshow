// gt is the GongShow CLI for managing multi-agent workspaces.
package main

import (
	"os"

	"github.com/KeithWyatt/gongshow/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
