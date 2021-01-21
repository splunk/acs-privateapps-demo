package main

import (
	"github.com/alecthomas/kong"
)

type context struct {
	Debug bool
}

var cli struct {
	Debug bool `kong:"help='enable debug mode'"`
	Vet   vet  `kong:"cmd,help='vet the app package against the app-inspect service'"`
}

func main() {
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
