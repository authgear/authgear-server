package graphql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var fraudProtectionDecisionEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "FraudProtectionDecision",
	Values: graphql.EnumValueConfigMap{
		"allowed": &graphql.EnumValueConfig{Value: model.FraudProtectionDecisionAllowed},
		"blocked": &graphql.EnumValueConfig{Value: model.FraudProtectionDecisionBlocked},
	},
})

var fraudProtectionActionEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "FraudProtectionAction",
	Values: graphql.EnumValueConfigMap{
		"send_sms": &graphql.EnumValueConfig{Value: model.FraudProtectionActionSendSMS},
	},
})

var fraudProtectionWarningTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "FraudProtectionWarningType",
	Values: graphql.EnumValueConfigMap{
		string(config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily): &graphql.EnumValueConfig{
			Value: string(config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily),
		},
		string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily): &graphql.EnumValueConfig{
			Value: string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily),
		},
		string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly): &graphql.EnumValueConfig{
			Value: string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly),
		},
		string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily): &graphql.EnumValueConfig{
			Value: string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily),
		},
		string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly): &graphql.EnumValueConfig{
			Value: string(config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly),
		},
	},
})

var fraudProtectionDecisionRecordData = graphqlutil.NewJSONObjectScalar(
	"FraudProtectionDecisionRecordData",
	"The `FraudProtectionDecisionRecordData` scalar type represents the raw fraud protection decision record payload.",
)

var fraudProtectionDecisionSendSMSActionDetailType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionDecisionSendSMSActionDetail",
	Fields: graphql.Fields{
		"recipient": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(model.FraudProtectionDecisionActionDetail)
				return source.Recipient, nil
			},
		},
		"type": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(model.FraudProtectionDecisionActionDetail)
				return source.Type, nil
			},
		},
		"phoneNumberCountryCode": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(model.FraudProtectionDecisionActionDetail)
				return source.PhoneNumberCountryCode, nil
			},
		},
	},
})

var fraudProtectionDecisionActionDetailUnion = graphql.NewUnion(graphql.UnionConfig{
	Name: "FraudProtectionDecisionActionDetail",
	Types: []*graphql.Object{
		fraudProtectionDecisionSendSMSActionDetailType,
	},
	ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
		switch p.Value.(type) {
		case model.FraudProtectionDecisionActionDetail:
			return fraudProtectionDecisionSendSMSActionDetailType
		default:
			return nil
		}
	},
})

const typeFraudProtectionDecisionRecord = "FraudProtectionDecisionRecord"

var fraudProtectionDecisionRecordType = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeFraudProtectionDecisionRecord,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeFraudProtectionDecisionRecord, nil),
			"createdAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.DateTime),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.CreatedAt, nil
				},
			},
			"decision": &graphql.Field{
				Type: graphql.NewNonNull(fraudProtectionDecisionEnum),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.Decision, nil
				},
			},
			"action": &graphql.Field{
				Type: graphql.NewNonNull(fraudProtectionActionEnum),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.Action, nil
				},
			},
			"actionDetail": &graphql.Field{
				Type: graphql.NewNonNull(fraudProtectionDecisionActionDetailUnion),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.ActionDetail, nil
				},
			},
			"triggeredWarnings": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionWarningTypeEnum))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.TriggeredWarnings, nil
				},
			},
			"userAgent": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.UserAgent, nil
				},
			},
			"ipAddress": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.IPAddress, nil
				},
			},
			"geoLocationCode": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					return source.Record.GeoLocationCode, nil
				},
			},
			"data": &graphql.Field{
				Type: fraudProtectionDecisionRecordData,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.FraudProtectionDecisionRecord)
					var m map[string]interface{}
					b, err := json.Marshal(source.Record)
					if err != nil {
						return nil, err
					}
					if err := json.Unmarshal(b, &m); err != nil {
						return nil, err
					}
					return m, nil
				},
			},
		},
	}),
	&audit.FraudProtectionDecisionRecord{},
	func(ctx context.Context, gqlCtx *Context, id string) (interface{}, error) {
		return gqlCtx.AuditLogFacade.GetFraudProtectionDecisionRecordByID(ctx, id)
	},
)

var connFraudProtectionDecisionRecord = graphqlutil.NewConnectionDef(fraudProtectionDecisionRecordType)

func fraudProtectionDecisionRecordQueryOptionsFromArgs(
	p graphql.ResolveParams,
) audit.FraudProtectionDecisionRecordQueryOptions {
	var rangeFrom *time.Time
	if t, ok := p.Args["rangeFrom"].(time.Time); ok {
		rangeFrom = &t
	}

	var rangeTo *time.Time
	if t, ok := p.Args["rangeTo"].(time.Time); ok {
		rangeTo = &t
	}

	sortDirection, _ := p.Args["sortDirection"].(model.SortDirection)

	var decisions []model.FraudProtectionDecision
	if arr, ok := p.Args["verdicts"].([]interface{}); ok {
		for _, value := range arr {
			if decision, ok := value.(model.FraudProtectionDecision); ok {
				decisions = append(decisions, decision)
			}
		}
	}

	toStringSlice := func(key string) []string {
		var out []string
		if arr, ok := p.Args[key].([]interface{}); ok {
			for _, value := range arr {
				if s, ok := value.(string); ok {
					out = append(out, s)
				}
			}
		}
		return out
	}

	var maximumWarningCount *int
	if n, ok := p.Args["maximumWarningCount"].(int); ok {
		maximumWarningCount = &n
	}

	var minimumWarningCount *int
	if n, ok := p.Args["minimumWarningCount"].(int); ok {
		minimumWarningCount = &n
	}

	return audit.FraudProtectionDecisionRecordQueryOptions{
		RangeFrom:           rangeFrom,
		RangeTo:             rangeTo,
		SortDirection:       sortDirection,
		Decisions:           decisions,
		ReasonCodes:         toStringSlice("reasonCodes"),
		MaximumWarningCount: maximumWarningCount,
		MinimumWarningCount: minimumWarningCount,
		Search: func() *string {
			search, _ := p.Args["search"].(string)
			if search == "" {
				return nil
			}
			return &search
		}(),
	}
}
