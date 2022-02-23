module github.com/MottainaiCI/lxd-compose

go 1.17

replace google.golang.org/grpc/naming => google.golang.org/grpc v1.29.1

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/icza/dyno v0.0.0-20200205103839-49cb13720835
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lxc/lxd v0.0.0-20220210225321-b29334016f17
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.1
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/sys v0.0.0-20220223155357-96fed51e1446
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.8.0
)

replace github.com/renstrom/dedent v1.1.0 => github.com/lithammer/dedent v1.1.0

replace github.com/Sirupsen/logrus v1.7.0 => github.com/sirupsen/logrus v1.7.0
