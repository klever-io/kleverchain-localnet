package main

import (
	"os"

	"github.com/klever-io/kleverchain-localnet/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
