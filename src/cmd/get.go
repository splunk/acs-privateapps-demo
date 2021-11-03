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
	"encoding/json"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/splunk/acs-privateapps-demo/src/acs"
)

type get struct {
	StackName  string `kong:"arg,help='the splunk cloud stack'"`
	AppName    string `kong:"arg,optional,help='the app'"`
	StackToken string `kong:"env='STACK_TOKEN',help='the stack sc_admin jwt token'"`
	AcsURL     string `kong:"env='ACS_URL',help='the acs url',default='https://admin.splunk.com'"`
	Victoria   bool   `kong:"help='whether the stack is a Victora stack'"`
}

func (g *get) Run(c *context) error {

	if g.StackToken == "" {
		survey.AskOne(&survey.Password{
			Message: "stack token:",
		}, &g.StackToken)
		fmt.Println("")
	}

	var cli acs.Client
	if g.Victoria {
		cli = acs.NewVictoriaWithURL(g.AcsURL, g.StackToken)
	} else {
		cli = acs.NewClassicWithURL(g.AcsURL, g.StackToken)
	}

	var object interface{}
	var err error
	if g.AppName == "" {
		object, err = cli.ListApps(g.StackName)
		if err != nil {
			return err
		}
	} else {
		object, err = cli.DescribeApp(g.StackName, g.AppName)
		if err != nil {
			return err
		}
	}

	if data, e := json.MarshalIndent(object, "", "    "); e == nil {
		fmt.Printf("%s\n", string(data))
	} else {
		fmt.Printf("%v\n", data)
	}
	return nil
}
