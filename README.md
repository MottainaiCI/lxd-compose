# LXD Compose
[![Build Status](https://github.com/MottainaiCI/lxd-compose/actions/workflows/push.yml/badge.svg)](https://github.com/MottainaiCI/lxd-compose/actions)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4753/badge)](https://bestpractices.coreinfrastructure.org/projects/4753)
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
From release `v0.39.0` lxd-compose implements two different executors, one based on LXD API and one based on Incus API.
The attribute `connection_type` is been added with values `incus` or `lxd-6` in order explicit the target group server.
The `connection_type` with value `lxd-6` could be used for LXD <6.0.

At the moment, we doesn't support VMs but we will add support to virtual-machine soon.

## Installation

**lxd-compose** is available as Macaroni OS package and installable in every Linux
distro through [anise](https://www.macaronios.org/docs/pms/#luet) tool with these steps:

```bash
$> curl https://raw.githubusercontent.com/macaroni-os/anise/geaaru/contrib/config/get_anise_root.sh | sudo -E sh
$> sudo anise install app-emulation/lxd-compose
```

### Upgrade lxd-compose

```bash

$> sudo anise repo update
$> sudo anise upgrade

```

## Documentation

The complete *lxd-compose* documentation is available [here](https://mottainaici.github.io/lxd-compose-docs/).

## Examples

We maintain a project that supply ready to build environments at [LXD Compose Galaxy](https://github.com/MottainaiCI/lxd-compose-galaxy).

## Community

The lxd-compose devs team is available through the [Mottainai](https://join.slack.com/t/mottainaici/shared_invite/zt-zdmrc651-IvxE9j~TT5ssv_CVo51uZg) Slack channel.
