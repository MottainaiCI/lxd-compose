module github.com/MottainaiCI/lxd-compose

go 1.14

replace github.com/docker/docker => github.com/Luet-lab/moby v17.12.0-ce-rc1.0.20200605210607-749178b8f80d+incompatible

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/MottainaiCI/mottainai-server v0.0.0-20210508100055-c7e87a8199bf
	github.com/flosch/pongo2 v0.0.0-20200529170236-5abacdfa4915 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/gosexy/gettext v0.0.0-20160830220431-74466a0a0c4a // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/icza/dyno v0.0.0-20200205103839-49cb13720835
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a // indirect
	github.com/juju/persistent-cookiejar v0.0.0-20171026135701-d5e5a8405ef9 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lxc/lxd v0.0.0-20210507013419-5ff45e701cb8
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/mod v0.3.1-0.20200828183125-ce943fd02449 // indirect
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd
	gopkg.in/macaroon-bakery.v2 v2.2.0 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.5.0
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)

replace github.com/renstrom/dedent v1.1.0 => github.com/lithammer/dedent v1.1.0

replace github.com/Sirupsen/logrusv1.7.0 => github.com/sirupsen/logrus v1.7.0
