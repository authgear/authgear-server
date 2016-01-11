package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type userInfo struct {
	ID             string        `db:"id"`
	Email          string        `db:"email"`
	HashedPassword []byte        `db:"password"`
	Auth           authInfoValue `db:"auth"`
}

// authInfoValue implements sql.Valuer and sql.Scanner s.t.
// skydb.AuthInfo can be saved into and recovered from postgresql
type authInfoValue skydb.AuthInfo

func (auth authInfoValue) Value() (driver.Value, error) {
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(auth); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (auth *authInfoValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("skydb: unsupported Scan pair: %T -> %T", value, auth)
	}

	return json.Unmarshal(b, auth)
}

// Ext is an interface for both sqlx.DB and sqlx.Tx
type Ext interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
}

type conn struct {
	db             *sqlx.DB // database wrapper
	tx             *sqlx.Tx // transaction wrapper, nil when no transaction
	txDone         bool     // transaction is done, can only have one tx currently
	RecordSchema   map[string]skydb.RecordSchema
	appName        string
	option         string
	statementCount uint64
	accessModel    skydb.AccessModel
}

// Db returns the current database wrapper, or a transaction wrapper when
// a transaction is in effect.
func (c *conn) Db() Ext {
	if c.tx != nil {
		return c.tx
	}
	return c.db
}

// Begin begins a transaction.
func (c *conn) Begin() (err error) {
	if c.txDone {
		return skydb.ErrDatabaseTxDone
	}

	if c.tx != nil {
		return skydb.ErrDatabaseTxDidBegin
	}

	c.tx, err = c.db.Beginx()
	return
}

// Commit commits a transaction.
func (c *conn) Commit() (err error) {
	if c.txDone {
		return skydb.ErrDatabaseTxDone
	}

	if c.tx == nil {
		return skydb.ErrDatabaseTxDidNotBegin
	}

	err = c.tx.Commit()
	if err == nil {
		c.txDone = true
	}
	return
}

// Rollback rollbacks a transaction.
func (c *conn) Rollback() (err error) {
	if c.txDone {
		return skydb.ErrDatabaseTxDone
	}

	if c.tx == nil {
		return skydb.ErrDatabaseTxDidNotBegin
	}

	err = c.tx.Rollback()
	if err == nil {
		c.txDone = true
	}
	return
}

func (c *conn) CreateUser(userinfo *skydb.UserInfo) error {
	var (
		username *string
		email    *string
	)
	if userinfo.Username != "" {
		username = &userinfo.Username
	} else {
		username = nil
	}
	if userinfo.Email != "" {
		email = &userinfo.Email
	} else {
		email = nil
	}

	builder := psql.Insert(c.tableName("_user")).Columns(
		"id",
		"username",
		"email",
		"password",
		"auth",
	).Values(
		userinfo.ID,
		username,
		email,
		userinfo.HashedPassword,
		authInfoValue(userinfo.Auth),
	)

	_, err := c.ExecWith(builder)
	if isUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}

	return err
}

func (c *conn) doScanUser(userinfo *skydb.UserInfo, scanner sq.RowScanner) error {
	var (
		id       string
		username sql.NullString
		email    sql.NullString
	)
	password, auth := []byte{}, authInfoValue{}
	err := scanner.Scan(
		&id,
		&username,
		&email,
		&password,
		&auth,
	)
	if err != nil {
		log.Infof(err.Error())
	}
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	userinfo.ID = id
	userinfo.Username = username.String
	userinfo.Email = email.String
	userinfo.HashedPassword = password
	userinfo.Auth = skydb.AuthInfo(auth)

	return err
}

func (c *conn) GetUser(id string, userinfo *skydb.UserInfo) error {
	log.Warnf(id)
	builder := psql.Select("id", "username", "email", "password", "auth").
		From(c.tableName("_user")).
		Where("id = ?", id)
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(userinfo, scanner)
}

func (c *conn) GetUserByUsernameEmail(username string, email string, userinfo *skydb.UserInfo) error {
	var builder sq.SelectBuilder
	if email == "" {
		builder = psql.Select("id", "username", "email", "password", "auth").
			From(c.tableName("_user")).
			Where("username = ?", username)
	} else if username == "" {
		builder = psql.Select("id", "username", "email", "password", "auth").
			From(c.tableName("_user")).
			Where("email = ?", email)
	} else {
		builder = psql.Select("id", "username", "email", "password", "auth").
			From(c.tableName("_user")).
			Where("username = ? AND email = ?", username, email)
	}
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(userinfo, scanner)
}

func (c *conn) GetUserByPrincipalID(principalID string, userinfo *skydb.UserInfo) error {
	builder := psql.Select("id", "username", "email", "password", "auth").
		From(c.tableName("_user")).
		Where("jsonb_exists(auth, ?)", principalID)
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(userinfo, scanner)
}

func (c *conn) QueryUser(emails []string) ([]skydb.UserInfo, error) {

	emailargs := make([]interface{}, len(emails))
	for i, v := range emails {
		emailargs[i] = interface{}(v)
	}

	builder := psql.Select("id", "username", "email", "password", "auth").
		From(c.tableName("_user")).
		Where("email IN ("+sq.Placeholders(len(emailargs))+") AND email IS NOT NULL AND email != ''", emailargs...)

	rows, err := c.QueryWith(builder)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	results := []skydb.UserInfo{}
	for rows.Next() {
		var (
			id       string
			username sql.NullString
			email    sql.NullString
		)
		password, auth := []byte{}, authInfoValue{}
		if err := rows.Scan(&id, &username, &email, &password, &auth); err != nil {
			panic(err)
		}

		userinfo := skydb.UserInfo{}
		userinfo.ID = id
		userinfo.Username = username.String
		userinfo.Email = email.String
		userinfo.HashedPassword = password
		userinfo.Auth = skydb.AuthInfo(auth)
		results = append(results, userinfo)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return results, nil
}

func (c *conn) UpdateUser(userinfo *skydb.UserInfo) error {
	var (
		username *string
		email    *string
	)
	if userinfo.Username != "" {
		username = &userinfo.Username
	} else {
		username = nil
	}
	if userinfo.Email != "" {
		email = &userinfo.Email
	} else {
		email = nil
	}
	builder := psql.Update(c.tableName("_user")).
		Set("username", username).
		Set("email", email).
		Set("password", userinfo.HashedPassword).
		Set("auth", authInfoValue(userinfo.Auth)).
		Where("id = ?", userinfo.ID)

	result, err := c.ExecWith(builder)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) DeleteUser(id string) error {
	builder := psql.Delete(c.tableName("_user")).
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
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) QueryRelation(user string, name string, direction string, config skydb.QueryConfig) []skydb.UserInfo {
	log.Debugf("Query Relation: %v, %v", user, name)
	var selectBuilder sq.SelectBuilder

	if direction == "outward" {
		selectBuilder = psql.Select("u.id", "u.username", "u.email").
			From(c.tableName("_user")+" AS u").
			Join(c.tableName(name)+" AS relation ON relation.right_id = u.id").
			Where("relation.left_id = ?", user)
	} else if direction == "inward" {
		selectBuilder = psql.Select("u.id", "u.username", "u.email").
			From(c.tableName("_user")+" AS u").
			Join(c.tableName(name)+" AS relation ON relation.left_id = u.id").
			Where("relation.right_id = ?", user)
	} else {
		selectBuilder = psql.Select("u.id", "u.username", "u.email").
			From(c.tableName("_user")+" AS u").
			Join(c.tableName(name)+" AS inward_relation ON inward_relation.left_id = u.id").
			Join(c.tableName(name)+" AS outward_relation ON outward_relation.right_id = u.id").
			Where("inward_relation.right_id = ?", user).
			Where("outward_relation.left_id = ?", user)
	}

	selectBuilder = selectBuilder.OrderBy("u.id").
		Offset(config.Offset)
	if config.Limit != 0 {
		selectBuilder = selectBuilder.Limit(config.Limit)
	}

	rows, err := c.QueryWith(selectBuilder)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	results := []skydb.UserInfo{}
	for rows.Next() {
		var (
			id       string
			username sql.NullString
			email    sql.NullString
		)
		if err := rows.Scan(&id, &username, &email); err != nil {
			panic(err)
		}
		userInfo := skydb.UserInfo{
			ID:       id,
			Username: username.String,
			Email:    email.String,
		}
		results = append(results, userInfo)
	}
	return results
}

func (c *conn) QueryRelationCount(user string, name string, direction string) (uint64, error) {
	log.Debugf("Query Relation Count: %v, %v, %v", user, name, direction)
	query := psql.Select("COUNT(*)").From(c.tableName(name) + "AS _primary")
	if direction == "outward" {
		query = query.Where("_primary.left_id = ?", user)
	} else if direction == "inward" {
		query = query.Where("_primary.right_id = ?", user)
	} else {
		query = query.
			Join(c.tableName(name)+" AS _secondary ON _secondary.left_id = _primary.right_id").
			Where("_primary.left_id = ?", user).
			Where("_secondary.right_id = ?", user)
	}
	var count uint64
	err := c.GetWith(&count, query)
	if err != nil {
		panic(err)
	}
	return count, err
}

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

func (c *conn) AddRelation(user string, name string, targetUser string) error {
	ralationPair := map[string]interface{}{
		"left_id":  user,
		"right_id": targetUser,
	}

	upsert := upsertQuery(c.tableName(name), ralationPair, nil)
	_, err := c.ExecWith(upsert)
	if err != nil {
		if isForienKeyViolated(err) {
			return fmt.Errorf("userID not exist")
		}
	}

	return err
}

func (c *conn) RemoveRelation(user string, name string, targetUser string) error {
	builder := psql.Delete(c.tableName(name)).
		Where("left_id = ? AND right_id = ?", user, targetUser)
	result, err := c.ExecWith(builder)

	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%v relation not exist {%v} => {%v}",
			name, user, targetUser)
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}
	return nil
}

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

func (c *conn) PublicDB() skydb.Database {
	return &database{
		c: c,
	}
}

func (c *conn) PrivateDB(userKey string) skydb.Database {
	return &database{
		c:      c,
		userID: userKey,
	}
}

func (c *conn) Close() error { return nil }

// return the raw unquoted schema name of this app
func (c *conn) schemaName() string {
	return "app_" + toLowerAndUnderscore(c.appName)
}

// return the quoted table name ready to be used as identifier (in the form
// "schema"."table")
func (c *conn) tableName(table string) string {
	return pq.QuoteIdentifier(c.schemaName()) + "." + pq.QuoteIdentifier(table)
}

type database struct {
	c      *conn
	userID string
	txDone bool
}

func (db *database) Conn() skydb.Conn { return db.c }
func (db *database) ID() string       { return "" }

// schemaName is a convenient method to access parent conn's schemaName
func (db *database) schemaName() string {
	return db.c.schemaName()
}

// tableName is a convenient method to access parent conn's tableName
func (db *database) tableName(table string) string {
	return db.c.tableName(table)
}

// this ensures that our structure conform to certain interfaces.
var (
	_ skydb.Conn     = &conn{}
	_ skydb.Database = &database{}

	_ driver.Valuer = authInfoValue{}
)
