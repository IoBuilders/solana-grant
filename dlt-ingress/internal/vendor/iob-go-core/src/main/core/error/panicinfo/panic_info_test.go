package panicinfo

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewPanicInfo_MessageFromString(t *testing.T) {
	var info PanicInfo

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic("something went wrong")
	}()

	if info.Message != "something went wrong" {
		t.Errorf("expected message %q, got %q", "something went wrong", info.Message)
	}
}

func TestNewPanicInfo_MessageFromError(t *testing.T) {
	var info PanicInfo
	err := fmt.Errorf("database connection failed")

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic(err)
	}()

	if info.Message != "database connection failed" {
		t.Errorf("expected message %q, got %q", "database connection failed", info.Message)
	}
}

func TestNewPanicInfo_MessageFromInt(t *testing.T) {
	var info PanicInfo

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic(42)
	}()

	if info.Message != "42" {
		t.Errorf("expected message %q, got %q", "42", info.Message)
	}
}

func TestNewPanicInfo_MessageFromStruct(t *testing.T) {
	var info PanicInfo

	type customErr struct{ Code int }

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic(customErr{Code: 500})
	}()

	if info.Message == "" {
		t.Error("expected non-empty message for struct panic value")
	}
}

func TestNewPanicInfo_FileIsPopulated(t *testing.T) {
	var info PanicInfo

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic("trigger")
	}()

	if info.File == "unknown" {
		t.Error("expected File to be resolved, got \"unknown\"")
	}
	if !strings.HasSuffix(info.File, ".go") {
		t.Errorf("expected File to end in .go, got %q", info.File)
	}
}

func TestNewPanicInfo_LineIsPositive(t *testing.T) {
	var info PanicInfo

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic("trigger")
	}()

	if info.Line <= 0 {
		t.Errorf("expected Line > 0, got %d", info.Line)
	}
}

func TestPanicInfo_ErrorMethod(t *testing.T) {
	p := PanicInfo{
		File:    "/app/service/callback.go",
		Line:    42,
		Message: "nil pointer dereference",
	}

	got := p.Error()

	if !strings.Contains(got, p.File) {
		t.Errorf("Error() missing file: %q", got)
	}
	if !strings.Contains(got, "42") {
		t.Errorf("Error() missing line: %q", got)
	}
	if !strings.Contains(got, p.Message) {
		t.Errorf("Error() missing message: %q", got)
	}
}

func TestPanicInfo_ImplementsError(t *testing.T) {
	var _ error = PanicInfo{}
}

func TestNewPanicInfo_NilValue(t *testing.T) {
	var info PanicInfo

	func() {
		defer func() {
			if r := recover(); r != nil {
				info = NewPanicInfo(r)
			}
		}()
		panic("nil-like string")
	}()

	if info.Message == "" {
		t.Error("expected non-empty Message even for nil-like panic values")
	}
}
