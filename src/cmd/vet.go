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
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/splunk/acs-privateapps-demo/src/appinspect"
	"io/ioutil"
	"path/filepath"
	"time"
)

type vet struct {
	PackageFilePath   string `kong:"arg,help='the path to the app-package (tar.gz) file',type='path'"`
	SplunkComUsername string `kong:"env='SPLUNK_COM_USERNAME',help='the splunkbase username'"`
	SplunkComPassword string `kong:"env='SPLUNK_COM_PASSWORD',help='the splunkbase password'"`
	JSONReportFile    string `kong:"help='the file to write the inspection report in json format',type='path'"`
	Victoria          bool   `kong:"help='whether the stack is a Victora stack'"`
}

func (v *vet) Run(c *context) error {

	pf, err := ioutil.ReadFile(v.PackageFilePath)
	if err != nil {
		return err
	}
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
	cli := appinspect.New()
	err = cli.Login(v.SplunkComUsername, v.SplunkComPassword)
	if err != nil {
		return err
	}
	submitRes, err := cli.Submit(filepath.Base(v.PackageFilePath), bytes.NewReader(pf), v.Victoria)
	if err != nil {
		return err
	}
	fmt.Printf("submitted app for inspection (requestId='%s')\n", submitRes.RequestID)

	status, err := cli.Status(submitRes.RequestID)
	if err != nil {
		return err
	}
	if status.Status == "PROCESSING" || status.Status == "PREPARING" || status.Status == "PENDING" {
		fmt.Printf("waiting for inspection to finish...\n")
		for {
			time.Sleep(2 * time.Second)
			status, err = cli.Status(submitRes.RequestID)
			if err != nil {
				return err
			}
			if status.Status != "PROCESSING" {
				break
			}
		}
	}
	if status.Status == "SUCCESS" {
		data, _ := json.MarshalIndent(status.Info, "", "    ")
		fmt.Printf("vetting completed, summary: \n%s\n", string(data))
		if status.Info.Failure > 0 || status.Info.Error > 0 {
			err = fmt.Errorf("vetting failed (failures=%d, errors=%d)", status.Info.Failure, status.Info.Error)
		}
	} else {
		err = fmt.Errorf("vetting failed to complete (status='%s')", status.Status)
	}

	if v.JSONReportFile != "" {
		report, e := cli.ReportJSON(submitRes.RequestID)
		if e != nil {
			fmt.Printf("failed to pull report: %s\n", e)
		}
		data, _ := json.MarshalIndent(report, "", "    ")
		e = ioutil.WriteFile(v.JSONReportFile, data, 0644)
		if e != nil {
			fmt.Printf("failed to write report: %s\n", e)
		}
	}
	return err
}
