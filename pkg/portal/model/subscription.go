package model

import (
	"fmt"
	"math"
	"time"
)

type SubscriptionCheckoutStatus string

const (
	SubscriptionCheckoutStatusOpen       SubscriptionCheckoutStatus = "open"
	SubscriptionCheckoutStatusCompleted  SubscriptionCheckoutStatus = "completed"
	SubscriptionCheckoutStatusSubscribed SubscriptionCheckoutStatus = "subscribed"
	SubscriptionCheckoutStatusCancelled  SubscriptionCheckoutStatus = "cancelled"
)

// Subscription represents an app subscription.
// The keys in JSON struct tags are in camel case
// because this struct is directly returned in the GraphQL endpoint.
// Making the keys in camel case saves us from writing boilerplate resolver code.
type Subscription struct {
	ID                   string     `json:"id"`
	AppID                string     `json:"appID"`
	StripeSubscriptionID string     `json:"stripeSubscriptionID"`
	StripeCustomerID     string     `json:"stripeCustomerID"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	CancelledAt          *time.Time `json:"cancelledAt"`
	EndedAt              *time.Time `json:"endedAt"`
}

type SubscriptionCheckout struct {
	ID                      string
	AppID                   string
	StripeCheckoutSessionID string
	StripeCustomerID        *string
	Status                  SubscriptionCheckoutStatus
	CreatedAt               time.Time
	UpdatedAt               time.Time
	ExpireAt                time.Time
}

type PriceType string

const (
	PriceTypeFixed PriceType = "fixed"
	PriceTypeUsage PriceType = "usage"
)

func (t PriceType) Valid() error {
	switch t {
	case PriceTypeFixed:
		return nil
	case PriceTypeUsage:
		return nil
	}
	return fmt.Errorf("unknown price_type: %#v", t)
}

type UsageType string

const (
	UsageTypeNone UsageType = ""
	UsageTypeSMS  UsageType = "sms"
	UsageTypeMAU  UsageType = "mau"
)

func (t UsageType) Valid() error {
	switch t {
	case UsageTypeNone:
		return nil
	case UsageTypeSMS:
		return nil
	case UsageTypeMAU:
		return nil
	}
	return fmt.Errorf("unknown usage_type: %#v", t)
}

type SMSRegion string

const (
	SMSRegionNone         SMSRegion = ""
	SMSRegionNorthAmerica SMSRegion = "north-america"
	SMSRegionOtherRegions SMSRegion = "other-regions"
)

func (r SMSRegion) Valid() error {
	switch r {
	case SMSRegionNone:
		return nil
	case SMSRegionNorthAmerica:
		return nil
	case SMSRegionOtherRegions:
		return nil
	}
	return fmt.Errorf("unknown sms_region: %#v", r)
}

type TransformQuantityRound string

const (
	TransformQuantityRoundNone TransformQuantityRound = ""
	TransformQuantityRoundUp   TransformQuantityRound = "up"
	TransformQuantityRoundDown TransformQuantityRound = "down"
)

func (r TransformQuantityRound) Valid() error {
	switch r {
	case TransformQuantityRoundNone:
		return nil
	case TransformQuantityRoundUp:
		return nil
	case TransformQuantityRoundDown:
		return nil
	}
	return fmt.Errorf("unknown round: %#v", r)
}

type SubscriptionUsage struct {
	NextBillingDate time.Time                `json:"nextBillingDate"`
	Items           []*SubscriptionUsageItem `json:"items,omitempty"`
}

type SubscriptionUsageItem struct {
	Type                      PriceType              `json:"type"`
	UsageType                 UsageType              `json:"usageType"`
	SMSRegion                 SMSRegion              `json:"smsRegion"`
	Currency                  *string                `json:"currency"`
	UnitAmount                *int                   `json:"unitAmount"`
	FreeQuantity              *int                   `json:"freeQuantity,omitempty"`
	TransformQuantityDivideBy *int                   `json:"transformQuantityDivideBy,omitempty"`
	TransformQuantityRound    TransformQuantityRound `json:"transformQuantityRound,omitempty"`

	Quantity    int  `json:"quantity"`
	TotalAmount *int `json:"totalAmount"`
}

type Price struct {
	StripePriceID             string                 `json:"stripePriceID"`
	StripeProductID           string                 `json:"stripeProductID"`
	Type                      PriceType              `json:"type"`
	UsageType                 UsageType              `json:"usageType,omitempty"`
	SMSRegion                 SMSRegion              `json:"smsRegion,omitempty"`
	Currency                  string                 `json:"currency"`
	UnitAmount                int                    `json:"unitAmount"`
	FreeQuantity              *int                   `json:"freeQuantity,omitempty"`
	TransformQuantityDivideBy *int                   `json:"transformQuantityDivideBy,omitempty"`
	TransformQuantityRound    TransformQuantityRound `json:"transformQuantityRound,omitempty"`
}

func (p *Price) ShouldClearUsage() bool {
	return p.Type == PriceTypeUsage && p.UsageType == UsageTypeMAU
}

func (i *SubscriptionUsageItem) Match(p *Price) bool {
	return i.Type == p.Type && i.UsageType == p.UsageType && i.SMSRegion == p.SMSRegion
}

func (i *SubscriptionUsageItem) FillFrom(p *Price) *SubscriptionUsageItem {
	i.Currency = &p.Currency
	i.UnitAmount = &p.UnitAmount
	i.FreeQuantity = p.FreeQuantity
	i.TransformQuantityDivideBy = p.TransformQuantityDivideBy
	i.TransformQuantityRound = p.TransformQuantityRound

	// Apply FreeQuantity first.
	quantity := i.Quantity
	if i.FreeQuantity != nil {
		quantity = quantity - *i.FreeQuantity
	}
	if quantity < 0 {
		quantity = 0
	}

	// Apply transformQuantity second.
	if i.TransformQuantityDivideBy != nil {
		quantityF := float64(quantity) / float64(*i.TransformQuantityDivideBy)
		switch i.TransformQuantityRound {
		case TransformQuantityRoundUp:
			quantityF = math.Ceil(quantityF)
		case TransformQuantityRoundDown:
			quantityF = math.Floor(quantityF)
		default:
			quantityF = math.Ceil(quantityF)
		}
		quantity = int(quantityF)
	}

	totalAmount := quantity * *i.UnitAmount
	i.TotalAmount = &totalAmount

	return i
}

type SubscriptionPlan struct {
	Name   string   `json:"name"`
	Prices []*Price `json:"prices,omitempty"`
}

func NewSubscriptionPlan(planName string) *SubscriptionPlan {
	return &SubscriptionPlan{
		Name: planName,
	}
}

type SubscriptionUpdatePreview struct {
	Currency  string `json:"currency"`
	AmountDue int    `json:"amountDue"`
}
