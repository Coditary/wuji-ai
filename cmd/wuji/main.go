package main

import (
	"fmt"
	"os"

	"github.com/coditary/wuji-ai/internal/cli"
)

func main() {
	app, err := cli.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app.Execute()
}
