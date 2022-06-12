module github.com/pgillich/kubeproxy-ext

go 1.17

require (
	github.com/bombsimon/logrusr/v3 v3.0.0
	github.com/go-logr/logr v1.2.3
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.12.0
	k8s.io/apimachinery v0.21.14-rc.0
	k8s.io/kubernetes v1.21.13
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.21.13 // indirect
	k8s.io/apiserver v0.21.13 // indirect
	k8s.io/client-go v0.21.13 // indirect
	k8s.io/component-base v0.21.13 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20211110012726-3cc51fd1e909 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

require (
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.2
	github.com/subosito/gotenv v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace k8s.io/api => k8s.io/api v0.21.13

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.13

replace k8s.io/apimachinery => k8s.io/apimachinery v0.21.14-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.21.13

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.13

replace k8s.io/client-go => k8s.io/client-go v0.21.13

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.13

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.13

replace k8s.io/code-generator => k8s.io/code-generator v0.21.14-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.21.13

replace k8s.io/component-helpers => k8s.io/component-helpers v0.21.13

replace k8s.io/controller-manager => k8s.io/controller-manager v0.21.13

replace k8s.io/cri-api => k8s.io/cri-api v0.21.14-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.13

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.13

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.13

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.13

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.13

replace k8s.io/kubectl => k8s.io/kubectl v0.21.13

replace k8s.io/kubelet => k8s.io/kubelet v0.21.13

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.13

replace k8s.io/metrics => k8s.io/metrics v0.21.13

replace k8s.io/mount-utils => k8s.io/mount-utils v0.21.14-rc.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.13

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.13

replace k8s.io/sample-controller => k8s.io/sample-controller v0.21.13
