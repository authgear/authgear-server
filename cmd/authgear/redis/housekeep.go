package redis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// housekeepTask cleans up one kind of key.
// To add a task, implement it in its own file and append it to housekeepTasks.
type housekeepTask struct {
	name  string
	match func(key string) bool
	// plan reads the keys through tx and returns what should be written.
	// A key without an action is left unchanged.
	plan func(ctx context.Context, tx *goredis.Tx, keys []string, now time.Time) ([]housekeepAction, error)
}

type housekeepAction struct {
	key    string
	result string
	apply  func(ctx context.Context, pipe goredis.Pipeliner)
}

const (
	housekeepResultDeleted     = "deleted"
	housekeepResultExpiryFixed = "expiry-fixed"
	housekeepResultPruned      = "pruned"
	housekeepResultSkipped     = "skipped"
	housekeepResultConflict    = "conflict"
)

var housekeepResults = []string{
	housekeepResultDeleted,
	housekeepResultExpiryFixed,
	housekeepResultPruned,
	housekeepResultSkipped,
	housekeepResultConflict,
}

var housekeepTasks = []housekeepTask{
	housekeepTaskRateLimit,
	housekeepTaskAccessEvents,
	housekeepTaskLockout,
	housekeepTaskFailedAttempts,
	housekeepTaskSessionList,
	housekeepTaskOfflineGrantList,
}

// Keys are watched in chunks, so a write by the application only makes that chunk retry.
const housekeepChunkSize = 100
const housekeepMaxAttempts = 3

type HousekeepOptions struct {
	ScanCountString string
	DryRun          bool
	Now             time.Time
}

type housekeepCounts map[string]int

func (c housekeepCounts) String() string {
	var parts []string
	for _, result := range housekeepResults {
		parts = append(parts, fmt.Sprintf("%v=%v", strings.ReplaceAll(result, "-", "_"), c[result]))
	}
	return strings.Join(parts, " ")
}

// Housekeep scans the whole keyspace once, one SCAN batch at a time, so memory use does not grow with the keyspace.
// Every task only changes keys that are safe to change, and the change is applied only if the keys
// were not written in between, so it can be run again at any time.
func Housekeep(ctx context.Context, redisClient *goredis.Client, stdout io.Writer, logger *log.Logger, options HousekeepOptions) error {
	return housekeep(ctx, redisClient, stdout, logger, options, housekeepTasks)
}

func housekeep(ctx context.Context, redisClient *goredis.Client, stdout io.Writer, logger *log.Logger, options HousekeepOptions, tasks []housekeepTask) error {
	scanCount, err := strconv.ParseInt(options.ScanCountString, 10, 64)
	if err != nil {
		return fmt.Errorf("SCAN count must be an integer: %v", options.ScanCountString)
	}
	if scanCount < 10 {
		return fmt.Errorf("SCAN count must greater than or equal to 10: %v", scanCount)
	}

	if options.Now.IsZero() {
		return fmt.Errorf("now cannot be zero")
	}

	run := &housekeepRun{
		redisClient:  redisClient,
		stdout:       stdout,
		logger:       logger,
		options:      options,
		tasks:        tasks,
		total:        housekeepCounts{},
		countsByTask: make([]housekeepCounts, len(tasks)),
	}
	if options.DryRun {
		run.prefix = "(dry-run) "
	}
	for i := range tasks {
		run.countsByTask[i] = housekeepCounts{}
	}

	var cursor uint64
	for {
		keys, nextCursor, err := redisClient.Scan(ctx, cursor, "", scanCount).Result()
		if err != nil {
			return err
		}
		run.scannedKeyCount += len(keys)

		for i, taskKeys := range run.classify(keys) {
			if err := run.runTask(ctx, i, taskKeys); err != nil {
				return err
			}
		}

		logger.Printf("%vSCAN %v scanned_total=%v %v\n", run.prefix, cursor, run.scannedKeyCount, run.total)

		if nextCursor == 0 {
			run.logSummary()
			return nil
		}
		cursor = nextCursor
	}
}

type housekeepRun struct {
	redisClient     *goredis.Client
	stdout          io.Writer
	logger          *log.Logger
	options         HousekeepOptions
	tasks           []housekeepTask
	prefix          string
	scannedKeyCount int
	total           housekeepCounts
	countsByTask    []housekeepCounts
}

// classify returns the keys of each task, in the order of r.tasks.
func (r *housekeepRun) classify(keys []string) [][]string {
	keysByTask := make([][]string, len(r.tasks))
	for _, key := range keys {
		for i, task := range r.tasks {
			if task.match(key) {
				keysByTask[i] = append(keysByTask[i], key)
				break
			}
		}
	}
	return keysByTask
}

func (r *housekeepRun) runTask(ctx context.Context, taskIndex int, keys []string) error {
	task := r.tasks[taskIndex]
	for start := 0; start < len(keys); start += housekeepChunkSize {
		chunk := keys[start:min(start+housekeepChunkSize, len(keys))]
		actions, ok, err := runHousekeepChunk(ctx, r.redisClient, task, chunk, r.options)
		if err != nil {
			return fmt.Errorf("%v: %w", task.name, err)
		}
		if !ok {
			r.count(taskIndex, housekeepResultConflict, len(chunk))
			r.logger.Printf("%v%v: keys kept changing, left for the next run: %v\n", r.prefix, task.name, strings.Join(chunk, " "))
			continue
		}
		for _, action := range actions {
			r.count(taskIndex, action.result, 1)
			fmt.Fprintf(r.stdout, "%v%v %v\n", r.prefix, action.result, action.key)
		}
		r.count(taskIndex, housekeepResultSkipped, len(chunk)-len(actions))
	}
	return nil
}

func (r *housekeepRun) count(taskIndex int, result string, n int) {
	r.countsByTask[taskIndex][result] += n
	r.total[result] += n
}

func (r *housekeepRun) logSummary() {
	for i, task := range r.tasks {
		r.logger.Printf("%vdone %v %v\n", r.prefix, task.name, r.countsByTask[i])
	}
	r.logger.Printf("%vdone scanned_total=%v %v\n", r.prefix, r.scannedKeyCount, r.total)
}

// runHousekeepChunk returns ok=false if the keys were written by someone else on every attempt.
func runHousekeepChunk(ctx context.Context, redisClient *goredis.Client, task housekeepTask, keys []string, options HousekeepOptions) (actions []housekeepAction, ok bool, err error) {
	for attempt := 0; attempt < housekeepMaxAttempts; attempt++ {
		err = redisClient.Watch(ctx, func(tx *goredis.Tx) error {
			var err error
			actions, err = task.plan(ctx, tx, keys, options.Now)
			if err != nil {
				return err
			}
			if options.DryRun || len(actions) == 0 {
				return nil
			}
			_, err = tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
				for _, action := range actions {
					action.apply(ctx, pipe)
				}
				return nil
			})
			return err
		}, keys...)
		if errors.Is(err, goredis.TxFailedErr) {
			continue
		}
		if err != nil {
			return nil, false, err
		}
		return actions, true, nil
	}
	return nil, false, nil
}

// readPipelined runs the reads in one round trip.
// The caller checks each command with housekeepCmdErr, so that a missing key or a key of an unexpected type
// only skips that key, while any other error still fails the run.
func readPipelined(ctx context.Context, tx *goredis.Tx, fn func(pipe goredis.Pipeliner)) error {
	_, err := tx.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		fn(pipe)
		return nil
	})
	if _, err = housekeepCmdErr(err); err != nil {
		return err
	}
	return nil
}

// housekeepCmdErr reports skip=true for the errors that only mean the key should be left alone.
func housekeepCmdErr(err error) (skip bool, _ error) {
	switch {
	case err == nil:
		return false, nil
	case errors.Is(err, goredis.Nil):
		return true, nil
	case strings.HasPrefix(err.Error(), "WRONGTYPE"):
		return true, nil
	default:
		return false, err
	}
}

func deleteAction(key string) housekeepAction {
	return housekeepAction{
		key:    key,
		result: housekeepResultDeleted,
		apply: func(ctx context.Context, pipe goredis.Pipeliner) {
			pipe.Del(ctx, key)
		},
	}
}
