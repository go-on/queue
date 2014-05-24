package queue

import (
	"fmt"
	"io"
)

// LogErrorsTo logs errors and panics to the given io.Writer with the given prefix
//
// LogErrorsTo() is an alternative to LogDebugTo() and they should no be called both, because they are both
// changing the logging target and verbosity.
//
// If more than one logging setter is called, only the last one
// has any effect.
func (q *Queue) LogErrorsTo(logTarget io.Writer) *Queue {
	q.logTarget = logTarget
	q.logverbose = false
	return q
}

// LogDebugTo logs debugging information to the given io.Writer with the given prefix
//
// LogDebugTo() is an alternative to LogErrorsTo() and they should no be called both, because they are both
// changing the logging target and verbosity.
//
// If more than one logging setter is called, only the last one
// has any effect.
func (q *Queue) LogDebugTo(logTarget io.Writer) *Queue {
	q.logTarget = logTarget
	q.logverbose = true
	return q
}

func (q *Queue) logprefix() (prefix string) {
	if q.name != "" {
		return q.name + " - "
	}
	return
}

func (q *Queue) logPanic(format string, a ...interface{}) {
	if q.logTarget != nil {
		fmt.Fprintf(q.logTarget, "\n"+q.logprefix()+"PANIC: "+format, a...)
	}
}

func (q *Queue) logError(format string, a ...interface{}) {
	if q.logTarget != nil {
		fmt.Fprintf(q.logTarget, "\n"+q.logprefix()+"ERROR: "+format, a...)
	}
}

func (q *Queue) logDebug(format string, a ...interface{}) {
	if q.logTarget != nil && q.logverbose {
		fmt.Fprintf(q.logTarget, "\n"+q.logprefix()+"DEBUG: "+format, a...)
	}
}
