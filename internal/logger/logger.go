package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Format string

const (
	Pretty Format = "pretty"
	JSON   Format = "json"
	Silent Format = "silent"
)

var format Format = Pretty

func SetFormat(f Format) { format = f }
func GetFormat() Format  { return format }

// Info logs a startup/info message. Always printed (even in silent mode).
func Info(msg string) {
	switch format {
	case JSON:
		writeJSON("info", msg, nil)
	default:
		fmt.Fprintln(os.Stdout, msg)
	}
}

// Warn logs a startup warning. Always printed.
func Warn(msg string) {
	switch format {
	case JSON:
		writeJSON("warn", msg, nil)
	default:
		fmt.Fprintf(os.Stdout, "  \033[33mwarn\033[0m  %s\n", msg)
	}
}

// Error logs an error. Always printed.
func Error(msg string) {
	switch format {
	case JSON:
		writeJSON("error", msg, nil)
	default:
		fmt.Fprintf(os.Stderr, "  \033[31merror\033[0m %s\n", msg)
	}
}

// Request logs a per-request line. Suppressed in silent mode.
func Request(status int, method, path string, dur time.Duration) {
	if format == Silent {
		return
	}
	ms := dur.Milliseconds()
	switch format {
	case JSON:
		writeJSON("request", "", map[string]interface{}{
			"status": status, "method": method, "path": path, "ms": ms,
		})
	default:
		color := statusColor(status)
		fmt.Printf("  %s%d\033[0m  %-7s %-30s %dms\n", color, status, method, path, ms)
	}
}

func statusColor(status int) string {
	switch {
	case status < 300:
		return "\033[32m" // green
	case status < 400:
		return "\033[34m" // blue
	case status < 500:
		return "\033[33m" // yellow
	default:
		return "\033[31m" // red
	}
}

func writeJSON(level, msg string, extra map[string]interface{}) {
	entry := map[string]interface{}{
		"time":  time.Now().UTC().Format(time.RFC3339),
		"level": level,
	}
	if msg != "" {
		entry["msg"] = msg
	}
	for k, v := range extra {
		entry[k] = v
	}
	b, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stdout, string(b))
}
