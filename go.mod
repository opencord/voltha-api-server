module github.com/opencord/voltha-api-server

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/onsi/ginkgo v1.10.2 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/opencord/voltha-lib-go/v2 v2.2.16
	github.com/opencord/voltha-protos/v2 v2.0.1
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20190930134127-c5a3c61f89f3
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/grpc v1.24.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery v0.15.7
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab // pseudo version corresponding to v12.0.0
)
