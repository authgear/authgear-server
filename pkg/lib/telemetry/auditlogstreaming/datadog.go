package auditlogstreaming

import (
	"encoding/json"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/event"
)

// ResolvedDatadog is the datadog-encoding-relevant part of a resolved
// stream. It is separate from ResolvedStream (resolve.go) because the
// encoder does not need the stream's transport, following the same split
// ResolvedSyslog (encoder.go) makes for syslog.
type ResolvedDatadog struct {
	APIKey  string
	Service string
	Source  string
	// Tags are datadog.tags, sorted by key, with app_id and activity_type
	// already removed -- those two are appended per entry, from the event.
	Tags []DatadogTag
}

type DatadogTag struct {
	Key   string
	Value string
}

type datadogLog struct {
	DDSource  string `json:"ddsource"`
	Service   string `json:"service"`
	Hostname  string `json:"hostname,omitempty"`
	DDTags    string `json:"ddtags,omitempty"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`

	Evt     *datadogEvt     `json:"evt,omitempty"`
	Usr     *datadogUsr     `json:"usr,omitempty"`
	Network *datadogNetwork `json:"network,omitempty"`
	HTTP    *datadogHTTP    `json:"http,omitempty"`

	Authgear datadogAuthgear `json:"authgear"`
}

type datadogEvt struct {
	Name string `json:"name,omitempty"`
}

type datadogUsr struct {
	ID string `json:"id,omitempty"`
}

type datadogNetwork struct {
	Client *datadogNetworkClient `json:"client,omitempty"`
}

type datadogNetworkClient struct {
	IP    string               `json:"ip,omitempty"`
	GeoIP *datadogNetworkGeoIP `json:"geoip,omitempty"`
}

type datadogNetworkGeoIP struct {
	Country datadogNetworkCountry `json:"country"`
}

type datadogNetworkCountry struct {
	ISOCode string `json:"iso_code,omitempty"`
}

type datadogHTTP struct {
	UserAgent string `json:"useragent,omitempty"`
	URL       string `json:"url,omitempty"`
	Referer   string `json:"referer,omitempty"`
}

// datadogAuthgear carries the event object verbatim. Event is the bytes
// the producer enqueued (QueuedEntry.Raw), not a re-marshalling of
// QueuedEntry.Event: that struct has no Payload (see QueuedEntry's doc
// comment), so re-marshalling would silently drop the whole payload.
type datadogAuthgear struct {
	Event json.RawMessage `json:"event"`
}

// EncodeDatadogLog builds one Datadog log object for entry, per
// docs/specs/audit-log-streaming.md#the-log-object.
func EncodeDatadogLog(d *ResolvedDatadog, hostname string, entry QueuedEntry) ([]byte, error) {
	e := entry.Event

	log := datadogLog{
		DDSource:  d.Source,
		Service:   d.Service,
		Hostname:  hostname,
		DDTags:    buildDDTags(d.Tags, e.Context.AppID, e.Type),
		Message:   string(e.Type),
		Status:    datadogStatus(e.Type),
		Timestamp: e.Context.Timestamp * 1000,
		Evt:       &datadogEvt{Name: string(e.Type)},
		Authgear:  datadogAuthgear{Event: json.RawMessage(entry.Raw)},
	}

	if e.Context.UserID != nil && *e.Context.UserID != "" {
		log.Usr = &datadogUsr{ID: *e.Context.UserID}
	}

	var geoip *datadogNetworkGeoIP
	if e.Context.GeoLocationCode != nil && *e.Context.GeoLocationCode != "" {
		geoip = &datadogNetworkGeoIP{Country: datadogNetworkCountry{ISOCode: *e.Context.GeoLocationCode}}
	}
	if e.Context.IPAddress != "" || geoip != nil {
		log.Network = &datadogNetwork{Client: &datadogNetworkClient{
			IP:    e.Context.IPAddress,
			GeoIP: geoip,
		}}
	}

	httpURL, _ := e.Context.AuditContext["http_url"].(string)
	httpReferer, _ := e.Context.AuditContext["http_referer"].(string)
	if e.Context.UserAgent != "" || httpURL != "" || httpReferer != "" {
		log.HTTP = &datadogHTTP{
			UserAgent: e.Context.UserAgent,
			URL:       httpURL,
			Referer:   httpReferer,
		}
	}

	return json.Marshal(log)
}

// datadogStatus maps an activity type to a Datadog status. It reads the
// same severityByType table PRI does (encoder.go), so adding an activity
// type to that table changes both types' classification at once.
func datadogStatus(t event.Type) string {
	if _, ok := severityByType[t]; ok {
		return "warning"
	}
	return "info"
}

// buildDDTags writes tags as "key:value" pairs joined by ",": every
// element of tags in order, then app_id, then activity_type. Nothing is
// escaped or normalised -- the spec is explicit that a tag which does not
// conform to Datadog's rules is Datadog's to normalise or reject.
func buildDDTags(tags []DatadogTag, appID string, activityType event.Type) string {
	var parts []string
	for _, tag := range tags {
		parts = append(parts, tag.Key+":"+tag.Value)
	}
	if appID != "" {
		parts = append(parts, "app_id:"+appID)
	}
	parts = append(parts, "activity_type:"+string(activityType))
	return strings.Join(parts, ",")
}
