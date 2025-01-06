package sqlmigrate

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

/// This file contains utilities for implementing a cobra-powered CLI program to manage migrations.

const CobraMigrateStatusUse = "status"
const CobraMigrateStatusShort = "Get database schema migration status"

var CobraMigrateStatusArgs = cobra.ExactArgs(0)

const CobraMigrateUpUse = "up [n]"
const CobraMigrateUpShort = "Migrate at most n versions up. When n is omitted, migrate to latest version. n must be a positive integer."

var CobraMigrateUpArgs = cobra.MaximumNArgs(1)

func CobraParseMigrateUpArgs(args []string) (n int, err error) {
	if len(args) == 0 {
		n = 0
		return
	}

	n, err = strconv.Atoi(args[0])
	if err != nil {
		err = fmt.Errorf("n must be a positive integer: %w", err)
		return
	}

	if n <= 0 {
		err = fmt.Errorf("n must be a positive integer: %v", n)
		return
	}

	return
}

const CobraMigrateDownUse = "down n"
const CobraMigrateDownShort = "Migrate n versions down. YOU ALMOST NEVER NEED THIS. n MUST BE a positive integer. As a special case, n can be the string 'all' to mean all versions."

var CobraMigrateDownArgs = cobra.ExactArgs(1)

func CobraParseMigrateDownArgs(args []string) (n int, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("n must be a positive integer")
		return
	}

	if args[0] == "all" {
		n = 0
		return
	}

	n, err = strconv.Atoi(args[0])
	if err != nil {
		err = fmt.Errorf("n must be a positive integer: %w", err)
		return
	}

	if n <= 0 {
		err = fmt.Errorf("n must be a positive integer: %v", n)
		return
	}

	return
}
