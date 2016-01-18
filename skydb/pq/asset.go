package pq

import (
	"database/sql"
	"errors"

	"github.com/oursky/skygear/skydb"
)

func (c *conn) GetAsset(name string, asset *skydb.Asset) error {
	builder := psql.Select("content_type", "size").
		From(c.tableName("_asset")).
		Where("id = ?", name)

	var (
		contentType string
		size        int64
	)
	err := c.QueryRowWith(builder).Scan(
		&contentType,
		&size,
	)
	if err == sql.ErrNoRows {
		return errors.New("asset not found")
	}

	asset.Name = name
	asset.ContentType = contentType
	asset.Size = size

	return err
}

func (c *conn) SaveAsset(asset *skydb.Asset) error {
	pkData := map[string]interface{}{
		"id": asset.Name,
	}
	data := map[string]interface{}{
		"content_type": asset.ContentType,
		"size":         asset.Size,
	}
	upsert := upsertQuery(c.tableName("_asset"), pkData, data)
	_, err := c.ExecWith(upsert)
	return err
}
