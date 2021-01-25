package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/splunk/acs-privateapps-demo/src/appinspect"
)

type login struct {
	SplunkComUsername string `kong:"env='SPLUNK_COM_USERNAME',help='the splunkbase username'"`
	SplunkComPassword string `kong:"env='SPLUNK_COM_PASSWORD',help='the splunkbase password'"`
}

func (v *login) Run(c *context) error {

	if v.SplunkComUsername == "" {
		survey.AskOne(&survey.Input{
			Message: "splunkbase username:",
		}, &v.SplunkComUsername)
	}
	if v.SplunkComPassword == "" {
		survey.AskOne(&survey.Password{
			Message: "splunkbase password:",
		}, &v.SplunkComPassword)
		fmt.Println("")
	}
	res, err := appinspect.Authenticate(v.SplunkComUsername, v.SplunkComPassword)
	if err != nil {
		return err
	}
	fmt.Printf("Token: %s\n", res.Data.Token)
	return nil
}
