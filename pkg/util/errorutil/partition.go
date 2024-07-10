package errorutil

import "errors"

// Partition takes a predicate and an error, and returns a pair of joined errors of which do and do not satisfy the predicate, respectively
//
// Example usage
//
//	   matched, notMatched := Partition(someJoinedErr, func(err error) bool {
//		   return errors.Is(err, errA)
//		 })
//
//		 if errors.Is(matched, errA) {
//			 ...
//		 }
func Partition(err error, predicate func(err error) bool) (matched error, notMatched error) {
	var _matched []error
	var _notMatched []error
	Unwrap(err, func(err error) {
		b := predicate(err)
		if b {
			_matched = append(_matched, err)
		} else {
			_notMatched = append(_notMatched, err)
		}
	})
	return errors.Join(_matched...), errors.Join(_notMatched...)
}
