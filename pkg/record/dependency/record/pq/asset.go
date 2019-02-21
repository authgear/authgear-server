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

package pq

import (
	"errors"

	sq "github.com/lann/squirrel"

	"github.com/skygeario/skygear-server/pkg/core/skydb/pq/builder"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

func (s *recordStore) GetAsset(name string, asset *record.Asset) error {
	assets, err := s.GetAssets([]string{name})

	if len(assets) == 0 {
		return errors.New("asset not found")
	}

	*asset = assets[0]

	return err
}

func (s *recordStore) GetAssets(names []string) ([]record.Asset, error) {
	if len(names) == 0 {
		return []record.Asset{}, nil
	}

	nameArgs := make([]interface{}, len(names))
	for idx, perName := range names {
		nameArgs[idx] = interface{}(perName)
	}

	builder := s.sqlBuilder.Select("id", "content_type", "size").
		From(s.sqlBuilder.FullTableName("asset")).
		Where("id IN ("+sq.Placeholders(len(names))+")", nameArgs...)

	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []record.Asset{}
	for rows.Next() {
		a := record.Asset{}
		if err := rows.Scan(
			&a.Name,
			&a.ContentType,
			&a.Size); err != nil {

			panic(err)
		}
		results = append(results, a)
	}

	return results, nil
}

func (s *recordStore) SaveAsset(asset *record.Asset) error {
	pkData := map[string]interface{}{
		"id": asset.Name,
	}
	data := map[string]interface{}{
		"content_type": asset.ContentType,
		"size":         asset.Size,
	}
	upsert := builder.UpsertQuery(s.sqlBuilder.FullTableName("asset"), pkData, data)
	_, err := s.sqlExecutor.ExecWith(upsert)
	return err
}
