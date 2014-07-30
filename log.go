package kocha

import (
	"io"

	"github.com/naoina/kocha/log"
)

// LoggerConfig represents the configuration of the logger.
type LoggerConfig struct {
	Writer    io.Writer     // output destination for the logger.
	Formatter log.Formatter // formatter for log entry.
	Level     log.Level     // log level.
}
