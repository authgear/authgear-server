package accesscontrol

type Role string

const EmptyRole Role = ""

type Subject string

// Level must start with 1.
type Level int

type T map[Subject]map[Role]Level

func (t T) GetLevel(subject Subject, role Role, defaultLevel Level) Level {
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
