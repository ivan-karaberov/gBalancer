package logger

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logLine := fmt.Sprintf("%s [%s] %s\n",
		strings.ToUpper(entry.Level.String()),
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Message,
	)

	return []byte(logLine), nil
}
