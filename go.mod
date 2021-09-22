module github.com/zanloy/bms-api

go 1.16

require (
	github.com/elgs/gojq v0.0.0-20201120033525-b5293fef2759
	github.com/elgs/gosplitargs v0.0.0-20161028071935-a491c5eeb3c8 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-contrib/logger v0.0.2
	github.com/gin-gonic/gin v1.6.3
	github.com/go-resty/resty/v2 v2.5.0
	github.com/gobwas/ws v1.1.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jarcoal/httpmock v1.0.8
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/rs/zerolog v1.20.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/vmware-tanzu/velero v1.5.3
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/net v0.0.0-20210330142815-c8897c278d10 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/metrics v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
)

// Lock our kubernetes version(s)
replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
	k8s.io/kubernetes => k8s.io/kubernetes v0.20.2
	k8s.io/metrics => k8s.io/metrics v0.20.2
)
