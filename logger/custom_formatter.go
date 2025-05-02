package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	gormLogger "gorm.io/gorm/logger"
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

func CustomGormLogger() gormLogger.Interface {
	newLogger := gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormLogger.Config{
			LogLevel: gormLogger.Silent,
		},
	)
	return newLogger
}
