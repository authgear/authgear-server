#export K6_HTTP_DEBUG=true
# We cannot use K6_DURATION and K6_VUS because
# using them will make k6 to ignore scenarios.
export DURATION=10s
export VUS=10
export ENDPOINT=http://localhost:3100
export CLIENT_ID=test
export REDIRECT_URI=com.example://host
export FIXTURE_FIXED_OTP=000000
export FIXTURE_PASSWORD=a12345678
export FIXTURE_EMAIL_DOMAIN=example.com
export FIXTURE_PHONE_NUMBER_COUNTRY_CALLING_CODE=+852
export FIXTURE_PHONE_NUMBER_LOW=52000000
export FIXTURE_PHONE_NUMBER_HIGH=57999999

.PHONY: signup
signup:
	k6 run --vus $(VUS) --duration $(DURATION) ./signup.js

.PHONY: login
login:
	k6 run --vus $(VUS) --duration $(DURATION) ./login.js

.PHONY: healthz
healthz:
	k6 run --vus $(VUS) --duration $(DURATION) ./healthz.js
