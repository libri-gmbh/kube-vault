package testing

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

// NewLogger returns a new logrus.Entry instance and a buffer which is configured as a stream output
func NewLogger() (*bytes.Buffer, *logrus.Entry) {
	logger := logrus.New()

	b := bytes.NewBufferString("")
	logger.SetOutput(b)
	logger.SetLevel(logrus.ErrorLevel)

	return b, logrus.NewEntry(logger)
}
