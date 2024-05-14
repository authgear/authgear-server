package graphql

import (
	"errors"
	"time"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
var totalUserCountChart = basicChart
var signupByMethodsChart = basicChart

var signupConversionRate = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SignupConversionRate",
	Description: "Signup conversion rate dashboard data",
	Fields: graphql.Fields{
		"totalSignup":               &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"totalSignupUniquePageView": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"conversionRate":            &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
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

var analyticArgs = graphql.FieldConfigArgument{
	"appID": &graphql.ArgumentConfig{
		Type:        graphql.NewNonNull(graphql.ID),
		Description: "Target app ID.",
	},
	"rangeFrom": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphqlutil.Date),
	},
	"rangeTo": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphqlutil.Date),
	},
}

func newAnalyticArgs(configMap graphql.FieldConfigArgument) graphql.FieldConfigArgument {
	for fieldName, argConfig := range analyticArgs {
		configMap[fieldName] = argConfig
	}
	return configMap
}

func getAnalyticArgs(args map[string]interface{}) (appID string, rangeFrom *time.Time, rangeTo *time.Time, err error) {
	appNodeID := args["appID"].(string)
	resolvedNodeID := relay.FromGlobalID(appNodeID)
	if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
		err = apierrors.NewInvalid("invalid app ID")
		return
	}
	appID = resolvedNodeID.ID

	if t, ok := args["rangeFrom"].(time.Time); ok {
		rangeFrom = &t
	}

	if t, ok := args["rangeTo"].(time.Time); ok {
		rangeTo = &t
	}
	return
}
