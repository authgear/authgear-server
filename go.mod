module github.com/authgear/authgear-server

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Masterminds/squirrel v1.4.0
	github.com/authgear/graphql-go-relay v0.0.0-20201016065100-df672205b892
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/getsentry/sentry-go v0.6.1
	github.com/go-redis/redis/v8 v8.4.2
	github.com/golang/mock v1.4.4
	github.com/gomodule/redigo v1.8.3
	github.com/google/uuid v1.1.2
	github.com/google/wire v0.4.0
	github.com/gorilla/csrf v1.7.0
	github.com/graphql-go/graphql v0.7.9
	github.com/graphql-go/handler v0.2.3
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/iawaknahc/gomessageformat v0.0.0-20200918074610-c0b982376e20
	github.com/iawaknahc/jsonschema v0.0.0-20201115095512-87990d0baba1
	github.com/iawaknahc/originmatcher v0.0.0-20200622040912-c5bfd3560192
	github.com/jetstack/cert-manager v1.1.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/joho/godotenv v1.3.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lestrrat-go/jwx v1.0.5
	github.com/lib/pq v1.9.0
	github.com/lithdew/quickjs v0.0.0-20200714182134-aaa42285c9d2
	github.com/njern/gonexmo v2.0.0+incompatible
	github.com/nyaruka/phonenumbers v1.0.60
	github.com/pquerna/otp v1.3.0
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/sfreiberg/gotwilio v0.0.0-20200916182813-169c4cd5c691
	github.com/sirupsen/logrus v1.7.0
	github.com/skygeario/go-confusable-homoglyphs v0.0.0-20191212061114-e2b2a60df110
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/afero v1.5.0
	github.com/spf13/cobra v1.1.1
	github.com/test-go/testify v1.1.4 // indirect
	github.com/trustelem/zxcvbn v1.0.1
	github.com/ua-parser/uap-go v0.0.0-20200325213135-e1c09f13e2fe
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	golang.org/x/net v0.0.0-20201207224615-747e23833adb
	golang.org/x/text v0.3.4
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	// gock v1.0.16 is buggy. See https://github.com/h2non/gock/issues/77
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	nhooyr.io/websocket v1.8.6
	sigs.k8s.io/yaml v1.2.0
)
