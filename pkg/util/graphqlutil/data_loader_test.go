package graphqlutil_test

import (
	"context"
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
		loader := graphqlutil.NewDataLoader(func(ctx context.Context, keys []interface{}) ([]interface{}, error) {
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

		load := func(ctx context.Context, id string) func() (string, error) {
			l := loader.Load(ctx, id)
			return func() (string, error) {
				value, err := l.Value()
				if err != nil {
					return "", err
				}
				return value.(string), nil
			}
		}

		Convey("should load values", func() {
			ctx := context.Background()
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 0)
			thunk2 := load(ctx, "2")

			So(loadCounter, ShouldEqual, 0)
			thunk3 := load(ctx, "3")

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
			ctx := context.Background()
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 0)
			thunk2 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 0)
			thunk3 := load(ctx, "2")

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
			ctx := context.Background()
			loadError := errors.New("fail to load")

			loaderErr = loadError
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 0)
			thunk2 := load(ctx, "2")

			So(loadCounter, ShouldEqual, 0)
			thunk3 := load(ctx, "3")

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
			ctx := context.Background()
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 0)
			value1, err := thunk1()
			So(err, ShouldBeNil)
			So(value1, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			thunk2 := load(ctx, "2")

			So(loadCounter, ShouldEqual, 1)
			value2, err := thunk2()
			So(err, ShouldBeNil)
			So(value2, ShouldEqual, "value 2")

			So(loadCounter, ShouldEqual, 2)
			thunk3 := load(ctx, "3")

			So(loadCounter, ShouldEqual, 2)
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "value 3")

			So(loadCounter, ShouldEqual, 3)
			So(loadedIDs, ShouldResemble, [][]string{
				{"1"}, {"2"}, {"3"},
			})
		})

		Convey("should load many", func() {
			ctx := context.Background()
			So(loadCounter, ShouldEqual, 0)

			lazy1 := loader.LoadMany(ctx, []interface{}{"1", "2"})
			values, err := lazy1.Value()
			So(err, ShouldBeNil)
			So(loadCounter, ShouldEqual, 1)
			So(values, ShouldResemble, []interface{}{"value 1", "value 2"})

			lazy2 := loader.LoadMany(ctx, []interface{}{"1", "2", "1"})
			values, err = lazy2.Value()
			So(err, ShouldBeNil)
			So(loadCounter, ShouldEqual, 1)
			So(values, ShouldResemble, []interface{}{"value 1", "value 2", "value 1"})

			loader.ClearAll()

			lazy3 := loader.LoadMany(ctx, []interface{}{"1", "2", "3", "4"})
			values, err = lazy3.Value()
			So(err, ShouldBeNil)
			So(loadCounter, ShouldEqual, 3)
			So(values, ShouldResemble, []interface{}{"value 1", "value 2", "value 3", "value 4"})
		})

		Convey("should reset cached value", func() {
			ctx := context.Background()
			So(loadCounter, ShouldEqual, 0)
			thunk1 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 0)
			value1, err := thunk1()
			So(err, ShouldBeNil)
			So(value1, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			thunk2 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 1)
			value2, err := thunk2()
			So(err, ShouldBeNil)
			So(value2, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 1)
			loader.Clear("1")

			So(loadCounter, ShouldEqual, 1)
			thunk3 := load(ctx, "1")

			So(loadCounter, ShouldEqual, 1)
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "value 1")

			So(loadCounter, ShouldEqual, 2)
			So(loadedIDs, ShouldResemble, [][]string{
				{"1"}, {"1"},
			})
		})

		Convey("should prime value", func() {
			ctx := context.Background()
			So(loadCounter, ShouldEqual, 0)

			thunk1 := load(ctx, "1")
			value1, err := thunk1()
			So(err, ShouldBeNil)
			So(value1, ShouldEqual, "value 1")

			loader.Prime("1", "prime value 1")

			thunk2 := load(ctx, "1")
			value2, err := thunk2()
			So(err, ShouldBeNil)
			So(value2, ShouldEqual, "value 1")

			loader.Clear("1")
			loader.Prime("1", "prime value 1")

			thunk3 := load(ctx, "1")
			value3, err := thunk3()
			So(err, ShouldBeNil)
			So(value3, ShouldEqual, "prime value 1")
		})
	})
}
