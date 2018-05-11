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

package router

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type RequestIDMiddleware struct {
	Next http.Handler
}

func (m *RequestIDMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New()
	newContext := context.WithValue(r.Context(), "RequestID", requestID)
	r = r.WithContext(newContext)

	w.Header().Set("X-Skygear-Request-Id", requestID)
	m.Next.ServeHTTP(w, r)
}
