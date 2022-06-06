package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type logFormatter struct {
	timeLayout string
}

var levelMap = map[logrus.Level]string{
	logrus.PanicLevel: "PNC",
	logrus.FatalLevel: "FAT",
	logrus.ErrorLevel: "ERR",
	logrus.WarnLevel:  "WAR",
	logrus.InfoLevel:  "INF",
	logrus.DebugLevel: "DBG",
	logrus.TraceLevel: "TAC",
}

func zipPath(p string) string {
	ps := strings.Split(p, "/")
	length := len(ps)
	for i := 0; i < length-1; i++ {
		ps[i] = ps[i][:1]
	}
	return strings.Join(ps, ".")
}

func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if strings.ContainsRune(entry.Message, '\n') {
		return []byte(f.formatMultiLine(entry)), nil
	}
	return []byte(fmt.Sprintf("[%s %s] %s().%d: %s\n",
			entry.Time.Format(f.timeLayout),
			levelMap[entry.Level],
			// entry.Caller.File,
			zipPath(entry.Caller.Function),
			entry.Caller.Line,
			entry.Message,
		)),
		nil
}

func (f *logFormatter) formatMultiLine(entry *logrus.Entry) string {
	prefix := fmt.Sprintf("[%s %s] %s().%d: ",
		entry.Time.Format(f.timeLayout),
		levelMap[entry.Level],
		// entry.Caller.File,
		zipPath(entry.Caller.Function),
		entry.Caller.Line,
	)
	msg := strings.Replace(entry.Message, "\n", " ↩\n"+prefix, -1) // ↴ ↩
	return prefix + msg + "\n"
}
