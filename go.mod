module github.com/rancher/norman

go 1.12

replace github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009

require (
	github.com/ghodss/yaml v1.0.0
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/websocket v1.4.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/maruel/panicparse v0.0.0-20171209025017-c0182c169410
	github.com/maruel/ut v1.0.0 // indirect
	github.com/matryer/moq v0.0.0-20190312154309-6cfb0558e1bd
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/rancher/wrangler v0.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/appengine v1.6.1 // indirect
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/gengo v0.0.0-20191120174120-e74f70b9b27e
)
