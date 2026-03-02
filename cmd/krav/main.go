package main

import (
	"os"

	"go.krav.sh/krav/internal/cmd"
)

func main() {
	os.Exit(int(cmd.Main()))
}
