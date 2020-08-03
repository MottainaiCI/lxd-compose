module github.com/MottainaiCI/lxd-compose

go 1.14

replace github.com/docker/docker => github.com/Luet-lab/moby v17.12.0-ce-rc1.0.20200605210607-749178b8f80d+incompatible

require (
	github.com/MottainaiCI/mottainai-cli v0.0.0-20190629163247-be90396f998d
	github.com/MottainaiCI/mottainai-server v0.0.0-20200319175456-fc3c442fd4a6
	github.com/flosch/pongo2 v0.0.0-20200529170236-5abacdfa4915 // indirect
	github.com/geaaru/time-master v0.0.0-20200801154724-b41fecc1f570
	github.com/gosexy/gettext v0.0.0-20160830220431-74466a0a0c4a // indirect
	github.com/jaypipes/ghw v0.6.1 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a // indirect
	github.com/juju/persistent-cookiejar v0.0.0-20171026135701-d5e5a8405ef9 // indirect
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lxc/lxd v0.0.0-20190810000350-cfa3c9083b40
	github.com/mudler/luet v0.0.0-20200717204249-ffa6fc3829d2
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	go.uber.org/zap v1.15.0
	golang.org/x/sys v0.0.0-20200727154430-2d971f7391a4
	gopkg.in/clog.v1 v1.2.0 // indirect
	gopkg.in/macaroon-bakery.v2 v2.2.0 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)
