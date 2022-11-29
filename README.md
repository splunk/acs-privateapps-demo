# acs-privateapps-demo

This repository demonstrates how one can use [splunkcloud's self-service apis](https://www.splunk.com/en_us/blog/platform/splunk-cloud-self-service-announcing-the-new-admin-config-service-api.html) to build a pipeline that can continuously deploy Splunk apps to Splunk Enterprise Cloud stacks.

## Steps
The pipeline primarily consists of 4 steps:
1. Build cloudctl (`make build-cloudctl`), the CLI that will be used for the remaining steps -- this step assumes that [go](https://golang.org) is installed.
1. Package the app artifacts into a tar gz archive (`make generate-app-package`) -- this step assumes there is a top-level directory called `testapp` which contains the app.
1. Upload the app-package to the app inspect service and wait for the inspection report (`make inspect-app`) -- this step assumes the existence of the environment variables defined below.
1. If the inspection is successful, install/update the app on the stack using the self-serive apis (`make install-app`) -- this step also assumes the existence of the environment variables defined below.

The steps of the pipeline can be found [here](https://github.com/splunk/acs-privateapps-demo/blob/main/.github/workflows/main.yml).

## Note
* Few steps (app-vetting and app-installation) have Victoria and Classic variations in the Makefile.
* For Stacks in Victoria Experience: Make sure your Victoria stack in at least on Butterfinger (8.2.2112) to use this github demo.

## Setting up the environment
The environment needs to be configured with a few variables. If leveraging this from a Github repository using Github Actions workflows, the variables will need to be set up as [secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets). If running this locally, these values simply need to be set as environment variables:
* SPLUNK_COM_USERNAME / SPLUNK_COM_PASSWORD - the [splunk.com](https://login.splunk.com/) credentials to use for authentication to perform app inspection.
* STACK_NAME - the name of the Splunk Cloud stack where you want to install/update the app package on.
* STACK_TOKEN - the [jwt token](https://docs.splunk.com/Documentation/Splunk/latest/Security/Setupauthenticationwithtokens) created on the stack.


## Publishing a new version
This repository has been used as dependencies for other projects.

To manage the dependencies easily by specifying different versions for this module, we will use the module version numbering as recommended by Go - https://go.dev/doc/modules/version-numbers. The versioning will be done using GitHub tags.

To create a new tag, run the following command:
```shell
$ git tag vx.x.x
$ git push origin vx.x.x
```

Alternatively, you can use the GitHub action `Release Package` to release generate new tags.