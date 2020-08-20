package graphqlutil_test

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func TestDataLoader(t *testing.T) {
	Convey("DataLoader", t, func() {
		loadCounter := 0
		var loadedIDs [][]string
		var loaderErr error
		loader := graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			loadCounter++
			ids := make([]string, len(keys))
			values := make([]interface{}, len(keys))
			for i, id := range keys {
				ids[i] = id.(string)
				values[i] = "value " + ids[i]
			}
			loadedIDs = append(loadedIDs, ids)

			if loaderErr != nil {
				return nil, loaderErr
			}
			return values, nil
		})
		loader.MaxBatch = 2

		load := func(id string) func() (string, error) {
			l := loader.Load(id)
			return func() (string, error) {
				value, err := l.Value()
				if err != nil {
					return "", err
				}
				return value.(string), nil
			}
		}

		Convey("should load values", func() {
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load("1")

			So(loadCounter, ShouldEqual, 0)
			thunk2 := load("2")

			So(loadCounter, ShouldEqual, 0)
			thunk3 := load("3")

			So(loadCounter, ShouldEqual, 1)
			value1, err := thunk1()
			So(err, ShouldBeNil)
			So(value1, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			value2, err := thunk2()
			So(err, ShouldBeNil)
			So(value2, ShouldEqual, "value 2")

			So(loadCounter, ShouldEqual, 1)
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "value 3")

			So(loadCounter, ShouldEqual, 2)
			So(loadedIDs, ShouldResemble, [][]string{
				{"1", "2"}, {"3"},
			})
		})

		Convey("should cache values", func() {
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load("1")

			So(loadCounter, ShouldEqual, 0)
			thunk2 := load("1")

			So(loadCounter, ShouldEqual, 0)
			thunk3 := load("2")

			So(loadCounter, ShouldEqual, 0)
			value1, err := thunk1()
			So(err, ShouldBeNil)
			So(value1, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			value2, err := thunk2()
			So(err, ShouldBeNil)
			So(value2, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "value 2")

			So(loadCounter, ShouldEqual, 1)
			So(loadedIDs, ShouldResemble, [][]string{
				{"1", "2"},
			})
		})

		Convey("should propagate errors", func() {
			loadError := errors.New("fail to load")

			loaderErr = loadError
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load("1")

			So(loadCounter, ShouldEqual, 0)
			thunk2 := load("2")

			So(loadCounter, ShouldEqual, 0)
			thunk3 := load("3")

			So(loadCounter, ShouldEqual, 1)
			_, err := thunk1()
			So(err, ShouldEqual, loadError)

			So(loadCounter, ShouldEqual, 1)
			_, err = thunk2()
			So(err, ShouldEqual, loadError)

			loaderErr = nil
			So(loadCounter, ShouldEqual, 1)
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "value 3")

			So(loadCounter, ShouldEqual, 2)
			So(loadedIDs, ShouldResemble, [][]string{
				{"1", "2"}, {"3"},
			})
		})

		Convey("should load individually", func() {
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load("1")

			So(loadCounter, ShouldEqual, 0)
			value1, err := thunk1()
			So(err, ShouldBeNil)
			So(value1, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			thunk2 := load("2")

			So(loadCounter, ShouldEqual, 1)
			value2, err := thunk2()
			So(err, ShouldBeNil)
			So(value2, ShouldEqual, "value 2")

			So(loadCounter, ShouldEqual, 2)
			thunk3 := load("3")

			So(loadCounter, ShouldEqual, 2)
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "value 3")

			So(loadCounter, ShouldEqual, 3)
			So(loadedIDs, ShouldResemble, [][]string{
				{"1"}, {"2"}, {"3"},
			})
		})
	})
}
