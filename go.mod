module github.com/rancher/norman

go 1.12

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/rancher/wrangler v0.1.4
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver v0.0.0-20190313205120-8b27c41bdbb1
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a
)
