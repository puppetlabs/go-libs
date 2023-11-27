// Package log provides facilities for logging.
package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

// DefaultLevel is the default logging level to use, when no other is specified.
const DefaultLevel = logrus.DebugLevel

// init runs when this library is first loaded.
// This function sets the defaults that we want to use.
//
//nolint:gochecknoinits
func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	// Output to stdout instead of the default stderr
	logrus.SetOutput(os.Stdout)

	// Only log the info severity or above.
	logrus.SetLevel(DefaultLevel)
}

// SetLogLevel parses the string and sets the log level to that which is requested.
// if it fails to parse it will fall back to the default level.
func SetLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = DefaultLevel
	}

	// If were up and running downgrade logging to the requested level
	logrus.Infof("Setting log level to %s", level)
	logrus.SetLevel(lvl)
}

// CreateLogger will create a logrus logger with the desired log level.
func CreateLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = logrus.StandardLogger().Formatter
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Warnf("Unable to setup log level %s - defaulting to INFO.", logLevel)
		logger.Level = logrus.InfoLevel
	} else {
		logger.Level = level
	}

	return logger
}
