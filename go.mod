module github.com/kubepreset/kubepreset

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/imdario/mergo v0.3.12
	github.com/kubepreset/custompod v0.0.0-20210620012502-9c6c8fc38a65
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	go.uber.org/zap v1.17.0
	golang.org/x/sys v0.0.0-20210611083646-a4fc73990273 // indirect
	golang.org/x/tools v0.1.3 // indirect
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	sigs.k8s.io/controller-runtime v0.9.0
)
