// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type CORSMiddleware struct {
}

func (cors CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod := r.Method
		corsMethod := r.Header.Get("Access-Control-Request-Method")
		corsHeaders := r.Header.Get("Access-Control-Request-Headers")

		tConfig := config.GetTenantConfig(r)
		w.Header().Set("Access-Control-Allow-Origin", tConfig.CORSHost)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if corsMethod != "" {
			w.Header().Set("Access-Control-Allow-Methods", corsMethod)
		}

		if corsHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
		}

		if requestMethod == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte{})
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
