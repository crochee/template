module template

go 1.17

require (
	github.com/0x4b53/amqp-rpc/v3 v3.1.0
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/bsm/redislock v0.7.0
	github.com/cenkalti/backoff/v4 v4.1.3
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/gin-gonic/gin v1.7.7
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.11.0
	github.com/go-redis/redis/v8 v8.8.0
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.6.0
	github.com/json-iterator/go v1.1.12
	github.com/mailgun/ttlmap v0.0.0-20170619185759-c1c17f74874f // No tag on the repo.
	github.com/mattn/go-colorable v0.1.6
	github.com/mitchellh/go-homedir v1.1.0
	github.com/orcaman/concurrent-map v1.0.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose v2.7.0+incompatible
	github.com/prometheus/client_golang v1.16.0
	github.com/prometheus/client_model v0.3.0
	github.com/rabbitmq/amqp091-go v1.7.0
	github.com/rs/cors v1.10.1
	github.com/rs/zerolog v1.15.0
	github.com/satori/go.uuid v1.2.0
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/afero v1.1.2
	github.com/spf13/viper v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.14.4
	github.com/vulcand/oxy v1.4.2
	github.com/xuri/excelize/v2 v2.6.1
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/sdk v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
	go.uber.org/automaxprocs v1.5.1
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.21.0
	golang.org/x/time v0.6.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gorm.io/datatypes v1.0.7
	gorm.io/driver/mysql v1.4.4
	gorm.io/gorm v1.23.8
	moul.io/http2curl v1.0.0
)

replace github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gravitational/trace v1.1.16-0.20220114165159-14a9a7dd6aaf // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.4.9
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.4 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mailgun/minheap v0.0.0-00010101000000-000000000000 // indirect
	github.com/mailgun/timetools v0.0.0-20141028012446-7e6055773c51 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/vulcand/predicate v1.2.0 // indirect
	github.com/xuri/efp v0.0.0-20220603152613-6918739fd470 // indirect
	github.com/xuri/nfp v0.0.0-20220409054826-5e722a1d9e22 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/otel/internal/metric v0.27.0 // indirect
	go.opentelemetry.io/otel/metric v0.27.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.13.0
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sync v0.2.0
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/term v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
