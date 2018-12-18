package testing

import (
	"bytes"

	"github.com/Sirupsen/logrus"
)

func NewLogger() (*bytes.Buffer, *logrus.Entry) {
	logger := logrus.New()

	b := bytes.NewBufferString("")
	logger.SetOutput(b)
	logger.SetLevel(logrus.ErrorLevel)

	return b, logrus.NewEntry(logger)
}
