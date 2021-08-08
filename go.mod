module github.com/jet/kube-webhook-certgen

go 1.15

require (
	github.com/lib/pq v1.10.0 // indirect
	github.com/oam-dev/kubevela-core-api v1.1.0-rc.2
	github.com/onrik/logrus v0.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.5
	sigs.k8s.io/yaml v1.2.0
)
