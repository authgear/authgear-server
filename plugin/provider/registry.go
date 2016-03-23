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

package provider

import (
	"fmt"
	"sync"
)

// Registry contains registered providers, which provide additional
// functionality implemented by plugins.
type Registry struct {
	mutex         sync.RWMutex
	authProviders map[string]AuthProvider
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		mutex:         sync.RWMutex{},
		authProviders: map[string]AuthProvider{},
	}
}

// RegisterAuthProvider registers an AuthProvider with the registry.
func (r *Registry) RegisterAuthProvider(name string, p AuthProvider) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.authProviders[name] = p
}

// GetAuthProvider gets an AuthProvider from the registry.
func (r *Registry) GetAuthProvider(name string) (AuthProvider, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	provider, ok := r.authProviders[name]
	if !ok {
		return nil, fmt.Errorf(`no auth provider of name "%s"`, name)
	}
	return provider, nil
}
