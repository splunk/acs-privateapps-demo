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
	AcsURL            string `kong:"env='ACS_URL',help='the acs url',default='https://admin.splunk.com'"`
	Victoria          bool   `kong:"help='whether the stack is a Victora stack'"`
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

	var cli acs.Client
	if i.Victoria {
		cli = acs.NewVictoriaWithURL(i.AcsURL, i.StackToken)
	} else {
		cli = acs.NewClassicWithURL(i.AcsURL, i.StackToken)
	}
	return cli.InstallApp(i.StackName, ar.Data.Token, filepath.Base(i.PackageFilePath), bytes.NewReader(pf))
}

type uninstall struct {
	StackName  string `kong:"arg,help='the splunk cloud stack'"`
	AppName    string `kong:"arg,optional,help='the app'"`
	StackToken string `kong:"env='STACK_TOKEN',help='the stack sc_admin jwt token'"`
	AcsURL     string `kong:"env='ACS_URL',help='the acs url',default='https://admin.splunk.com'"`
	Victoria   bool   `kong:"help='whether the stack is a Victora stack'"`
}

func (u *uninstall) Run(c *context) error {

	if u.StackToken == "" {
		survey.AskOne(&survey.Password{
			Message: "stack token:",
		}, &u.StackToken)
		fmt.Println("")
	}

	var cli acs.Client
	if u.Victoria {
		cli = acs.NewVictoriaWithURL(u.AcsURL, u.StackToken)
	} else {
		cli = acs.NewClassicWithURL(u.AcsURL, u.StackToken)
	}
	return cli.UninstallApp(u.StackName, u.AppName)
}
