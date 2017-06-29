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

package skydb

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQuery(t *testing.T) {
	Convey("Query", t, func() {
		Convey("Accept", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			Convey("should call QueryVisitor", func() {
				q := Query{}

				v := NewMockQueryVisitor(ctrl)
				call := v.EXPECT().VisitQuery(gomock.Eq(q))
				v.EXPECT().EndVisitQuery(gomock.Eq(q)).After(call)
				q.Accept(v)
			})

			Convey("should call all visitors", func() {
				helloExpr := Expression{
					Type:  Literal,
					Value: "helllo",
				}
				q := Query{
					Predicate: Predicate{},
					Sorts: []Sort{
						{
							Expression: Expression{},
						},
					},
					ComputedKeys: map[string]Expression{
						"hello": helloExpr,
					},
				}

				v := NewMockFullQueryVisitor(ctrl)
				call := v.EXPECT().VisitQuery(gomock.Eq(q))
				call = v.EXPECT().VisitPredicate(gomock.Eq(q.Predicate)).After(call)
				call = v.EXPECT().EndVisitPredicate(gomock.Eq(q.Predicate)).After(call)
				call = v.EXPECT().VisitSort(gomock.Eq(q.Sorts[0])).After(call)
				call = v.EXPECT().VisitExpression(gomock.Any()).After(call)
				call = v.EXPECT().EndVisitExpression(gomock.Any()).After(call)
				call = v.EXPECT().EndVisitSort(gomock.Eq(q.Sorts[0])).After(call)
				call = v.EXPECT().VisitExpression(gomock.Eq(helloExpr)).After(call)
				call = v.EXPECT().EndVisitExpression(gomock.Eq(helloExpr)).After(call)
				v.EXPECT().EndVisitQuery(gomock.Eq(q)).After(call)
				q.Accept(v)
			})
		})
	})
}

func TestPredicate(t *testing.T) {
	Convey("Predicate", t, func() {
		Convey("Accept", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			Convey("should call PredicateVisitor", func() {
				p := Predicate{}

				v := NewMockPredicateVisitor(ctrl)
				call := v.EXPECT().VisitPredicate(gomock.Eq(p))
				v.EXPECT().EndVisitPredicate(gomock.Eq(p)).After(call)
				p.Accept(v)
			})

			Convey("should call PredicateVisitor for compound predicate", func() {
				child := Predicate{}
				p := Predicate{
					Operator: Not,
					Children: []interface{}{child},
				}

				v := NewMockPredicateVisitor(ctrl)
				call := v.EXPECT().VisitPredicate(gomock.Eq(p))
				call = v.EXPECT().VisitPredicate(gomock.Eq(child)).After(call)
				call = v.EXPECT().EndVisitPredicate(gomock.Eq(child)).After(call)
				v.EXPECT().EndVisitPredicate(gomock.Eq(p)).After(call)
				p.Accept(v)
			})

			Convey("should call ExpressionVisitor for simple predicate", func() {
				child := Expression{}
				p := Predicate{
					Operator: Equal,
					Children: []interface{}{child},
				}

				v := NewMockExpressionVisitor(ctrl)
				call := v.EXPECT().VisitExpression(gomock.Eq(child))
				v.EXPECT().EndVisitExpression(gomock.Eq(child)).After(call)
				p.Accept(v)
			})
		})
	})
}

func TestSort(t *testing.T) {
	Convey("Sort", t, func() {
		Convey("Accept", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			Convey("should call SortVisitor", func() {
				child := Expression{}
				sort := Sort{
					Expression: child,
				}

				v := NewMockSortVisitor(ctrl)
				call := v.EXPECT().VisitSort(gomock.Eq(sort))
				v.EXPECT().EndVisitSort(gomock.Eq(sort)).After(call)
				sort.Accept(v)
			})

			Convey("should call ExpressionVisitor", func() {
				child := Expression{}
				sort := Sort{
					Expression: child,
				}

				v := NewMockExpressionVisitor(ctrl)
				call := v.EXPECT().VisitExpression(gomock.Eq(child))
				v.EXPECT().EndVisitExpression(gomock.Eq(child)).After(call)
				sort.Accept(v)
			})
		})
	})
}

func TestExpression(t *testing.T) {
	Convey("is empty", t, func() {
		expr := Expression{}
		So(expr.IsEmpty(), ShouldBeTrue)
	})

	Convey("is literal null", t, func() {
		expr := Expression{Type: Literal, Value: nil}
		So(expr.IsLiteralNull(), ShouldBeTrue)
	})

	Convey("Accept", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("should call ExpressionVisitor", func() {
			expr := Expression{}

			v := NewMockExpressionVisitor(ctrl)
			call := v.EXPECT().VisitExpression(gomock.Eq(expr))
			v.EXPECT().EndVisitExpression(gomock.Eq(expr)).After(call)
			expr.Accept(v)
		})
	})
}

func TestUserDiscoverFunc(t *testing.T) {
	Convey("func with only emails", t, func() {
		f := UserDiscoverFunc{
			Emails: []string{"abc@example.com"},
		}

		Convey("should have args for email", func() {
			So(f.HaveArgsByName("email"), ShouldBeTrue)
		})

		Convey("should not have args for username", func() {
			So(f.HaveArgsByName("username"), ShouldBeFalse)
		})

		Convey("should return args for email", func() {
			So(f.ArgsByName("email"), ShouldResemble, []interface{}{"abc@example.com"})
		})

		Convey("should return args for username", func() {
			So(f.ArgsByName("username"), ShouldResemble, []interface{}{})
		})
	})
}

func TestMalformedPredicate(t *testing.T) {
	Convey("Predicate with Equal", t, func() {
		Convey("comparing array", func() {
			predicate := Predicate{
				Operator: Equal,
				Children: []interface{}{
					Expression{
						Type:  KeyPath,
						Value: "categories",
					},
					Expression{
						Type:  Literal,
						Value: []interface{}{},
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})

		Convey("comparing map", func() {
			predicate := Predicate{
				Operator: Equal,
				Children: []interface{}{
					Expression{
						Type:  KeyPath,
						Value: "categories",
					},
					Expression{
						Type:  Literal,
						Value: map[string]interface{}{},
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Predicate with User Discover", t, func() {
		Convey("cannot be combined", func() {
			predicate := Predicate{
				Not,
				[]interface{}{
					Predicate{
						Functional,
						[]interface{}{
							Expression{
								Type: Function,
								Value: UserDiscoverFunc{
									Emails: []string{
										"john.doe@example.com",
										"jane.doe@example.com",
									},
								},
							},
						},
					},
				},
			}

			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})
	})
}
