# Terraform-restrictor

A small binary which read a [terraform plan file](https://www.terraform.io/docs/commands/plan.html#out-path) 
to perform some check. This was primarily made for continuous integration.

## Installation

### On *nix system

You can install this via the command-line with either `curl` or `wget`.

#### via curl

```bash
$ sh -c "$(curl -fsSL https://raw.github.com/orange-cloudfoundry/terraform-restrictor/master/bin/install.sh)"
```

#### via wget

```bash
$ sh -c "$(wget https://raw.github.com/orange-cloudfoundry/terraform-restrictor/master/bin/install.sh -O -)"
```

### On windows

You can install it by downloading the `.exe` corresponding to your cpu from releases page: https://github.com/orange-cloudfoundry/terraform-restrictor/releases .
Alternatively, if you have terminal interpreting shell you can also use command line script above, it will download file in your current working dir.

### From go command line

Simply run in terminal:

```bash
$ go get github.com/orange-cloudfoundry/terraform-restrictor
```

## Usage

You will need to create a plan file by using:

```bash
$ terraform plan -out=myplan.plan
```

Create a `restrictions.yml` file in this format:

```yml
- type: .* # restrict to provider type, regex are allowed
  unauthorized: ["create", "update", "delete"] # restrict on different methods
  name: test.* # restrict against ressource name, regex are allowed
  check_attrs: # validating a particular attribute on this type and name
  - key: "name" # select your attribute, regex are allowed
    unauthorized: ["create", "update", "delete"] # restrict on different methods, this can be empty
    validate: ["toto"] # Validate attribute value against a regex, this is useful to validate naming for example, define multiple perform an `and` 
``` 

Now you can validate by running `terraform-restrictor` as follow:

```bash
$ terraform-restrictor ./myplan.plan
```

## Help page

```
terraform-restrictor: Usage:
  terraform-restrictor [OPTIONS] PATH

Application Options:
  -f, --file=    Path to the restrictions definition yaml file (default: restrictions.yml)
  -v, --verbose  Verbose output

Help Options:
  -h, --help     Show this help message

Arguments:
  PATH:          Path to a terraform plan file (use - to load plan from stdin)
```