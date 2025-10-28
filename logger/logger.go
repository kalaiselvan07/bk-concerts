package logger

import (
	"io"
	"log/slog"
	"os"
)

// Log is a global, exported variable that your main package can access.
var Log *slog.Logger

// init() runs automatically when the 'logger' package is imported.
func init() {
	// Open the log file.
	file, err := os.OpenFile("service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		// We can't use our logger here, so we'll panic.
		panic("Failed to open log file: " + err.Error())
	}

	// Create a writer that writes to both the console (os.Stdout) and the file.
	writer := io.MultiWriter(os.Stdout, file)

	// Create a new slog logger with a TextHandler that writes to our multi-writer.
	// We set the log level to Debug to see all log levels.
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Set our global Log variable.
	Log = slog.New(handler)
}
