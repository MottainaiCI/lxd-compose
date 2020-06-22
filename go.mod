module github.com/MottainaiCI/lxd-compose

go 1.14

require (
	github.com/MottainaiCI/mottainai-server v0.0.0-20200319175456-fc3c442fd4a6
	github.com/fsouza/go-dockerclient v1.6.5 // indirect
	github.com/jaypipes/ghw v0.6.1 // indirect
	github.com/mudler/luet v0.0.0-20200612174137-ee3b59348e36
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	gopkg.in/clog.v1 v1.2.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/yaml.v2 v2.2.7
)

replace github.com/docker/docker => github.com/Luet-lab/moby v17.12.0-ce-rc1.0.20200605210607-749178b8f80d+incompatible
