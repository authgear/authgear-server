package graphql

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"
)

var periodicalEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "Periodical",
	Values: graphql.EnumValueConfigMap{
		"MONTHLY": &graphql.EnumValueConfig{
			Value: "monthly",
		},
		"WEEKLY": &graphql.EnumValueConfig{
			Value: "weekly",
		},
	},
})

var datapoint = graphql.NewObject(graphql.ObjectConfig{
	Name: "DataPoint",
	Fields: graphql.Fields{
		"label": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"data": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Float),
		},
	},
})

var basicChart = graphql.NewObject(graphql.ObjectConfig{
	Name: "Chart",
	Fields: graphql.Fields{
		"dataset": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(datapoint)),
		},
	},
})

var activeUserChart = basicChart

var signupSummary = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SignupSummary",
	Description: "Signup summary for analytic dashboard",
	Fields: graphql.Fields{
		"totalUserCount":            &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalSignup":               &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalSignupPageCount":      &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalSignupUniquePageView": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalLoginPageView":        &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalLoginUniquePageView":  &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"conversionRate":            &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
		"signupByChannelChart":      &graphql.Field{Type: graphql.NewNonNull(basicChart)},
		"totalUserCountChart":       &graphql.Field{Type: graphql.NewNonNull(basicChart)},
	},
})

// checkChartDateRangeInput check the date range input and limit the range
func checkChartDateRangeInput(rangeFrom *time.Time, rangeTo *time.Time) error {
	if rangeFrom == nil || rangeTo == nil {
		return errors.New("missing date range for chart")
	}
	rangeToLimit := rangeFrom.AddDate(1, 0, 0)
	if rangeToLimit.Before(*rangeTo) {
		return errors.New("exceed the maximum 1 year date range")
	}
	return nil
}
