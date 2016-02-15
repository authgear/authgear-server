package exec

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	skyplugin "github.com/oursky/skygear/plugin"
	"github.com/oursky/skygear/plugin/common"
	"github.com/oursky/skygear/skyconfig"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRun(t *testing.T) {
	Convey("test args and stdout", t, func() {
		transport := &execTransport{
			Path: "/bin/echo",
			Args: []string{},
		}

		originalCommand := startCommand
		defer func() {
			startCommand = originalCommand
		}()

		Convey("init", func() {
			out, err := transport.RunInit()
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "init")
		})

		startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
			out, err = originalCommand(cmd, in)
			out = append([]byte(`{"result":"`), out...)
			out = append(out, []byte(`"}`)...)
			return
		}

		Convey("op", func() {
			out, err := transport.RunLambda(nil, "hello:world", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"op hello:world"`)
		})

		Convey("handler", func() {
			out, err := transport.RunHandler("hello:world", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"handler hello:world"`)
		})
	})

	Convey("test stdin", t, func() {
		transport := &execTransport{
			Path: "/bin/sh",
			Args: []string{"-c", `"cat"`},
		}

		Convey("init", func() {
			out, err := transport.RunInit()
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "")
		})

		Convey("op", func() {
			out, err := transport.RunLambda(nil, "hello:world", []byte(`{"result": "hello world"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"hello world"`)
		})

		Convey("handler", func() {
			out, err := transport.RunHandler("hello:world", []byte(`{"result": "hello world"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"hello world"`)
		})
	})

	Convey("test lambda", t, func() {
		transport := &execTransport{
			Path: "/never/invoked",
			Args: nil,
		}

		// expect child test case to override startCommand
		// save the original and defer setting it back
		originalCommand := startCommand
		defer func() {
			startCommand = originalCommand
		}()

		Convey("pass context as environment variable", func() {
			found := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				for _, envLine := range cmd.Env {
					envTuple := strings.SplitN(envLine, "=", 2)
					if envTuple[0] == "SKYGEAR_CONTEXT" {
						decodedCtx := map[string]interface{}{}
						err := common.DecodeBase64JSON(envTuple[1], &decodedCtx)
						So(err, ShouldBeNil)
						So(decodedCtx, ShouldResemble, map[string]interface{}{
							"user_id": "user",
						})
						found = true
						break
					}
				}
				return []byte(`{"result": {}}`), nil
			}

			ctx := context.WithValue(context.Background(), "UserID", "user")
			transport.RunLambda(ctx, "work", []byte{})
			So(found, ShouldBeTrue)
		})
	})

	Convey("test hook", t, func() {
		transport := &execTransport{
			Path: "/never/invoked",
			Args: nil,
		}

		// expect child test case to override startCommand
		// save the original and defer setting it back
		originalCommand := startCommand
		defer func() {
			startCommand = originalCommand
		}()

		recordin := skydb.Record{
			ID:      skydb.NewRecordID("note", "id"),
			OwnerID: "john.doe@example.com",
			ACL: skydb.RecordACL{
				skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
				skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
			},
			Data: map[string]interface{}{
				"content":   "some note content",
				"noteOrder": float64(1),
				"tags":      []interface{}{"test", "unimportant"},
				"date":      time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC),
				"ref":       skydb.NewReference("category", "1"),
				"asset":     &skydb.Asset{Name: "asset-name"},
			},
		}

		recordold := skydb.Record{
			ID:      skydb.NewRecordID("note", "id"),
			OwnerID: "john.doe@example.com",
			ACL: skydb.RecordACL{
				skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
				skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
			},
			Data: map[string]interface{}{
				"content":   "original content",
				"noteOrder": float64(1),
				"tags":      []interface{}{},
				"date":      time.Date(2017, 7, 21, 19, 30, 24, 0, time.UTC),
			},
		}

		Convey("executes beforeSave correctly", func() {
			called := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				called = true
				So(cmd.Path, ShouldEqual, "/never/invoked")
				So(cmd.Args, ShouldResemble, []string{"/never/invoked", "hook", "note_beforeSave"})
				So(in, ShouldEqualJSON, `{
					"record": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"content": "some note content",
						"noteOrder": 1,
						"tags": ["test", "unimportant"],
						"date": {
							"$type": "date",
							"$date": "2017-07-23T19:30:24Z"
						},
						"ref": {
							"$type": "ref",
							"$id": "category/1"
						},
						"asset":{
							"$type": "asset",
							"$name": "asset-name"
						},
						"_access": [{
							"relation": "friend",
							"level": "write"
						}, {
							"relation": "$direct",
							"level": "read",
							"user_id": "user_id"
						}]
					},
					"original": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"content": "original content",
						"noteOrder": 1,
						"tags": [],
						"date": {
							"$type": "date",
							"$date": "2017-07-21T19:30:24Z"
						},
						"_access": [{
							"relation": "friend",
							"level": "write"
						}, {
							"relation": "$direct",
							"level": "read",
							"user_id": "user_id"
						}]
					}
				}`)

				return []byte(`{
					"result": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"content": "content has been modified",
						"noteOrder": 1,
						"tags": ["test", "unimportant"],
						"date": {
							"$type": "date",
							"$date": "2017-07-23T19:30:24Z"
						},
						"ref": {
							"$type": "ref",
							"$id": "category/1"
						},
						"asset":{
							"$type": "asset",
							"$name": "asset-name"
						},
						"_access": [{
							"relation": "friend",
							"level": "write"
						}, {
							"relation": "$direct",
							"level": "read",
							"user_id": "user_id"
						}]
					}
				}`), nil
			}

			recordout, err := transport.RunHook(nil, "note_beforeSave", &recordin, &recordold)
			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)

			datein := recordin.Data["date"].(time.Time)
			delete(recordin.Data, "date")
			So(recordin, ShouldResemble, skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "some note content",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			})
			// GoConvey's bug, ShouldEqual and ShouldResemble doesn't work on time.Time
			So(datein == time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC), ShouldBeTrue)

			dateout := recordout.Data["date"].(time.Time)
			delete(recordout.Data, "date")
			So(*recordout, ShouldResemble, skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "content has been modified",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			})
			So(dateout == time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC), ShouldBeTrue)
		})

		Convey("executes beforeSave with original", func() {
			called := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				called = true
				So(cmd.Path, ShouldEqual, "/never/invoked")
				So(cmd.Args, ShouldResemble, []string{"/never/invoked", "hook", "note_beforeSave"})
				So(in, ShouldEqualJSON, `{
					"record": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"content": "some note content",
						"noteOrder": 1,
						"tags": ["test", "unimportant"],
						"date": {
							"$type": "date",
							"$date": "2017-07-23T19:30:24Z"
						},
						"ref": {
							"$type": "ref",
							"$id": "category/1"
						},
						"asset":{
							"$type": "asset",
							"$name": "asset-name"
						},
						"_access": [{
							"relation": "friend",
							"level": "write"
						}, {
							"relation": "$direct",
							"level": "read",
							"user_id": "user_id"
						}]
					},
					"original": null
				}`)

				return []byte(`{
					"result": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"content": "content has been modified",
						"noteOrder": 1,
						"tags": ["test", "unimportant"],
						"date": {
							"$type": "date",
							"$date": "2017-07-23T19:30:24Z"
						},
						"ref": {
							"$type": "ref",
							"$id": "category/1"
						},
						"asset":{
							"$type": "asset",
							"$name": "asset-name"
						},
						"_access": [{
							"relation": "friend",
							"level": "write"
						}, {
							"relation": "$direct",
							"level": "read",
							"user_id": "user_id"
						}]
					}
				}`), nil
			}

			recordout, err := transport.RunHook(nil, "note_beforeSave", &recordin, nil)
			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)

			datein := recordin.Data["date"].(time.Time)
			delete(recordin.Data, "date")
			So(recordin, ShouldResemble, skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "some note content",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			})
			// GoConvey's bug, ShouldEqual and ShouldResemble doesn't work on time.Time
			So(datein == time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC), ShouldBeTrue)

			dateout := recordout.Data["date"].(time.Time)
			delete(recordout.Data, "date")
			So(*recordout, ShouldResemble, skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "content has been modified",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			})
			So(dateout == time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC), ShouldBeTrue)
		})

		Convey("serialize meta data correctly", func() {
			recordin := skydb.Record{
				ID:        skydb.NewRecordID("note", "id"),
				OwnerID:   "john.doe@example.com",
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				CreatorID: "creatorID",
				UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				UpdaterID: "updaterID",
				Data:      map[string]interface{}{},
			}

			called := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				called = true
				So(string(in), ShouldEqualJSON, `{
					"record": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"_ownerID": "john.doe@example.com",
						"_created_at": "2006-01-02T15:04:05Z",
						"_created_by": "creatorID",
						"_updated_at": "2006-01-02T15:04:05Z",
						"_updated_by": "updaterID",
						"_access": null
					},
					"original": null
				}`)
				return []byte(`{
					"result": {
						"_id": "note/id",
						"_ownerID": "john.doe@example.com",
						"_access": null
					}
				}`), nil
			}

			recordout, err := transport.RunHook(nil, "note_beforeSave", &recordin, nil)
			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)
			So(*recordout, ShouldResemble, recordin)

		})

		Convey("parses null ACL correctly", func() {
			recordin := skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL:     nil,
				Data:    map[string]interface{}{},
			}

			called := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				called = true
				So(string(in), ShouldEqualJSON, `{
					"record": {
						"_id": "note/id",
						"_type": "record",
						"_ownerID": "john.doe@example.com",
						"_access": null
					},
					"original": null
				}`)
				return []byte(`{
					"result": {
						"_id": "note/id",
						"_ownerID": "john.doe@example.com",
						"_access": null
					}
				}`), nil
			}

			recordout, err := transport.RunHook(nil, "note_beforeSave", &recordin, nil)
			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)
			So(*recordout, ShouldResemble, recordin)
		})

		Convey("returns err if command failed", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return nil, errors.New("worrying error")
			}

			recordout, err := transport.RunHook(nil, "note_afterSave", &recordin, nil)
			So(err.Error(), ShouldEqual, "worrying error")
			So(recordout, ShouldBeNil)
		})

		Convey("returns err if command returns invalid response", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return []byte("I am not a json"), nil
			}

			recordout, err := transport.RunHook(nil, "note_afterSave", &recordin, nil)
			So(err.Error(), ShouldEqual, "failed to parse response: invalid character 'I' looking for beginning of value")
			So(recordout, ShouldBeNil)
		})

		Convey("returns err if commands returns with inner error", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return []byte(`{
					"result": {
						"ignore": "me"
					},
					"error": {
						"code": 24601,
						"message": "Too strong to lift a feather",
						"info": {}
					}
				}`), nil
			}

			recordout, err := transport.RunHook(nil, "note_afterSave", &recordin, nil)
			sError, ok := err.(skyerr.Error)
			So(ok, ShouldBeTrue)
			So(sError.Message(), ShouldEqual, `Too strong to lift a feather`)
			So(sError.Code(), ShouldEqual, 24601)
			So(recordout, ShouldBeNil)
		})

		Convey("pass context as environment variable", func() {
			found := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				for _, envLine := range cmd.Env {
					envTuple := strings.SplitN(envLine, "=", 2)
					if envTuple[0] == "SKYGEAR_CONTEXT" {
						decodedCtx := map[string]interface{}{}
						err := common.DecodeBase64JSON(envTuple[1], &decodedCtx)
						So(err, ShouldBeNil)
						So(decodedCtx, ShouldResemble, map[string]interface{}{
							"user_id": "user",
						})
						found = true
						break
					}
				}
				return []byte(`{"result": {}}`), nil
			}

			ctx := context.WithValue(context.Background(), "UserID", "user")
			transport.RunHook(ctx, "note_afterSave", &recordin, nil)
			So(found, ShouldBeTrue)
		})
	})

	Convey("test exec error", t, func() {
		Convey("file not found", func() {
			transport := &execTransport{
				Path: "/tmp/nonexistent",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})

		Convey("not executable", func() {
			transport := &execTransport{
				Path: "/dev/null",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})

		Convey("return false", func() {
			transport := &execTransport{
				Path: "/bin/false",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})
	})

	Convey("test provider", t, func() {
		transport := &execTransport{
			Path: "/never/invoked",
			Args: nil,
		}

		// expect child test case to override startCommand
		// save the original and defer setting it back
		originalCommand := startCommand
		defer func() {
			startCommand = originalCommand
		}()

		Convey("executes provider passing auth data", func() {
			called := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				called = true
				So(cmd.Path, ShouldEqual, "/never/invoked")
				So(cmd.Args, ShouldResemble, []string{"/never/invoked", "provider", "com.example", "login"})
				So(in, ShouldEqualJSON, `{
					"auth_data": {"password": "secret"}
				}`)

				return []byte(`{
					"result": {
						"principal_id": "johndoe",
						"auth_data": {"token": "A_TOKEN"}
					}
				}`), nil
			}

			authData := map[string]interface{}{
				"password": "secret",
			}
			req := skyplugin.AuthRequest{"com.example", "login", authData}

			resp, err := transport.RunProvider(&req)
			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)
			So(resp.PrincipalID, ShouldEqual, "johndoe")
			So(resp.AuthData, ShouldResemble, map[string]interface{}{
				"token": "A_TOKEN",
			})

		})

		Convey("executes provider passing error", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return nil, errors.New("worrying error")
			}

			authData := map[string]interface{}{}
			req := skyplugin.AuthRequest{"com.example", "login", authData}

			resp, err := transport.RunProvider(&req)
			So(err.Error(), ShouldEqual, "worrying error")
			So(resp, ShouldBeNil)
		})
	})
}

func TestFactory(t *testing.T) {
	Convey("test factory", t, func() {
		factory := &execTransportFactory{}
		transport := factory.Open("/bin/echo", []string{"plugin"}, skyconfig.Configuration{})

		So(transport, ShouldHaveSameTypeAs, &execTransport{})
		So(transport.(*execTransport).Path, ShouldResemble, "/bin/echo")
		So(transport.(*execTransport).Args, ShouldResemble, []string{
			"plugin",
			"--subprocess",
		})
	})
}
