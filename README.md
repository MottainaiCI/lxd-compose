# LXD Compose

**lxd-compose** supply a way to deploy a complex environment to an LXD Cluster or LXD standalone installation.

It permits to organize and trace all configuration steps of infrastructure and create test suites.

All configuration files could be created at runtime through two different template engines: Mottainai or Jinja2 (require `j2cli` tool).

It's under heavy development phase and specification could be changed in the near future.

## Deploy an environment

```bash

$> lxd-compose apply myproject

```

## Destroy an environment 

```bash

$> lxd-compose destroy myproject

```
