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

package exec

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	skyplugin "github.com/skygeario/skygear-server/pkg/server/plugin"
	"github.com/skygeario/skygear-server/pkg/server/plugin/common"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func getEnviron(cmd *exec.Cmd, name string) string {
	for _, envdef := range cmd.Env {
		tuple := strings.SplitN(envdef, "=", 2)
		if tuple[0] == name {
			return tuple[1]
		}
	}
	return ""
}

func shouldRunWithContext(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("ShouldEqualJSON receives only one expected argument")
	}

	cmd, ok := actual.(*exec.Cmd)
	if !ok {
		return fmt.Sprintf("%[1]v is %[1]T, not *exec.Cmd", actual)
	}

	name := "SKYGEAR_CONTEXT"
	value := getEnviron(cmd, name)
	if value == "" {
		return fmt.Sprintf(`exec.Cmd does not have environ "%s"`, name)
	}

	var decoded interface{}
	err := common.DecodeBase64JSON(value, &decoded)
	if err != nil {
		return fmt.Sprintf(`unable to decode JSON in environ "%s"`, name)
	}
	return ShouldResemble(decoded, expected[0])
}

func shouldRunWithConfig(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("ShouldEqualJSON receives only one expected argument")
	}

	cmd, ok := actual.(*exec.Cmd)
	if !ok {
		return fmt.Sprintf("%[1]v is %[1]T, not *exec.Cmd", actual)
	}

	name := "SKYGEAR_CONFIG"
	value := getEnviron(cmd, name)
	if value == "" {
		return fmt.Sprintf(`exec.Cmd does not have environ "%s"`, name)
	}

	decoded := skyconfig.Configuration{}
	err := common.DecodeBase64JSON(value, &decoded)
	if err != nil {
		return fmt.Sprintf(`unable to decode JSON in environ "%s"`, name)
	}
	return ShouldResemble(decoded, expected[0])
}

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

		Convey("event", func() {
			out, err := transport.SendEvent("foo-bar", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"event foo-bar"`)
		})

		Convey("handler", func() {
			out, err := transport.RunHandler(nil, "hello:world", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"handler hello:world"`)
		})
	})

	Convey("test stdin", t, func() {
		transport := &execTransport{
			Path: "/bin/sh",
			Args: []string{"-c", `"cat"`},
		}

		Convey("op", func() {
			out, err := transport.RunLambda(nil, "hello:world", []byte(`{"result": "hello world"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"hello world"`)
		})

		Convey("event", func() {
			out, err := transport.SendEvent("foo-bar", []byte(`{"result": "haha"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"haha"`)
		})

		Convey("handler", func() {
			out, err := transport.RunHandler(nil, "hello:world", []byte(`{"result": "hello world"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"hello world"`)
		})
	})

	Convey("test lambda", t, func() {
		appconfig := skyconfig.Configuration{}
		appconfig.App.Name = "app-name"
		transport := &execTransport{
			Path:   "/never/invoked",
			Args:   nil,
			Config: appconfig,
		}

		// expect child test case to override startCommand
		// save the original and defer setting it back
		originalCommand := startCommand
		defer func() {
			startCommand = originalCommand
		}()

		Convey("pass context as environment variable", func() {
			executed := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				So(cmd, shouldRunWithContext, map[string]interface{}{
					"user_id":         "user",
					"access_key_type": "master",
				})
				So(cmd, shouldRunWithConfig, appconfig)
				executed = true
				return []byte(`{"result": {}}`), nil
			}

			ctx := context.WithValue(context.Background(), router.UserIDContextKey, "user")
			ctx = context.WithValue(ctx, router.AccessKeyTypeContextKey, router.MasterAccessKey)
			transport.RunLambda(ctx, "work", []byte{})
			So(executed, ShouldBeTrue)
		})
	})

	Convey("test hook", t, func() {
		appconfig := skyconfig.Configuration{}
		appconfig.App.Name = "app-name"
		transport := &execTransport{
			Path:   "/never/invoked",
			Args:   nil,
			Config: appconfig,
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
				"content":        "some note content",
				"noteOrder":      float64(1),
				"tags":           []interface{}{"test", "unimportant"},
				"date":           time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC),
				"ref":            skydb.NewReference("category", "1"),
				"auto_increment": skydb.Sequence{},
				"asset":          &skydb.Asset{Name: "asset-name"},
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
						"asset": {
							"$type": "asset",
							"$name": "asset-name"
						},
						"auto_increment": {
							"$type": "seq"
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
						"asset": {
							"$type": "asset",
							"$name": "asset-name"
						},
						"auto_increment": {
							"$type": "seq"
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
					"content":        "some note content",
					"noteOrder":      float64(1),
					"tags":           []interface{}{"test", "unimportant"},
					"ref":            skydb.NewReference("category", "1"),
					"auto_increment": skydb.Sequence{},
					"asset":          &skydb.Asset{Name: "asset-name"},
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
					"content":        "content has been modified",
					"noteOrder":      float64(1),
					"tags":           []interface{}{"test", "unimportant"},
					"ref":            skydb.NewReference("category", "1"),
					"auto_increment": skydb.Sequence{},
					"asset":          &skydb.Asset{Name: "asset-name"},
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
						"asset": {
							"$type": "asset",
							"$name": "asset-name"
						},
						"auto_increment": {
							"$type": "seq"
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
						"auto_increment": {
							"$type": "seq"
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
					"content":        "some note content",
					"noteOrder":      float64(1),
					"tags":           []interface{}{"test", "unimportant"},
					"ref":            skydb.NewReference("category", "1"),
					"auto_increment": skydb.Sequence{},
					"asset":          &skydb.Asset{Name: "asset-name"},
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
					"content":        "content has been modified",
					"noteOrder":      float64(1),
					"tags":           []interface{}{"test", "unimportant"},
					"ref":            skydb.NewReference("category", "1"),
					"auto_increment": skydb.Sequence{},
					"asset":          &skydb.Asset{Name: "asset-name"},
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
			So(err.Error(), ShouldEqual, "failed to parse plugin response: invalid character 'I' looking for beginning of value")
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
			executed := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				So(cmd, shouldRunWithContext, map[string]interface{}{
					"user_id":         "user",
					"access_key_type": "master",
				})
				So(cmd, shouldRunWithConfig, appconfig)
				executed = true
				return []byte(`{"result": {}}`), nil
			}

			ctx := context.WithValue(context.Background(), router.UserIDContextKey, "user")
			ctx = context.WithValue(ctx, router.AccessKeyTypeContextKey, router.MasterAccessKey)
			transport.RunHook(ctx, "note_afterSave", &recordin, nil)
			So(executed, ShouldBeTrue)
		})
	})

	Convey("test exec error", t, func() {
		Convey("file not found", func() {
			transport := &execTransport{
				Path: "/tmp/nonexistent",
				Args: []string{},
			}

			_, err := transport.SendEvent("init", []byte{})
			So(err, ShouldNotBeNil)
		})

		Convey("not executable", func() {
			transport := &execTransport{
				Path: "/dev/null",
				Args: []string{},
			}

			_, err := transport.SendEvent("init", []byte{})
			So(err, ShouldNotBeNil)
		})

		Convey("return false", func() {
			transport := &execTransport{
				Path: "/bin/false",
				Args: []string{},
			}

			_, err := transport.SendEvent("init", []byte{})
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
		appconfig := skyconfig.Configuration{}
		appconfig.App.Name = "app-name"
		transport := factory.Open("/bin/echo", []string{"plugin"}, appconfig)

		So(transport, ShouldHaveSameTypeAs, &execTransport{})
		So(transport.(*execTransport).Path, ShouldResemble, "/bin/echo")
		So(transport.(*execTransport).Args, ShouldResemble, []string{
			"plugin",
			"--subprocess",
		})
	})
}
