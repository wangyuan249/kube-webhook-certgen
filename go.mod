module github.com/jet/kube-webhook-certgen

go 1.15

require (
	github.com/oam-dev/kubevela-core-api v1.1.0-rc.1.0.20210810095328-9427af8e26c1
	github.com/onrik/logrus v0.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.5
	sigs.k8s.io/yaml v1.2.0

)

replace (
	k8s.io/api => k8s.io/api v0.21.3
	k8s.io/client-go => k8s.io/client-go v0.21.3
)
