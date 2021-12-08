package accesscontrol

type Role string

// RoleGreatest is the greatest role.
// The level associated with it is RoleGreatest.
const RoleGreatest Role = "__greatest__"

type Subject string

// Level must start with 1.
type Level int

// LevelGreatest is the greatest level.
// It is always greater than any other role.
const LevelGreatest Level = 1000

type T map[Subject]map[Role]Level

func (t T) GetLevel(subject Subject, role Role, defaultLevel Level) Level {
	if role == RoleGreatest {
		return LevelGreatest
	}

	roleLevel, ok := t[subject]
	if !ok {
		return defaultLevel
	}
	level, ok := roleLevel[role]
	if !ok {
		return defaultLevel
	}
	return level
}
