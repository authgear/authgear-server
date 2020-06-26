module github.com/skygeario/skygear-server

go 1.13

require (
	github.com/FZambia/sentinel v1.1.0
	github.com/Masterminds/squirrel v1.1.0
	github.com/getsentry/sentry-go v0.3.0
	github.com/go-gomail/gomail v0.0.0-20150902115704-41f357289737
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/golang/mock v1.4.0
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.4.0
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/mux v1.7.4
	github.com/h2non/gock v1.0.12
	github.com/iawaknahc/gomessageformat v0.0.0-20200406084228-8abc010113fa
	github.com/iawaknahc/jsonschema v0.0.0-20200321082404-507b9d186df7
	github.com/iawaknahc/originmatcher v0.0.0-20200622040912-c5bfd3560192
	github.com/jmoiron/sqlx v0.0.0-20170430194603-d9bd385d68c0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lestrrat-go/jwx v1.0.2
	github.com/lib/pq v1.3.0
	github.com/mattn/go-sqlite3 v1.12.0 // indirect
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32 // indirect
	github.com/nbutton23/zxcvbn-go v0.0.0-20171102151520-eafdab6b0663
	github.com/njern/gonexmo v2.0.0+incompatible
	github.com/nyaruka/phonenumbers v1.0.45
	github.com/pquerna/otp v1.2.0
	github.com/sfreiberg/gotwilio v0.0.0-20181012193634-a13e5b0d458a
	github.com/sirupsen/logrus v1.4.2
	github.com/skygeario/go-confusable-homoglyphs v0.0.0-20191212061114-e2b2a60df110
	github.com/skygeario/openapi3-gen v0.0.0-20190808034633-90418c3d9171
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/ua-parser/uap-go v0.0.0-20190826212731-daf92ba38329
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect
	golang.org/x/text v0.3.2
	golang.org/x/tools v0.0.0-20200417140056-c07e33ef3290
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.2.8
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/xeipuuv/gojsonschema => github.com/skygeario/gojsonschema v1.2.1-0.20200107025531-9fad5cb886b4
