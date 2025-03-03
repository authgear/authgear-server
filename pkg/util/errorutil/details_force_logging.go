package errorutil

const (
	forceLoggingDetail DetailTag = "force_logging"
)

func ForceLogging(err error) error {
	return WithDetails(err, Details{
		"force_logging": forceLoggingDetail.Value(true),
	})
}

func IsForceLogging(err error) bool {
	details := CollectDetails(err, nil)
	details = FilterDetails(details, forceLoggingDetail)
	if len(details) > 0 {
		return true
	}
	return false
}
