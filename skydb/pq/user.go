package pq

import (
	"database/sql"
	"fmt"

	log "github.com/Sirupsen/logrus"
	sq "github.com/lann/squirrel"
	"github.com/oursky/skygear/skydb"
)

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
