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

// Client to interface with the appinspect service
type Client struct {
	resty *resty.Client
	token string
}

// Error ...
type Error struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

// NewWithURL client to interface with the appinspect service
func NewWithURL(acsURL, token string) *Client {
	return &Client{
		resty: resty.New().SetHostURL(acsURL).SetError(&Error{}).SetAuthScheme("Bearer").SetAuthToken(token),
	}
}

// InstallApp ...
func (c *Client) InstallApp(stack, token, packageFileName string, packageReader io.Reader) error {
	resp, err := c.resty.R().SetFormData(map[string]string{"token": token}).
		SetFileReader("package", packageFileName, packageReader).
		Post("/" + stack + "/adminconfig/v1/apps")
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

// ListApps ...
func (c *Client) ListApps(stack string) (map[string]App, error) {
	resp, err := c.resty.R().SetResult(&map[string]App{}).Get("/" + stack + "/adminconfig/v1/apps")
	if err != nil {
		return nil, fmt.Errorf("error while listing app: %s", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error while listing apps: %s: %s", resp.Status(), resp.String())
	}
	apps, ok := resp.Result().(*map[string]App)
	if !ok {
		return nil, fmt.Errorf("error while parsing response")
	}
	return *apps, nil
}

// DescribeApp ...
func (c *Client) DescribeApp(stack string, appName string) (*App, error) {
	resp, err := c.resty.R().SetResult(&App{}).Get("/" + stack + "/adminconfig/v1/apps/" + appName)
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

// UninstallApp ...
func (c *Client) UninstallApp(stack string, appName string) error {
	resp, err := c.resty.R().Delete("/" + stack + "/adminconfig/v1/apps/" + appName)
	if err != nil {
		return fmt.Errorf("error while uninstalling app: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error while uninstalling apps: %s: %s", resp.Status(), resp.String())
	}
	return nil
}
