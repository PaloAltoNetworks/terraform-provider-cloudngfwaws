Terraform Provider for Palo Alto Networks Cloud NGFW for AWS
============================================================

The Terraform provider for the Palo Alto Networks Cloud Next-Gen Firewall for AWS.

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) v0.14+
- [Go](https://golang.org) v1.15+ (to build the provider)

Testing the Provider
--------------------

In order to test the provider, you can use `make test` in order to run the acceptance tests for the provider.

**Note:** acceptance tests create real resources, and often cost money to run:

```sh
make test
```

Building the Provider
---------------------

1. Install [GoLang](https://go.dev/dl/)

2. Clone the needed repositories side-by-side to a new directory

```sh
mkdir cloudngfwaws-terraform
cd cloudngfwaws-terraform
git clone https://github.com/PaloAltoNetworks/cloud-ngfw-aws-go.git
git clone https://github.com/PaloAltoNetworks/terraform-provider-cloudngfwaws
```

3. Build the provider

```sh
cd terraform-provider-cloudngfwaws
make
```

4. Navigate to `/Users/<user-name>/.terraform.d/plugins` folder and create a folder structure as per the image below. When testing out a local provider, Terraform expects the folder structure to be HOSTNAME/NAMESPACE/TYPE/VERSION/TARGET (under `/Users/<user-name>/.terraform.d/plugins`) where

HOSTNAME: Needs to be in FQDN format. Default is `registry.terraform.io`. So to mimic that, we have created `terraform.local`.
NAMESPACE: Default is hashicorp. We have named it local, to denote that we are storing local providers here.
TYPE: Name of your provider. In our case cloudngfwaws.
VERSION: Version of the provider in string format, like 1.0.0.
TARGET: Specifies a particular target platform using a format like darwin_amd64 (MacOS), linux_arm, windows_amd64, etc.

The image below shows setup on a Mac. If you are setting it up for a different platform, please change the folder name accordingly.
The final part id the name of the provider. When Terraform search for a provider locally, it expects it be named as `terraform-provider-<TYPE>_v<VERSION>`. If you do not follow this format, it errors like below

`Error: Failed to install provider - Error while installing terraform.local/local/cloudngfwaws v1.0.0: provider binary not found: could not find executable file starting with terraform-provider-cloudngfwaws`

<img width="405" alt="image" src="https://user-images.githubusercontent.com/56643631/165510930-8fb70302-b2ba-425b-8d56-c53c6f65037c.png">

Once the above setup is done, check the information under ./docs on how to use it in your code.

