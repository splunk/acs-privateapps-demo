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
	ForSplunkCloud    bool   `kong:"env='FOR_SPLUNKCLOUD',help='perform inspections for splunk cloud',default='true'"`
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
	submitRes, err := cli.Submit(filepath.Base(v.PackageFilePath), bytes.NewReader(pf), v.ForSplunkCloud)
	if err != nil {
		return err
	}
	fmt.Printf("submitted app for inspection (requestId='%s')\n", submitRes.RequestID)

	status, err := cli.Status(submitRes.RequestID)
	if err != nil {
		return err
	}
	if status.Status == "PROCESSING" || status.Status == "PREPARING" {
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
		fmt.Printf("vetting completed successfully\n")
	} else {
		err = fmt.Errorf("vetting failed with status='%s'", status.Status)
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
