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

package acs

import (
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
)

type Client interface {
	InstallApp(stack, token, packageFileName string, packageReader io.Reader) error
	DescribeApp(stack string, appName string) (*App, error)
	ListApps(stack string) ([]App, error)
	UninstallApp(stack string, appName string) error
}

// victoriaClient is a client used to interface with ACS for Victoria stacks
type victoriaClient struct {
	client
}

// classicClient is a client used to interface with ACS for Classic (non-Victoria) stacks
type classicClient struct {
	client
}

type client struct {
	resty *resty.Client
}

type acsError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e *acsError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

func newClient(acsURL, token string) client {
	return client{
		resty: resty.New().SetHostURL(acsURL).SetError(&acsError{}).SetAuthScheme("Bearer").SetAuthToken(token),
	}
}

// NewVictoriaWithURL creates a new VictoriaClient
func NewVictoriaWithURL(acsURL, token string) Client {
	return &victoriaClient{
		client: newClient(acsURL, token),
	}
}

// NewClassicWithURL creates a new ClassicClient
func NewClassicWithURL(acsURL, token string) Client {
	return &classicClient{
		client: newClient(acsURL, token),
	}
}

// InstallApp installs an app on a classic stack
func (c *classicClient) InstallApp(stack, token, packageFileName string, packageReader io.Reader) error {
	resp, err := c.resty.R().SetFormData(map[string]string{"token": token}).
		SetFileReader("package", packageFileName, packageReader).
		Post("/" + stack + "/adminconfig/v2/apps")
	if err != nil {
		return fmt.Errorf("error while installing app: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error while submit: %s: %s", resp.Status(), resp.String())
	}
	return nil
}

// InstallApp installs an app on a victoria stack
func (c *victoriaClient) InstallApp(stack, token, packageFileName string, packageReader io.Reader) error {
	resp, err := c.resty.R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Proxy-Authorization", token).
		SetHeader("ACS-Legal-Ack", "Y").
		SetBody(packageReader).
		Post("/" + stack + "/adminconfig/v2/apps/victoria")
	if err != nil {
		return fmt.Errorf("error while installing app: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error while submit: %s: %s", resp.Status(), resp.String())
	}
	return nil
}

// App ...
type App struct {
	Label   *string `json:"label,omitempty"`
	Package *string `json:"package,omitempty"`
	Status  string  `json:"status"`
	Version *string `json:"version,omitempty"`
}

// ListApps on a classic stack
func (c *classicClient) ListApps(stack string) ([]App, error) {
	return listApps(c.client, fmt.Sprintf("/%s/adminconfig/v2/apps", stack))
}

// ListApps on a victoria stack
func (c *victoriaClient) ListApps(stack string) ([]App, error) {
	return listApps(c.client, fmt.Sprintf("/%s/adminconfig/v2/apps/victoria", stack))
}

func listApps(c client, url string) ([]App, error) {
	type listAppsResponse struct {
		Apps []App
	}
	resp, err := c.resty.R().SetResult(&listAppsResponse{}).Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while listing apps: %s", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error while listing apps: %s: %s", resp.Status(), resp.String())
	}
	apps, ok := resp.Result().(*listAppsResponse)
	if !ok {
		return nil, fmt.Errorf("error while parsing response")
	}
	return apps.Apps, nil
}

// DescribeApp on a classic stack
func (c *classicClient) DescribeApp(stack string, appName string) (*App, error) {
	return describeApp(c.client, fmt.Sprintf("/%s/adminconfig/v2/apps/%s", stack, appName))
}

// DescribeApp on a classic stack
func (c *victoriaClient) DescribeApp(stack string, appName string) (*App, error) {
	return describeApp(c.client, fmt.Sprintf("/%s/adminconfig/v2/apps/victoria/%s", stack, appName))
}

func describeApp(c client, url string) (*App, error) {
	resp, err := c.resty.R().SetResult(&App{}).Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while describing app: %s", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error while describing apps: %s: %s", resp.Status(), resp.String())
	}
	app, ok := resp.Result().(*App)
	if !ok {
		return nil, fmt.Errorf("error while parsing response")
	}
	return app, nil
}

// UninstallApp on a classic stack
func (c *classicClient) UninstallApp(stack string, appName string) error {
	return uninstallApp(c.client, fmt.Sprintf("/%s/adminconfig/v2/apps/%s", stack, appName))
}

// UninstallApp on a victoria stack
func (c *victoriaClient) UninstallApp(stack string, appName string) error {
	return uninstallApp(c.client, fmt.Sprintf("/%s/adminconfig/v2/apps/victoria/%s", stack, appName))
}

func uninstallApp(c client, url string) error {
	resp, err := c.resty.R().Delete(url)
	if err != nil {
		return fmt.Errorf("error while uninstalling app: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error while uninstalling apps: %s: %s", resp.Status(), resp.String())
	}
	return nil
}
