package log_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/naoina/kocha/log"
)

func TestLTSVFormatter_Format(t *testing.T) {
	now := time.Now()
	for _, v := range []struct {
		entry    *log.Entry
		expected string
	}{
		{&log.Entry{
			Level:   log.DEBUG,
			Time:    now,
			Message: "test_ltsv_log1",
			Fields: log.Fields{
				"first":  1,
				"second": "2",
				"third":  "san",
			},
		}, "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test_ltsv_log1\tfirst:1\tsecond:2\tthird:san"},
		{&log.Entry{
			Level:   log.INFO,
			Time:    now,
			Message: "test_ltsv_log2",
		}, "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test_ltsv_log2"},
		{&log.Entry{
			Level: log.WARN,
			Time:  now,
		}, "level:WARN\ttime:" + now.Format(time.RFC3339Nano)},
		{&log.Entry{
			Level: log.ERROR,
			Time:  now,
		}, "level:ERROR\ttime:" + now.Format(time.RFC3339Nano)},
		{&log.Entry{
			Level: log.FATAL,
		}, "level:FATAL"},
		{&log.Entry{}, "level:NONE"},
	} {
		var buf bytes.Buffer
		formatter := &log.LTSVFormatter{}
		if err := formatter.Format(&buf, v.entry); err != nil {
			t.Errorf(`LTSVFormatter.Format(&buf, %#v) => %#v; want %#v`, v.entry, err, nil)
		}
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`LTSVFormatter.Format(&buf, %#v); buf => %#v; want %#v`, v.entry, actual, expected)
		}
	}
}
