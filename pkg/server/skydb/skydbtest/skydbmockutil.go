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

package skydbtest

import (
	"github.com/golang/mock/gomock"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
)

func ExpectDBSaveUser(db *mock_skydb.MockTxDatabase, extendedSchema *skydb.RecordSchema, assertSavedUserRecord interface{}) {
	db.EXPECT().ID().Return("_public").AnyTimes()

	// no record found
	db.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(skydb.ErrRecordNotFound).
		AnyTimes()

	// extend Schema
	if extendedSchema != nil {
		ExpectDBExtendSchema(db, *extendedSchema)
	}

	db.EXPECT().
		Save(gomock.Any()).
		Do(assertSavedUserRecord).
		Return(nil).
		AnyTimes()
}

func ExpectDBExtendSchema(db *mock_skydb.MockTxDatabase, extendedSchema skydb.RecordSchema) {
	db.EXPECT().UserRecordType().Return("user").AnyTimes()
	db.EXPECT().GetSchema("user").Return(skydb.RecordSchema{}, nil).AnyTimes()
	db.EXPECT().Extend("user", extendedSchema).Return(len(extendedSchema) > 0, nil).AnyTimes()
}
