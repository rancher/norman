module github.com/rancher/norman/v2

go 1.12

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/rancher/wrangler v0.2.1-0.20191015042916-f2a6ecca4f20
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/urfave/cli v1.22.1
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	k8s.io/gengo v0.0.0-20190116091435-f8a0810f38af
	k8s.io/klog v0.3.1
)
