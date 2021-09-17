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
			Type: graphql.NewNonNull(graphql.Int),
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
