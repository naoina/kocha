package log

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// Formatter is an interface that formatter for a log entry.
type Formatter interface {
	// Format formats a log entry.
	// Format writes formatted entry to the w.
	Format(w io.Writer, entry *Entry) error
}

// RawFormatter is a formatter that doesn't format.
// RawFormatter doesn't output the almost fields of the entry except the
// Message.
type RawFormatter struct{}

// Format outputs entry.Message.
func (f *RawFormatter) Format(w io.Writer, entry *Entry) error {
	_, err := io.WriteString(w, entry.Message)
	return err
}

// LTSVFormatter is the formatter of Labeled Tab-separated Values.
// See http://ltsv.org/ for more details.
type LTSVFormatter struct {
}

// Format formats an entry to Labeled Tab-separated Values format.
func (f *LTSVFormatter) Format(w io.Writer, entry *Entry) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "level:%v", entry.Level)
	if !entry.Time.IsZero() {
		fmt.Fprintf(&buf, "\ttime:%v", entry.Time.Format(time.RFC3339Nano))
	}
	if entry.Message != "" {
		fmt.Fprintf(&buf, "\tmessage:%v", entry.Message)
	}
	for _, k := range entry.Fields.OrderedKeys() {
		fmt.Fprintf(&buf, "\t%v:%v", k, entry.Fields.Get(k))
	}
	_, err := io.Copy(w, &buf)
	return err
}
