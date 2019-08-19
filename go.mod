module github.com/cermakm/argo-await-operator

go 1.12

require (
	github.com/argoproj/argo v2.3.0+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/tidwall/gjson v1.3.2
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-rc.0
	sigs.k8s.io/kustomize/v3 v3.1.0 // indirect
)
