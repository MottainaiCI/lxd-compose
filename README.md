# LXD Compose
[![Build Status](https://github.com/MottainaiCI/lxd-compose/actions/workflows/push.yml/badge.svg)](https://github.com/MottainaiCI/lxd-compose/actions)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4753/badge)](https://bestpractices.coreinfrastructure.org/projects/4753)
[![Go Report Card](https://goreportcard.com/badge/github.com/MottainaiCI/lxd-compose)](https://goreportcard.com/report/github.com/MottainaiCI/lxd-compose)
[![CodeQL](https://github.com/MottainaiCI/lxd-compose/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/MottainaiCI/lxd-compose/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/MottainaiCI/lxd-compose/branch/master/graph/badge.svg?token=2nKASyitjI)](https://codecov.io/gh/MottainaiCI/lxd-compose)
[![Github All Releases](https://img.shields.io/github/downloads/MottainaiCI/lxd-compose/total.svg)](https://github.com/MottainaiCI/lxd-compose/releases)

**lxd-compose** supply a way to deploy a complex environment to an LXD Cluster or LXD standalone installation.

It permits organizing and tracing all infrastructure configuration steps and creating test suites, following
the IAAS (Infrastructure As A Code) paradigm.

All configuration files could be created at runtime through two different template engines: Helm or Jinja2 (require `j2cli` tool).

To keep API changes fast we haven't yet release a major released but we consider
the tool pretty stable.

From release `v0.33.0` lxd-compose uses by default the Instance API to works with `Incus`.

At the moment, we doesn't support VMs but we will add support to virtual-machine soon.

## Installation

**lxd-compose** is available as Macaroni OS package and installable in every Linux
distro through [luet](https://www.macaronios.org/docs/pms/#luet) tool with these steps:

```bash
$> curl https://raw.githubusercontent.com/macaroni-os/anise/macaroni/contrib/config/get_luet_root.sh | sudo sh
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

## Community

The lxd-compose devs team is available through the [Mottainai](https://join.slack.com/t/mottainaici/shared_invite/zt-zdmrc651-IvxE9j~TT5ssv_CVo51uZg) Slack channel.
