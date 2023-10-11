module main

go 1.19

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/net v0.17.0
	k8s.io/api v0.27.3
	k8s.io/kubernetes v1.15.0-alpha.0
)

require github.com/wI2L/jsondiff v0.4.0 // indirect

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/salrashid123/k82_webhook_helloworld/server/webhook/admission v0.0.0
	github.com/salrashid123/k82_webhook_helloworld/server/webhook/mutation v0.0.0 // indirect
	github.com/salrashid123/k82_webhook_helloworld/server/webhook/validation v0.0.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.27.3 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/utils v0.0.0-20230209194617-a36077c30491 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	github.com/salrashid123/k82_webhook_helloworld/server/webhook/admission => ./webhook/admission
	github.com/salrashid123/k82_webhook_helloworld/server/webhook/mutation => ./webhook/mutation
	github.com/salrashid123/k82_webhook_helloworld/server/webhook/validation => ./webhook/validation
)
