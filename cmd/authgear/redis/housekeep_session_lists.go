package redis

import (
	"context"
	"fmt"
	"regexp"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// A session list maps each session of a user to its expiry, and is never given an expiry itself.
// The application removes an entry only when it lists the sessions and finds the session gone,
// so lists of users who never list their sessions grow forever.
// This removes the same entries the application would, and Redis deletes a list once it is empty.
//
// An entry is removed only if its session key is gone AND its stored expiry has passed.
// The application writes the entry before the session key when it creates a session,
// and the stored expiry of a new session is always in the future, so a new session is never removed.
// A session key never comes back once it is gone, so the session keys need not be watched.

// The key formats in pkg/lib/session/idpsession/store_redis.go and pkg/lib/oauth/redis/keys.go.
var sessionListKeyRegexp = regexp.MustCompile(`^app:[^:]+:session-list:`)
var offlineGrantListKeyRegexp = regexp.MustCompile(`^app:([^:]+):offline-grant-list:`)

var housekeepTaskSessionList = housekeepTask{
	name:  "session-list",
	match: sessionListKeyRegexp.MatchString,
	plan: func(ctx context.Context, tx *goredis.Tx, keys []string, now time.Time) ([]housekeepAction, error) {
		// The field is the session key itself.
		return planPruneSessionLists(ctx, tx, keys, now, func(listKey string, field string) string {
			return field
		})
	},
}

var housekeepTaskOfflineGrantList = housekeepTask{
	name:  "offline-grant-list",
	match: offlineGrantListKeyRegexp.MatchString,
	plan: func(ctx context.Context, tx *goredis.Tx, keys []string, now time.Time) ([]housekeepAction, error) {
		// The field is the offline grant ID.
		return planPruneSessionLists(ctx, tx, keys, now, func(listKey string, field string) string {
			appID := offlineGrantListKeyRegexp.FindStringSubmatch(listKey)[1]
			return fmt.Sprintf("app:%s:offline-grant:%s", appID, field)
		})
	},
}

type sessionList struct {
	key  string
	size int
	// expired are the entries whose stored expiry has passed, with a check of whether their session key exists.
	expired []sessionListEntry
}

type sessionListEntry struct {
	field  string
	exists *goredis.IntCmd
}

func planPruneSessionLists(
	ctx context.Context,
	tx *goredis.Tx,
	keys []string,
	now time.Time,
	sessionKeyOf func(listKey string, field string) string,
) ([]housekeepAction, error) {
	lists, err := readSessionLists(ctx, tx, keys, now, sessionKeyOf)
	if err != nil {
		return nil, err
	}

	var actions []housekeepAction
	for _, list := range lists {
		fields, err := list.deadFields()
		if err != nil {
			return nil, err
		}
		if len(fields) == 0 {
			continue
		}

		// Redis deletes the list when its last entry is removed.
		result := housekeepResultPruned
		if len(fields) == list.size {
			result = housekeepResultDeleted
		}
		key := list.key
		actions = append(actions, housekeepAction{
			key:    key,
			result: result,
			apply: func(ctx context.Context, pipe goredis.Pipeliner) {
				pipe.HDel(ctx, key, fields...)
			},
		})
	}
	return actions, nil
}

// readSessionLists reads the lists, then checks the session keys of their expired entries.
// A key that is not a hash is left out.
func readSessionLists(
	ctx context.Context,
	tx *goredis.Tx,
	keys []string,
	now time.Time,
	sessionKeyOf func(listKey string, field string) string,
) ([]*sessionList, error) {
	hgetalls := make([]*goredis.MapStringStringCmd, len(keys))
	err := readPipelined(ctx, tx, func(pipe goredis.Pipeliner) {
		for i, key := range keys {
			hgetalls[i] = pipe.HGetAll(ctx, key)
		}
	})
	if err != nil {
		return nil, err
	}

	var lists []*sessionList
	err = readPipelined(ctx, tx, func(pipe goredis.Pipeliner) {
		for i, key := range keys {
			entries, err := hgetalls[i].Result()
			if err != nil {
				continue
			}
			list := &sessionList{key: key, size: len(entries)}
			for _, field := range expiredFields(entries, now) {
				list.expired = append(list.expired, sessionListEntry{
					field:  field,
					exists: pipe.Exists(ctx, sessionKeyOf(key, field)),
				})
			}
			lists = append(lists, list)
		}
	})
	if err != nil {
		return nil, err
	}

	for _, hgetall := range hgetalls {
		if _, err := housekeepCmdErr(hgetall.Err()); err != nil {
			return nil, err
		}
	}
	return lists, nil
}

// expiredFields returns the fields whose stored expiry has passed.
// An entry whose expiry cannot be read is kept.
func expiredFields(entries map[string]string, now time.Time) []string {
	var fields []string
	for field, value := range entries {
		var expireAt time.Time
		if err := expireAt.UnmarshalText([]byte(value)); err != nil {
			continue
		}
		if expireAt.Before(now) {
			fields = append(fields, field)
		}
	}
	return fields
}

// deadFields returns the expired entries whose session key is gone.
func (l *sessionList) deadFields() ([]string, error) {
	var fields []string
	for _, entry := range l.expired {
		exists, err := entry.exists.Result()
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			fields = append(fields, entry.field)
		}
	}
	return fields, nil
}
