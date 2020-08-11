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
$> curl https://gist.githubusercontent.com/mudler/8b8d6c53c4669f4b9f9a72d1a2b92172/raw/e9d38b8e0702e7f1ef9a5db1bfa428add12a2d24/get_luet_root.sh | sudo sh
$> sudo luet install repository/mocaccino-extra
$> sudo luet install app-emulation/lxd-compose
```

### Upgrade lxd-compose

```bash

$> sudo luet upgrade app-emulation/lxd-compose

```

## Getting Started

### Deploy an environment

```bash

$> lxd-compose apply myproject

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

```

### Diagnose loaded variables

```bash

$> lxd-compose diagnose vars project1

```


