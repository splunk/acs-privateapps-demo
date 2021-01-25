package main

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/splunk/acs-privateapps-demo/src/acs"
	"github.com/splunk/acs-privateapps-demo/src/appinspect"
	"io/ioutil"
	"path/filepath"
)

type install struct {
	StackName         string `kong:"arg,help='the splunk cloud stack'"`
	SplunkComUsername string `kong:"env='SPLUNK_COM_USERNAME',help='the splunkbase username'"`
	SplunkComPassword string `kong:"env='SPLUNK_COM_PASSWORD',help='the splunkbase password'"`
	PackageFilePath   string `kong:"arg,help='the path to the app-package (tar.gz) file',type='path'"`
	StackToken        string `kong:"env='STACK_TOKEN',help='the stack sc_admin jwt token'"`
	AcsURL            string `kong:"env='ACS_URL',help='the acs url',default:'http://localhost:8443/'"`
}

func (i *install) Run(c *context) error {

	pf, err := ioutil.ReadFile(i.PackageFilePath)
	if err != nil {
		return err
	}
	if i.SplunkComUsername == "" {
		survey.AskOne(&survey.Input{
			Message: "splunkbase username:",
		}, &i.SplunkComUsername)
	}
	if i.SplunkComPassword == "" {
		survey.AskOne(&survey.Password{
			Message: "splunkbase password:",
		}, &i.SplunkComPassword)
		fmt.Println("")
	}

	if i.StackToken == "" {
		survey.AskOne(&survey.Password{
			Message: "stack token:",
		}, &i.StackToken)
		fmt.Println("")
	}
	ar, err := appinspect.Authenticate(i.SplunkComUsername, i.SplunkComPassword)
	if err != nil {
		return err
	}

	cli := acs.NewWithURL(i.AcsURL, i.StackToken)
	return cli.InstallApp(i.StackName, ar.Data.Token, filepath.Base(i.PackageFilePath), bytes.NewReader(pf))
}

type uninstall struct {
	StackName  string `kong:"arg,help='the splunk cloud stack'"`
	AppName    string `kong:"arg,optional,help='the app'"`
	StackToken string `kong:"env='STACK_TOKEN',help='the stack sc_admin jwt token'"`
	AcsURL     string `kong:"env='ACS_URL',help='the acs url',default:'http://localhost:8443/'"`
}

func (u *uninstall) Run(c *context) error {

	if u.StackToken == "" {
		survey.AskOne(&survey.Password{
			Message: "stack token:",
		}, &u.StackToken)
		fmt.Println("")
	}

	cli := acs.NewWithURL(u.AcsURL, u.StackToken)
	return cli.UninstallApp(u.StackName, u.AppName)
}
