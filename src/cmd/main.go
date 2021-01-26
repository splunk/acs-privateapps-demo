package main

import (
	"github.com/alecthomas/kong"
)

type context struct {
	Debug bool
}

var cli struct {
	Debug     bool      `kong:"help='enable debug mode'"`
	Login     login     `kong:"cmd,help='login to splunkbase and generate token'"`
	Vet       vet       `kong:"cmd,help='vet the app package against the app-inspect service'"`
	Install   install   `kong:"cmd,help=install the app package on the splunk stack"`
	Uninstall uninstall `kong:"cmd,help=uninstall the app package from the splunk stack"`
	Get       get       `kong:"cmd,help=get an app/apps installed on the splunk stack"`
}

func main() {
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
