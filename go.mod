module github.com/rancher/norman

go 1.12

replace github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.1
	github.com/gorilla/websocket v0.0.0-20150714140627-6eb6ad425a89
	github.com/kr/pretty v0.1.0 // indirect
	github.com/maruel/panicparse v0.0.0-20171209025017-c0182c169410
	github.com/maruel/ut v1.0.0 // indirect
	github.com/matryer/moq v0.0.0-20190312154309-6cfb0558e1bd
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/rancher/wrangler v0.1.5
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.1 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/appengine v1.6.1 // indirect
	k8s.io/api v0.0.0-20190805182251-6c9aa3caf3d6
	k8s.io/apiextensions-apiserver v0.0.0-20190805184801-2defa3e98ef1
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190805182715-88a2adca7e76+incompatible
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a
)
