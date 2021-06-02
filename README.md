# acs-privateapps-demo

This repository demonstrates how one can use the [splunkcloud's self-service apis](https://www.splunk.com/en_us/blog/platform/splunk-cloud-self-service-announcing-the-new-admin-config-service-api.html) to build a pipeline that can continuously deploy splunk apps to the splunk enterprise cloud stacks.

## Steps
The pipeline primarily consists of 3 steps:
1. Packages the app artifacts into a tar gz archive.
1. Uploads the app-package to the app inspect service and wait for the inspection report. 
1. If the inspection is successful, install/update the app on the stack using the self-serive apis.

## Setting up the environment
The CI needs to be configured with a few variables:
* SPLUNK_COM_USERNAME / SPLUNK_COM_PASSWORD - the [splunk.com](https://login.splunk.com/)  credentials to use for authentication to perform app inspection. 
* STACK_NAME - the name of the stack where the CI will install/update the app package on
* STACK_TOKEN - the [jwt token](https://docs.splunk.com/Documentation/Splunk/latest/Security/Setupauthenticationwithtokens) created on the stack.
