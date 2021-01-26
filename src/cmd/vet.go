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
	submitRes, err := cli.Submit(filepath.Base(v.PackageFilePath), bytes.NewReader(pf))
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
