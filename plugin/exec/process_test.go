package exec

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/oursky/ourd/oddb"
	. "github.com/oursky/ourd/ourtest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRun(t *testing.T) {
	Convey("test args and stdout", t, func() {
		transport := execTransport{
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
			out, err := transport.RunLambda("hello:world", []byte{})
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
		transport := execTransport{
			Path: "/bin/sh",
			Args: []string{"-c", `"cat"`},
		}

		Convey("init", func() {
			out, err := transport.RunInit()
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "")
		})

		Convey("op", func() {
			out, err := transport.RunLambda("hello:world", []byte(`{"result": "hello world"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"hello world"`)
		})

		Convey("handler", func() {
			out, err := transport.RunHandler("hello:world", []byte(`{"result": "hello world"}`))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, `"hello world"`)
		})
	})

	Convey("test hook", t, func() {
		transport := execTransport{
			Path: "/never/invoked",
			Args: nil,
		}

		// expect child test case to override startCommand
		// save the original and defer setting it back
		originalCommand := startCommand
		defer func() {
			startCommand = originalCommand
		}()

		recordin := oddb.Record{
			ID:      oddb.NewRecordID("note", "id"),
			OwnerID: "john.doe@example.com",
			ACL: oddb.RecordACL{
				oddb.NewRecordACLEntryRelation("friend", oddb.WriteLevel),
				oddb.NewRecordACLEntryDirect("user_id", oddb.ReadLevel),
			},
			Data: map[string]interface{}{
				"content":   "some note content",
				"noteOrder": float64(1),
				"tags":      []interface{}{"test", "unimportant"},
			},
		}

		Convey("executes beforeSave correctly", func() {
			called := false
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				called = true
				So(cmd.Path, ShouldEqual, "/never/invoked")
				So(cmd.Args, ShouldResemble, []string{"/never/invoked", "hook", "note:beforeSave"})
				So(in, ShouldEqualJSON, `{
					"_id": "note/id",
					"_ownerID": "john.doe@example.com",
					"_access": [{
						"relation": "friend",
						"level": "write"
					}, {
						"relation": "$direct",
						"level": "read",
						"user_id": "user_id"
					}],
					"data": {
						"content": "some note content",
						"noteOrder": 1,
						"tags": ["test", "unimportant"]
					}
				}`)

				return []byte(`{
					"result": {
						"_id": "note/id",
						"_ownerID": "john.doe@example.com",
						"_access": [{
							"relation": "friend",
							"level": "write"
						}, {
							"relation": "$direct",
							"level": "read",
							"user_id": "user_id"
						}],
						"data": {
							"content": "content has been modified",
							"noteOrder": 1,
							"tags": ["test", "unimportant"]
						}
					}
				}`), nil
			}

			recordout, err := transport.RunHook("note", "beforeSave", &recordin)
			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)

			So(recordin, ShouldResemble, oddb.Record{
				ID:      oddb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: oddb.RecordACL{
					oddb.NewRecordACLEntryRelation("friend", oddb.WriteLevel),
					oddb.NewRecordACLEntryDirect("user_id", oddb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "some note content",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
				},
			})
			So(*recordout, ShouldResemble, oddb.Record{
				ID:      oddb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: oddb.RecordACL{
					oddb.NewRecordACLEntryRelation("friend", oddb.WriteLevel),
					oddb.NewRecordACLEntryDirect("user_id", oddb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "content has been modified",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
				},
			})
		})

		Convey("returns err if command failed", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return nil, errors.New("worrying error")
			}

			recordout, err := transport.RunHook("note", "afterSave", &recordin)
			So(err.Error(), ShouldEqual, "run note:afterSave: worrying error")
			So(recordout, ShouldBeNil)
		})

		Convey("returns err if command returns invalid response", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return []byte("I am not a json"), nil
			}

			recordout, err := transport.RunHook("note", "afterSave", &recordin)
			So(err.Error(), ShouldEqual, "run note:afterSave: failed to parse response: invalid character 'I' looking for beginning of value")
			So(recordout, ShouldBeNil)
		})

		Convey("returns err if commands returns with inner error", func() {
			startCommand = func(cmd *exec.Cmd, in []byte) (out []byte, err error) {
				return []byte(`{
					"result": {
						"ignore": "me"
					},
					"error": {
						"name": "StrongError",
						"desc": "Too strong to lift a feather"
					}
				}`), nil
			}

			recordout, err := transport.RunHook("note", "afterSave", &recordin)
			So(err.Error(), ShouldEqual, `run note:afterSave: StrongError
Too strong to lift a feather`)
			So(recordout, ShouldBeNil)
		})
	})

	Convey("test exec error", t, func() {
		Convey("file not found", func() {
			transport := execTransport{
				Path: "/tmp/nonexistent",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})

		Convey("not executable", func() {
			transport := execTransport{
				Path: "/dev/null",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})

		Convey("return false", func() {
			transport := execTransport{
				Path: "/bin/false",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestFactory(t *testing.T) {
	Convey("test factory", t, func() {
		factory := execTransportFactory{}
		transport := factory.Open("/bin/echo", []string{"plugin"})

		So(transport, ShouldHaveSameTypeAs, execTransport{})
		So(transport.(execTransport).Path, ShouldResemble, "/bin/echo")
		So(transport.(execTransport).Args, ShouldResemble, []string{"plugin"})
	})
}
