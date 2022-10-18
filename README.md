# LXD Compose

[![Build Status](https://travis-ci.com/MottainaiCI/lxd-compose.svg?branch=master)](https://travis-ci.com/MottainaiCI/lxd-compose)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4753/badge)](https://bestpractices.coreinfrastructure.org/projects/4753)
[![Go Report Card](https://goreportcard.com/badge/github.com/MottainaiCI/lxd-compose)](https://goreportcard.com/report/github.com/MottainaiCI/lxd-compose)
[![CodeQL](https://github.com/MottainaiCI/lxd-compose/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/MottainaiCI/lxd-compose/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/MottainaiCI/lxd-compose/branch/master/graph/badge.svg?token=2nKASyitjI)](https://codecov.io/gh/MottainaiCI/lxd-compose)
[![Github All Releases](https://img.shields.io/github/downloads/MottainaiCI/lxd-compose/total.svg)](https://github.com/MottainaiCI/lxd-compose/releases)

**lxd-compose** supply a way to deploy a complex environment to an LXD Cluster or LXD standalone installation.

It permits to organize and trace all configuration steps of infrastructure and create test suites.

All configuration files could be created at runtime through two different template engines: Mottainai or Jinja2 (require `j2cli` tool).

It's under heavy development phase and specification could be changed in the near future.

## Installation

**lxd-compose** is available as Macaroni OS package and installable in every Linux
distro through [luet](https://luet-lab.github.io/docs/) tool with these steps:

```bash
$> curl https://raw.githubusercontent.com/geaaru/luet/geaaru/contrib/config/get_luet_root.sh | sudo sh
$> sudo luet install app-emulation/lxd-compose
```

### Upgrade lxd-compose

```bash

$> sudo luet repo update
$> sudo luet upgrade

```

## Documentation

The complete *lxd-compose* documentation is available [here](https://mottainaici.github.io/lxd-compose-docs/).

## Examples

We maintain a project that supply ready to build environments at [LXD Compose Galaxy](https://github.com/MottainaiCI/lxd-compose-galaxy).

## Getting Started

### Deploy an environment

```bash

$> lxd-compose apply myproject

# Disable hooks with flag foo
$> lxd-compose apply --disable-flag foo

# Execute only hooks with flag foo
$> lxd-compose apply --enable-flag foo

```


### Destroy an environment

```bash

$> lxd-compose destroy myproject

```

### Stop an environment

For environment with containers not ephemeral.

```bash
$> lxd-compose stop myproject
```

### Validate environemnts


```bash

$> lxd-compose validate

```

### Create single node

```bash

$> lxd-compose node create node1 --hooks

# Execute only hooks with flag foo
$> lxd-compose node create node1 --hooks --enable-flag foo

# Disable hooks with flag foo
$> lxd-compose node create node1 --hooks --disable-flag foo

```

### Diagnose loaded variables

```bash

$> lxd-compose diagnose vars project1

```

### Show list of the project

```bash

$> lxd-compose project list

```

## Community

The lxd-compose devs team is available through the [Mottainai](https://join.slack.com/t/mottainaici/shared_invite/zt-zdmrc651-IvxE9j~TT5ssv_CVo51uZg) Slack channel.
