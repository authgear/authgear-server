package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql"
)

type mockTransactionHook struct {
	Label string
}

var _ TransactionHook = (*mockTransactionHook)(nil)

func (*mockTransactionHook) WillCommitTx(ctx context.Context) error { return nil }
func (*mockTransactionHook) DidCommitTx(ctx context.Context)        {}

func TestContext(t *testing.T) {
	type contextKeyType struct{}
	var contextKey = contextKeyType{}

	type contextValues struct {
		tx    *sql.Tx
		hooks []TransactionHook
	}

	withValue := func(ctx context.Context, val *contextValues) context.Context {
		return context.WithValue(ctx, contextKey, val)
	}

	getValue := func(ctx context.Context) (*contextValues, bool) {
		val, ok := ctx.Value(contextKey).(*contextValues)
		if !ok {
			return nil, false
		}
		return val, true
	}

	Convey("Test context behavior", t, func() {
		base := context.Background()

		level1 := withValue(base, &contextValues{
			tx: &sql.Tx{},
		})

		// Able to retrieve the value.
		level1Value, ok := getValue(level1)
		So(ok, ShouldBeTrue)
		So(level1Value.tx, ShouldNotBeNil)
		So(level1Value.hooks, ShouldHaveLength, 0)

		// Support in-place modification to the value.
		// The in-place modification is visible.
		level1Value.hooks = append(level1Value.hooks, &mockTransactionHook{Label: "level1-hook1"})
		level1Value, ok = getValue(level1)
		So(ok, ShouldBeTrue)
		So(level1Value.tx, ShouldNotBeNil)
		So(level1Value.hooks, ShouldHaveLength, 1)
		So(level1Value.hooks, ShouldResemble, []TransactionHook{&mockTransactionHook{Label: "level1-hook1"}})

		// Support nesting.
		level2 := withValue(level1, &contextValues{
			tx: &sql.Tx{},
		})
		level2Value, ok := getValue(level2)
		So(ok, ShouldBeTrue)
		So(level2Value.tx, ShouldNotBeNil)
		So(level2Value.hooks, ShouldHaveLength, 0)

		level2Value.hooks = append(level2Value.hooks, &mockTransactionHook{Label: "level2-hook1"})
		level2Value, ok = getValue(level2)
		So(ok, ShouldBeTrue)
		So(level2Value.tx, ShouldNotBeNil)
		So(level2Value.hooks, ShouldHaveLength, 1)
		So(level2Value.hooks, ShouldResemble, []TransactionHook{&mockTransactionHook{Label: "level2-hook1"}})

		// The tx is different
		So(level1Value.tx != level2Value.tx, ShouldBeTrue)
	})
}

type transactionHookThatUseHookHandle struct {
	HookHandle *HookHandle
	Counter    int
}

var _ TransactionHook = (*transactionHookThatUseHookHandle)(nil)

func (h *transactionHookThatUseHookHandle) WillCommitTx(ctx context.Context) error { return nil }
func (h *transactionHookThatUseHookHandle) DidCommitTx(ctx context.Context) {
	_ = h.HookHandle.WithTx(ctx, func(ctx context.Context) error {
		return nil
	})
	h.Counter += 1
}

func TestHookHandle(t *testing.T) {
	Convey("HookHandle", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		pool := NewMockPool_(ctrl)

		hookHandle := &HookHandle{
			Pool: pool,
		}

		// Construct a real connection pool of size 1.
		connPool, err := oteldatabasesql.Open(oteldatabasesql.OpenOptions{
			DriverName: "sqlite3",
			DSN:        ":memory:",
		})
		So(err, ShouldBeNil)
		connPool.SetMaxOpenConns(1)
		connPool.SetMaxIdleConns(0)
		connPool.SetConnMaxLifetime(0)
		connPool.SetConnMaxIdleTime(0)

		pool.EXPECT().Open(gomock.Any(), gomock.Any()).AnyTimes().Return(connPool, nil)

		ctx := context.Background()

		Convey("Does nothing in WithTx() should not block.", func() {
			err := hookHandle.WithTx(ctx, func(ctx context.Context) error {
				return nil
			})
			So(err, ShouldBeNil)
		})

		// Before the fix, this test will block indefinitely.
		Convey("Add a hook that call WithTx() again should not block.", func() {
			hook := &transactionHookThatUseHookHandle{
				HookHandle: hookHandle,
				Counter:    1,
			}
			err := hookHandle.WithTx(ctx, func(ctx context.Context) error {
				hookHandle.UseHook(ctx, hook)
				return nil
			})
			So(hook.Counter, ShouldEqual, 2)
			So(err, ShouldBeNil)
		})
	})
}
