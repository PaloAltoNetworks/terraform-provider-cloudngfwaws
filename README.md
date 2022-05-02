Terraform Provider for Palo Alto Networks Cloud NGFW for AWS
============================================================

**NOTE:**  This Terraform provider is currently available as BETA code.  Once this is officially released, it will be available from the Terraform registry just like all other providers.

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

4. Specify the `dev_overrides` configuration per the next section below. This tells Terraform where to find the provider you just built. The directory to specify is the full path to the cloned provider repo. For example: `/home/myuser/cloudngfwaws-terraform/terraform-provider-cloudngfwaws`

When using the provider, refer to the documentation in the `./docs` directory for all resources and parameters.

Developing the Provider
-----------------------

With Terraform v0.14 and later, [development overrides for provider developers](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers) can be leveraged in order to use the provider built from source.

To do this, populate a Terraform CLI configuration file (`~/.terraformrc` for all platforms other than Windows; `terraform.rc` in the `%APPDATA%` directory when using Windows) with at least the following options:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/paloaltonetworks/cloudngfwaws" = "/directory/containing/the/cloudngfwaws/binary/here"
  }

  direct {}
}
```
