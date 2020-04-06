module github.com/skygeario/skygear-server

go 1.13

require (
	cloud.google.com/go/storage v1.1.0
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest/adal v0.6.0 // indirect
	github.com/FZambia/sentinel v1.1.0
	github.com/Masterminds/squirrel v1.1.0
	github.com/aws/aws-sdk-go v1.25.6
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc
	github.com/davidbyttow/govips v0.0.0-20190304175058-d272f04c0fea
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/getsentry/sentry-go v0.3.0
	github.com/go-gomail/gomail v0.0.0-20150902115704-41f357289737
	github.com/golang/mock v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.4.0
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/mux v1.7.4
	github.com/h2non/gock v1.0.12
	github.com/iawaknahc/originmatcher v0.0.0-20191203065535-c77f92cc0a75
	github.com/jmoiron/sqlx v0.0.0-20170430194603-d9bd385d68c0
	github.com/joho/godotenv v0.0.0-20150907010228-4ed13390c0ac
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/lestrrat-go/jwx v0.9.0
	github.com/lib/pq v1.2.0
	github.com/mitchellh/gox v1.0.1
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32 // indirect
	github.com/nbutton23/zxcvbn-go v0.0.0-20171102151520-eafdab6b0663
	github.com/njern/gonexmo v2.0.0+incompatible
	github.com/nyaruka/phonenumbers v1.0.45
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pquerna/otp v1.2.0
	github.com/sfreiberg/gotwilio v0.0.0-20181012193634-a13e5b0d458a
	github.com/sirupsen/logrus v1.4.2
	github.com/skygeario/go-confusable-homoglyphs v0.0.0-20191212061114-e2b2a60df110
	github.com/skygeario/skygear-server/migrate v0.0.0-20200403033656-d4cdc710bca2 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337
	github.com/tinylib/msgp v1.1.0
	github.com/ttacon/chalk v0.0.0-20160626202418-22c06c80ed31 // indirect
	github.com/ua-parser/uap-go v0.0.0-20190826212731-daf92ba38329
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/text v0.3.2
	golang.org/x/tools v0.0.0-20190917162342-3b4f30a44f3b
	google.golang.org/api v0.9.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.2.7
)

// The reason why we have to use a fork is to support background color in the embed() operation.
// See https://github.com/SkygearIO/govips/commit/b7b5b9596467e8b6b5f11f2178c754df83e9a35c
replace github.com/davidbyttow/govips => github.com/skygeario/govips v0.0.0-20191017114550-b7b5b9596467

replace github.com/xeipuuv/gojsonschema => github.com/skygeario/gojsonschema v1.2.1-0.20200107025531-9fad5cb886b4

replace gopkg.in/yaml.v2 => github.com/skygeario/go-yaml v0.0.0-20191213113752-45105225b50d
