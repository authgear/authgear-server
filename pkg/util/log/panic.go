package log

func PanicValue(logger *Logger, err error) {
	logger.WithError(err).Error("panic occurred")
}
