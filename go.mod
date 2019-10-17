module github.com/skygeario/skygear-server

go 1.13

require (
	cloud.google.com/go/storage v1.1.0
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest/adal v0.6.0 // indirect
	github.com/FZambia/sentinel v1.1.0
	github.com/Masterminds/squirrel v1.1.0
	github.com/aws/aws-sdk-go v1.25.6
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc
	github.com/davidbyttow/govips v0.0.0-20190304175058-d272f04c0fea
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/evalphobia/logrus_fluent v0.4.0
	github.com/fluent/fluent-logger-golang v1.3.0 // indirect
	github.com/go-gomail/gomail v0.0.0-20150902115704-41f357289737
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/golang/mock v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/schema v1.0.2 // indirect
	github.com/h2non/gock v1.0.12
	github.com/iawaknahc/originmatcher v0.0.0-20190816101335-7c3f833688c0
	github.com/jmoiron/sqlx v0.0.0-20170430194603-d9bd385d68c0
	github.com/joho/godotenv v0.0.0-20150907010228-4ed13390c0ac
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/lestrrat-go/jwx v0.9.0
	github.com/lib/pq v0.0.0-20171113044440-8c6ee72f3e6b
	github.com/louischan-oursky/gojsonschema v1.1.1-0.20190618084317-891d0e852428
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/mitchellh/gox v1.0.1
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32 // indirect
	github.com/nbutton23/zxcvbn-go v0.0.0-20171102151520-eafdab6b0663
	github.com/njern/gonexmo v2.0.0+incompatible
	github.com/nyaruka/phonenumbers v1.0.45
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/pquerna/otp v1.2.0
	github.com/rifflock/lfshook v0.0.0-20171219153109-1fdc019a3514
	github.com/sfreiberg/gotwilio v0.0.0-20181012193634-a13e5b0d458a
	github.com/sirupsen/logrus v1.0.3
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a
	github.com/tinylib/msgp v1.1.0
	github.com/ttacon/chalk v0.0.0-20160626202418-22c06c80ed31 // indirect
	github.com/ua-parser/uap-go v0.0.0-20190826212731-daf92ba38329
	github.com/xeipuuv/gojsonpointer v0.0.0-20190809123943-df4f5c81cb3b // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/tools v0.0.0-20190917162342-3b4f30a44f3b
	google.golang.org/api v0.9.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/davidbyttow/govips => github.com/skygeario/govips v0.0.0-20191017114550-b7b5b9596467
