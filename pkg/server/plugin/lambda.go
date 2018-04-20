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

package plugin

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type LambdaHandler struct {
	Plugin            *Plugin
	Name              string
	AccessKeyRequired bool
	UserRequired      bool

	AssetStore asset.Store `inject:"AssetStore"`

	Authenticator         router.Processor `preprocessor:"authenticator"`
	InjectIDAuthenticator router.Processor `preprocessor:"inject_auth_id"`
	DBConn                router.Processor `preprocessor:"dbconn"`
	InjectAuth            router.Processor `preprocessor:"require_auth"`
	CheckUser             router.Processor `preprocessor:"check_user"`
	PluginReady           router.Processor `preprocessor:"plugin_ready"`
	preprocessors         []router.Processor
}

func NewLambdaHandler(info map[string]interface{}, p *Plugin) *LambdaHandler {
	handler := &LambdaHandler{
		Plugin: p,
		Name:   info["name"].(string),
	}
	handler.AccessKeyRequired, _ = info["key_required"].(bool)
	handler.UserRequired, _ = info["user_required"].(bool)
	return handler
}

func (h *LambdaHandler) Setup() {
	if h.UserRequired {
		h.preprocessors = []router.Processor{
			h.Authenticator,
			h.DBConn,
			h.InjectAuth,
			h.CheckUser,
			h.PluginReady,
		}
	} else if h.AccessKeyRequired {
		h.preprocessors = []router.Processor{
			h.Authenticator,
			h.DBConn,
			h.PluginReady,
		}
	} else {
		h.preprocessors = []router.Processor{
			h.InjectIDAuthenticator,
			h.DBConn,
			h.PluginReady,
		}
	}
}

func (h *LambdaHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle executes lambda function implemented by the plugin.
func (h *LambdaHandler) Handle(payload *router.Payload, response *router.Response) {
	inbytes, err := h.marshalToPlugin(payload)
	if err != nil {
		response.Err = err
		return
	}

	outbytes, transportErr := h.Plugin.transport.RunLambda(payload.Context, h.Name, inbytes)
	if transportErr != nil {
		switch e := transportErr.(type) {
		case skyerr.Error:
			response.Err = e
		case error:
			response.Err = skyerr.MakeError(transportErr)
		}
		return
	}

	out, err := h.unmarshalFromPlugin(payload, outbytes)
	if err != nil {
		response.Err = err
		return
	}

	log.WithFields(logrus.Fields{
		"name": h.Name,
		"err":  err,
	}).Debugf("Executed a lambda with result")

	response.Result = skyconv.ToLiteral(out)
}

func (h *LambdaHandler) marshalToPlugin(payload *router.Payload) ([]byte, skyerr.Error) {
	in, err := skyconv.TryParseLiteral(payload.Data)
	if err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, err.Error())
	}

	var urlSigner asset.URLSigner
	if signer, ok := h.AssetStore.(asset.URLSigner); ok {
		urlSigner = signer
	}

	in, err = completeAssets(payload.DBConn, urlSigner, in)
	if err != nil {
		return nil, skyerr.MakeError(err)
	}

	inbytes, err := json.Marshal(skyconv.ToLiteral(in))
	if err != nil {
		return nil, skyerr.MakeError(err)
	}
	return inbytes, nil
}

func (h *LambdaHandler) unmarshalFromPlugin(payload *router.Payload, outbytes []byte) (interface{}, skyerr.Error) {
	var outjson interface{}
	err := json.Unmarshal(outbytes, &outjson)
	if err != nil {
		return nil, skyerr.MakeError(err)
	}

	out, err := skyconv.TryParseLiteral(outjson)
	if err != nil {
		return nil, skyerr.MakeError(err)
	}

	var urlSigner asset.URLSigner
	if signer, ok := h.AssetStore.(asset.URLSigner); ok {
		urlSigner = signer
	}

	out, err = completeAssets(payload.DBConn, urlSigner, out)
	if err != nil {
		return nil, skyerr.MakeError(err)
	}
	return out, nil
}

func completeAssets(conn skydb.Conn, urlSigner asset.URLSigner, tree interface{}) (interface{}, error) {
	result := tree
	assetNames := []string{}
	result = walkAssets(result, func(asset *skydb.Asset) *skydb.Asset {
		assetNames = append(assetNames, asset.Name)
		return asset
	})

	assets, err := conn.GetAssets(assetNames)
	if err != nil {
		return nil, err
	}

	assetsByName := map[string]skydb.Asset{}
	for _, asset := range assets {
		assetsByName[asset.Name] = asset
	}

	result = walkAssets(result, func(asset *skydb.Asset) *skydb.Asset {
		completedAsset, ok := assetsByName[asset.Name]
		if !ok {
			asset.Signer = urlSigner
			return asset
		}
		completedAsset.Signer = urlSigner
		return &completedAsset
	})
	return result, nil
}

func walkAssets(tree interface{}, fn func(*skydb.Asset) *skydb.Asset) interface{} {
	switch node := tree.(type) {
	case []interface{}:
		for k, v := range node {
			node[k] = walkAssets(v, fn)
		}
		return node
	case map[string]interface{}:
		for k, v := range node {
			node[k] = walkAssets(v, fn)
		}
		return node
	case skydb.Record:
		walkAssets(map[string]interface{}(node.Data), fn)
		walkAssets(map[string]interface{}(node.Transient), fn)
		return node
	case *skydb.Record:
		walkAssets(map[string]interface{}(node.Data), fn)
		walkAssets(map[string]interface{}(node.Transient), fn)
		return node
	case skydb.Asset:
		result := fn(&node)
		if result == nil {
			result = &node
		}
		return *result
	case *skydb.Asset:
		result := fn(node)
		if result == nil {
			result = node
		}
		return result
	}
	return tree
}
