module github.com/vmware-tanzu-labs/operator-builder

go 1.15

require (
	github.com/gertd/go-pluralize v0.1.7
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/vmware-tanzu-labs/object-code-generator-for-k8s v0.2.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.22.0-alpha.0 // indirect
	sigs.k8s.io/kubebuilder/v3 v3.0.0
	sigs.k8s.io/yaml v1.2.0
)
