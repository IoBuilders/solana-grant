package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCore_BasicInstance(t *testing.T) {

	var buf bytes.Buffer

	//slogger impl
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	sLogger := slog.New(handler)
	slog.SetDefault(sLogger)

	//custom instance
	myLogger := NewSlogLogger(sLogger)
	SetDefaultLogger(myLogger)

	infoMessage := "This is an example log message"
	Info(infoMessage, "key_a", "value_a", "key_b", "value_b")
	output := buf.String()
	assert.Contains(t, output, infoMessage)
	assert.Contains(t, output, "\"key_a\":\"value_a\"")
	assert.Contains(t, output, "\"key_b\":\"value_b\"")

	//just call all methods.
	Debug(infoMessage, "key_a", "value_a", "key_b", "value_b")
	Warn(infoMessage, "key_a", "value_a", "key_b", "value_b")
	Error(infoMessage, "key_a", "value_a", "key_b", "value_b")

	//with ctx
	ctx := context.Background()
	InfoWithCtx(ctx, infoMessage, "key_a", "value_a", "key_b", "value_b")
	ctxOutput := buf.String()
	assert.Contains(t, ctxOutput, infoMessage)
	assert.Contains(t, ctxOutput, "\"key_a\":\"value_a\"")
	assert.Contains(t, ctxOutput, "\"key_b\":\"value_b\"")

	DebugWithCtx(ctx, infoMessage, "key_a", "value_a", "key_b", "value_b")
	WarnWithCtx(ctx, infoMessage, "key_a", "value_a", "key_b", "value_b")
	ErrorWithCtx(ctx, infoMessage, "key_a", "value_a", "key_b", "value_b")

}

func TestCore_InvalidKeyPair(t *testing.T) {

	var buf bytes.Buffer

	//slogger impl
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	sLogger := slog.New(handler)
	slog.SetDefault(sLogger)

	//custom instance
	myLogger := NewSlogLogger(sLogger)
	SetDefaultLogger(myLogger)

	infoMessage := "This is an example log message"
	Info(infoMessage, "key_a", "value_a", "key_b")
	output := buf.String()

	assert.Contains(t, output, infoMessage)
	assert.Contains(t, output, "\"!BADKEY\":\"key_b\"")

}
