package panicinfo

import (
	"fmt"
	"runtime"
	"strings"
)

const PANIC_RECOVERED_TAG = "[PANIC-RECOVERED]"

type PanicInfo struct {
	File       string
	Line       int
	Message    string
	StackTrace []StackFrame
}

type StackFrame struct {
	File     string
	Line     int
	Function string
}

func NewPanicInfo(r any) PanicInfo {
	file := "unknown"
	line := 0

	executionStack := make([]uintptr, 64)
	n := runtime.Callers(0, executionStack)

	var frames []StackFrame
	if n > 0 {
		runtimeFrames := runtime.CallersFrames(executionStack[:n])

		var allFrames []runtime.Frame
		for {
			f, more := runtimeFrames.Next()
			allFrames = append(allFrames, f)
			if !more {
				break
			}
		}

		panicIdx := findPanicOriginIndex(allFrames)

		first := true
		for i := panicIdx; i < len(allFrames); i++ {
			f := allFrames[i]

			if isInternalFrame(f.Function) {
				continue
			}

			if first {
				file = f.File
				line = f.Line
				first = false
			}

			frames = append(frames, StackFrame{
				File:     f.File,
				Line:     f.Line,
				Function: f.Function,
			})
		}
	}

	return PanicInfo{
		File:       file,
		Line:       line,
		Message:    fmt.Sprintf("%v", r),
		StackTrace: frames,
	}
}

func findPanicOriginIndex(frames []runtime.Frame) int {
	for i, f := range frames {
		if strings.Contains(f.Function, "runtime.gopanic") {
			return i + 1
		}
	}
	return 0
}

func (p PanicInfo) FormatStackTrace() string {
	var sb strings.Builder
	for i, f := range p.StackTrace {
		fmt.Fprintf(&sb, "#%d %s\n\t%s:%d\n", i, f.Function, f.File, f.Line)
	}
	return sb.String()
}

func (p PanicInfo) Error() string {
	return fmt.Sprintf("panic at %s:%d — %s", p.File, p.Line, p.Message)
}

func isInternalFrame(fn string) bool {
	defaults := []string{
		"runtime.",
		"net/http.",
	}
	for _, prefix := range defaults {
		if strings.HasPrefix(fn, prefix) {
			return true
		}
	}
	return false
}
