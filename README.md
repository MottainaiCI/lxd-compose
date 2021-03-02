# LXD Compose

[![Build Status](https://travis-ci.com/MottainaiCI/lxd-compose.svg?branch=master)](https://travis-ci.com/MottainaiCI/lxd-compose)
[![Go Report Card](https://goreportcard.com/badge/github.com/MottainaiCI/lxd-compose)](https://goreportcard.com/report/github.com/MottainaiCI/lxd-compose)

**lxd-compose** supply a way to deploy a complex environment to an LXD Cluster or LXD standalone installation.

It permits to organize and trace all configuration steps of infrastructure and create test suites.

All configuration files could be created at runtime through two different template engines: Mottainai or Jinja2 (require `j2cli` tool).

It's under heavy development phase and specification could be changed in the near future.

## Installation

**lxd-compose** is available as Mocaccino OS package and installable in every Linux
distro through [luet](https://luet-lab.github.io/docs/) tool with these steps:

```bash
$> curl https://get.mocaccino.org/luet/get_luet_root.sh | sudo sh
$> sudo luet install repository/mocaccino-extra-stable
$> sudo luet install app-emulation/lxd-compose
```

### Upgrade lxd-compose

```bash

$> sudo luet upgrade

```

### Use lxd-compose with LXD installed from snapd

LXD available through snapd doesn't expose local unix socket under default path
`/var/lib/lxd/unix.socket` but normally under the path `/var/snap/lxd/common/lxd/unix.socket`.

This means that to use `local` connection it's better to create under the config.yaml an entry like this:

```yaml
  local-snapd:
    addr: unix:///var/snap/lxd/common/lxd/unix.socket
    public: false
```

and then to use `local-snapd` in `connection` option.

Instead, if it's used the HTTPS API this is not needed.

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

