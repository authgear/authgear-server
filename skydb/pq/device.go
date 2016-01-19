package pq

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oursky/skygear/skydb"
)

func (c *conn) GetDevice(id string, device *skydb.Device) error {
	builder := psql.Select("type", "token", "user_id", "last_registered_at").
		From(c.tableName("_device")).
		Where("id = ?", id)

	var nullToken sql.NullString
	err := c.QueryRowWith(builder).Scan(
		&device.Type,
		&nullToken,
		&device.UserInfoID,
		&device.LastRegisteredAt,
	)

	if err == sql.ErrNoRows {
		return skydb.ErrDeviceNotFound
	} else if err != nil {
		return err
	}

	device.Token = nullToken.String

	device.LastRegisteredAt = device.LastRegisteredAt.In(time.UTC)
	device.ID = id

	return nil
}

func (c *conn) QueryDevicesByUser(user string) ([]skydb.Device, error) {
	builder := psql.Select("id", "type", "token", "user_id", "last_registered_at").
		From(c.tableName("_device")).
		Where("user_id = ?", user)

	rows, err := c.QueryWith(builder)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	results := []skydb.Device{}
	for rows.Next() {
		d := skydb.Device{}
		if err := rows.Scan(
			&d.ID,
			&d.Type,
			&d.Token,
			&d.UserInfoID,
			&d.LastRegisteredAt); err != nil {

			panic(err)
		}
		d.LastRegisteredAt = d.LastRegisteredAt.UTC()
		results = append(results, d)
	}
	if err == sql.ErrNoRows {
		return nil, nil
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
		"user_id":            device.UserInfoID,
		"last_registered_at": device.LastRegisteredAt.UTC(),
	}

	if device.Token != "" {
		data["token"] = device.Token
	}

	upsert := upsertQuery(c.tableName("_device"), pkData, data)
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

func (c *conn) DeleteDeviceByToken(token string, t time.Time) error {
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
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
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
