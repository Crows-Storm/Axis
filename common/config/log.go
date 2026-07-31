package config

import (
	"os"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
)

var DefaultLogger *logrus.Entry

func InitLogger(serviceName string) *logrus.Entry {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	level := os.Getenv("LOG_LEVEL")

	if level == "" {
		logger.SetLevel(logrus.WarnLevel)
	} else {
		parseLevel, err := logrus.ParseLevel(level)
		if err != nil {
			return nil
		}
		logger.SetLevel(parseLevel)
	}

	entry := logrus.NewEntry(logger).WithField("service", serviceName)

	grpc_logrus.ReplaceGrpcLogger(entry)

	DefaultLogger = entry
	return entry
}
