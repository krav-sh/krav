package main

import (
	"os"

	"pkg.krav.sh/krav/internal/cmd"
)

func main() {
	os.Exit(int(cmd.Main()))
}
