module github.com/authgear/authgear-server

go 1.20

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Masterminds/squirrel v1.5.4
	github.com/abadojack/whatlanggo v1.0.1
	github.com/authgear/graphql-go-relay v0.0.0-20201016065100-df672205b892
	github.com/boombuler/barcode v1.0.1
	// We do not actually use btcd, but it is a dependency of go-ethereum.
	// But if we do not add this, we will run into dependency resolution issue.
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.2
	// https://github.com/elastic/go-elasticsearch#compatibility
	// The client should have equal or less minor version.
	github.com/elastic/go-elasticsearch/v7 v7.16.0
	github.com/felixge/httpsnoop v1.0.3
	github.com/getsentry/sentry-go v0.21.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redsync/redsync/v4 v4.8.1
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/csrf v1.7.1
	github.com/graphql-go/graphql v0.8.1
	github.com/graphql-go/handler v0.2.3
	github.com/iawaknahc/gomessageformat v0.0.0-20210428033148-c3f8592094b5
	github.com/iawaknahc/jsonschema v0.0.0-20211026064614-d05c07b7760d
	github.com/iawaknahc/originmatcher v0.0.0-20221013032908-274b9eda03bc
	github.com/jetstack/cert-manager v1.5.5
	github.com/jmoiron/sqlx v1.3.5
	github.com/joho/godotenv v1.5.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lestrrat-go/jwx v1.2.26
	github.com/lib/pq v1.10.9
	github.com/lithdew/quickjs v0.0.0-20200714182134-aaa42285c9d2
	github.com/njern/gonexmo v2.0.0+incompatible
	github.com/nyaruka/phonenumbers v1.1.7
	github.com/oschwald/geoip2-golang v1.8.0
	github.com/pquerna/otp v1.4.0
	github.com/rubenv/sql-migrate v1.4.0
	github.com/sirupsen/logrus v1.9.2
	github.com/skygeario/go-confusable-homoglyphs v0.0.0-20191212061114-e2b2a60df110
	github.com/smartystreets/goconvey v1.8.0
	github.com/spf13/afero v1.9.5
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.15.0
	github.com/trustelem/zxcvbn v1.0.1
	github.com/ua-parser/uap-go v0.0.0-20211112212520-00c877edfe0f
	golang.org/x/crypto v0.9.0
	golang.org/x/net v0.10.0
	golang.org/x/oauth2 v0.8.0
	golang.org/x/text v0.9.0
	google.golang.org/api v0.123.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/h2non/gock.v1 v1.1.2
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.15
	k8s.io/apimachinery v0.22.15
	k8s.io/client-go v0.22.15
	nhooyr.io/websocket v1.8.7
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/storage v1.30.1
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/aws/aws-sdk-go v1.44.267
	github.com/davidbyttow/govips/v2 v2.13.0
	github.com/duo-labs/webauthn v0.0.0-20220815211337-00c9fb5711f5
	github.com/ethereum/go-ethereum v1.11.6
	github.com/goccy/go-json v0.10.2
	github.com/spruceid/siwe-go v0.2.0
	github.com/stripe/stripe-go/v72 v72.122.0
	github.com/tdewolff/parse/v2 v2.6.6
	github.com/vimeo/go-magic v1.0.0
	golang.org/x/term v0.8.0
)

require (
	github.com/alicebob/miniredis/v2 v2.30.2
	github.com/twilio/twilio-go v1.7.1
	github.com/yudai/gojsondiff v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/compute v1.19.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/cfssl v1.6.1 // indirect
	github.com/cncf/udpa/go v0.0.0-20220112060539-c52dc94e7fbe // indirect
	github.com/cncf/xds/go v0.0.0-20230105202645-06c439db220b // indirect
	github.com/cockroachdb/apd/v2 v2.0.2 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/envoyproxy/go-control-plane v0.10.3 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.9.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fullstorydev/grpcurl v1.8.1 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/certificate-transparency-go v1.1.2-0.20210511102531-373a877eec92 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.3 // indirect
	github.com/google/subcommands v1.0.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.8.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/h2non/parth v0.0.0-20190131123155-b4df798d6542 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/holiman/uint256 v1.2.2-0.20230321075855-87b91420868c // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jhump/protoreflect v1.8.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/oschwald/maxminddb-golang v1.10.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/relvacode/iso8601 v1.1.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/smartystreets/assertions v1.13.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/test-go/testify v1.1.4 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/urfave/cli v1.22.5 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd/api/v3 v3.5.6 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.6 // indirect
	go.etcd.io/etcd/client/v2 v2.305.6 // indirect
	go.etcd.io/etcd/client/v3 v3.5.6 // indirect
	go.etcd.io/etcd/etcdctl/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/server/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/tests/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/v3 v3.5.0-alpha.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/image v0.5.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	k8s.io/apiextensions-apiserver v0.21.3 // indirect
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace github.com/duo-labs/webauthn v0.0.0-20220815211337-00c9fb5711f5 => github.com/authgear/webauthn v0.0.0-20220906031926-bba627c5c32a
