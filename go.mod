module github.com/jet/kube-webhook-certgen

go 1.15

require (
	github.com/onrik/logrus v0.3.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
