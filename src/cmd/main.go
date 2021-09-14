// Copyright 2021 Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
