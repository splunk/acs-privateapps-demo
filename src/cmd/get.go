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
	AcsURL     string `kong:"env='ACS_URL',help='the acs url',default:'http://localhost:8443/'"`
}

func (g *get) Run(c *context) error {

	if g.StackToken == "" {
		survey.AskOne(&survey.Password{
			Message: "stack token:",
		}, &g.StackToken)
		fmt.Println("")
	}

	cli := acs.NewWithURL(g.AcsURL, g.StackToken)

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
