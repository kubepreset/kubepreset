module github.com/kubepreset/kubepreset

go 1.16

require (
	github.com/go-logr/logr v0.3.0
	github.com/imdario/mergo v0.3.10
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.10.2
	go.uber.org/zap v1.15.0
	golang.org/x/sys v0.0.0-20210611083646-a4fc73990273 // indirect
	golang.org/x/tools v0.1.3 // indirect
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
