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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq/builder"
)

func (c *conn) GetDevice(id string, device *skydb.Device) error {
	builder := psql.Select("type", "token", "auth_id", "topic", "last_registered_at").
		From(c.tableName("_device")).
		Where("id = ?", id)

	nullableToken := sql.NullString{}
	nullableTopic := sql.NullString{}
	nullableUserID := sql.NullString{}
	err := c.QueryRowWith(builder).Scan(
		&device.Type,
		&nullableToken,
		&nullableUserID,
		&nullableTopic,
		&device.LastRegisteredAt,
	)

	if err == sql.ErrNoRows {
		return skydb.ErrDeviceNotFound
	} else if err != nil {
		return err
	}

	device.Token = nullableToken.String
	device.Topic = nullableTopic.String
	device.AuthInfoID = nullableUserID.String
	device.LastRegisteredAt = device.LastRegisteredAt.In(time.UTC)
	device.ID = id

	return nil
}

func (c *conn) QueryDevicesByUser(user string) ([]skydb.Device, error) {
	builder := psql.Select("id", "type", "token", "auth_id", "topic", "last_registered_at").
		From(c.tableName("_device")).
		Where("auth_id = ?", user)

	rows, err := c.QueryWith(builder)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	results := []skydb.Device{}
	for rows.Next() {
		nullableToken := sql.NullString{}
		nullableTopic := sql.NullString{}
		d := skydb.Device{}
		if err := rows.Scan(
			&d.ID,
			&d.Type,
			&nullableToken,
			&d.AuthInfoID,
			&nullableTopic,
			&d.LastRegisteredAt); err != nil {

			panic(err)
		}
		d.Token = nullableToken.String
		d.Topic = nullableTopic.String
		d.LastRegisteredAt = d.LastRegisteredAt.UTC()
		results = append(results, d)
	}

	return results, nil
}

func (c *conn) QueryDevicesByUserAndTopic(user, topic string) ([]skydb.Device, error) {
	builder := psql.Select("id", "type", "token", "auth_id", "topic", "last_registered_at").
		From(c.tableName("_device")).
		Where("auth_id = ? AND topic = ?", user, topic)

	rows, err := c.QueryWith(builder)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	results := []skydb.Device{}
	for rows.Next() {
		var nullableToken sql.NullString
		d := skydb.Device{}
		if err := rows.Scan(
			&d.ID,
			&d.Type,
			&nullableToken,
			&d.AuthInfoID,
			&d.Topic,
			&d.LastRegisteredAt); err != nil {

			panic(err)
		}
		d.Token = nullableToken.String
		d.LastRegisteredAt = d.LastRegisteredAt.UTC()
		results = append(results, d)
	}

	return results, nil
}

func (c *conn) SaveDevice(device *skydb.Device) error {
	if device.ID == "" || device.Type == "" || device.LastRegisteredAt.IsZero() {
		return errors.New("invalid device: empty id, type, or last registered at")
	}

	pkData := map[string]interface{}{"id": device.ID}
	data := map[string]interface{}{
		"type":               device.Type,
		"auth_id":            nil,
		"last_registered_at": device.LastRegisteredAt.UTC(),
	}

	if device.AuthInfoID != "" {
		data["auth_id"] = device.AuthInfoID
	}

	if device.Token != "" {
		data["token"] = device.Token
	}

	if device.Topic != "" {
		data["topic"] = device.Topic
	}

	upsert := builder.UpsertQuery(c.tableName("_device"), pkData, data)
	_, err := c.ExecWith(upsert)
	return err
}

func (c *conn) DeleteDevice(id string) error {
	builder := psql.Delete(c.tableName("_device")).
		Where("id = ?", id)
	result, err := c.ExecWith(builder)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrDeviceNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) DeleteDevicesByToken(token string, t time.Time) error {
	builder := psql.Delete(c.tableName("_device")).
		Where("token = ?", token)
	if t != skydb.ZeroTime {
		builder = builder.Where("last_registered_at < ?", t)
	}
	result, err := c.ExecWith(builder)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrDeviceNotFound
	}

	return nil
}

func (c *conn) DeleteEmptyDevicesByTime(t time.Time) error {
	builder := psql.Delete(c.tableName("_device")).
		Where("token IS NULL")
	if t != skydb.ZeroTime {
		builder = builder.Where("last_registered_at < ?", t)
	}
	result, err := c.ExecWith(builder)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrDeviceNotFound
	}

	return nil
}
