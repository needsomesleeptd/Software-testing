package unit_test_utils

import (
	"bytes"
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	buff       bytes.Buffer
	MockLogger = logrus.New()
	ErrEmpty   = errors.New("")
)

func init() {
	MockLogger.Out = &buff
	MockLogger.Level = logrus.DebugLevel
	MockLogger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
}
