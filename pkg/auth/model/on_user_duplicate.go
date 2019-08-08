package model

// OnUserDuplicate is the strategy to handle user duplicate
type OnUserDuplicate string

// OnUserDuplicate constants
const (
	OnUserDuplicateAbort  OnUserDuplicate = "abort"
	OnUserDuplicateMerge  OnUserDuplicate = "merge"
	OnUserDuplicateCreate OnUserDuplicate = "create"
)

// OnUserDuplicateDefault is OnUserDuplicateAbort
const OnUserDuplicateDefault = OnUserDuplicateAbort

// IsValidOnUserDuplicateForSSO validates OnUserDuplicate
func IsValidOnUserDuplicateForSSO(input OnUserDuplicate) bool {
	allVariants := []OnUserDuplicate{OnUserDuplicateAbort, OnUserDuplicateMerge, OnUserDuplicateCreate}
	for _, v := range allVariants {
		if input == v {
			return true
		}
	}
	return false
}

// IsAllowedOnUserDuplicate checks if input is allowed
func IsAllowedOnUserDuplicate(onUserDuplicateAllowMerge bool, onUserDuplicateAllowCreate bool, input OnUserDuplicate) bool {
	if !onUserDuplicateAllowMerge && input == OnUserDuplicateMerge {
		return false
	}
	if !onUserDuplicateAllowCreate && input == OnUserDuplicateCreate {
		return false
	}
	return true
}
