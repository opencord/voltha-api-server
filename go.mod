module github.com/opencord/voltha-api-server

go 1.12

require (
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/opencord/voltha-lib-go/v2 v2.2.16
	github.com/opencord/voltha-protos/v2 v2.0.1
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	google.golang.org/grpc v1.27.0
	k8s.io/api v0.20.0-alpha.2
	k8s.io/apimachinery v0.20.0-alpha.2
	k8s.io/client-go v0.20.0-alpha.2 // pseudo version corresponding to v12.0.0
)
