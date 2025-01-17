package channel

import (
	"bytes"
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

var originalLogger = Logger

// setupLogger is a helper function to use a testable logger
func setupLogger() *bytes.Buffer {
	b := new(bytes.Buffer)
	Logger = zerolog.New(b)

	return b
}

// restoreLogger is a helper function to restore the original logger
func restoreLogger() {
	Logger = originalLogger
}

func TestWithExpiration(t *testing.T) {
	c := WithExpiration[bool](1)
	require.NotNil(t, c)
}

func TestPushWithContextSuccess(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	c.PushWithContext(ctx, 0)
	require.False(t, strings.Contains(l.String(), BlockedPush))
}

func TestPushWithTimeoutSuccess(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	c.PushWithTimeout(time.Millisecond, 0)
	require.False(t, strings.Contains(l.String(), BlockedPush))
}

func TestPushSuccess(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	c.Push(0)
	require.False(t, strings.Contains(l.String(), BlockedPush))
}

func TestPopWithContextSuccess(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	c.Push(0)
	v := c.PopWithContext(ctx)
	require.False(t, strings.Contains(l.String(), BlockedPush))
	require.Equal(t, 0, v)
}

func TestPopWithTimeoutSuccess(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	c.Push(1)
	v := c.PopWithTimeout(time.Millisecond)
	require.False(t, strings.Contains(l.String(), BlockedPush))
	require.Equal(t, 1, v)
}

func TestPopSuccess(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	c.Push(1)
	v := c.Pop()
	require.False(t, strings.Contains(l.String(), BlockedPush))
	require.Equal(t, 1, v)
}

func TestPushFail(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	c.PushWithTimeout(time.Millisecond, 0)

	go func() {
		c.PushWithTimeout(time.Millisecond, 0)
	}()

	// need a looong time on Windows to see the logs in the buffer
	time.Sleep(time.Millisecond * 100)
	require.True(t, strings.Contains(l.String(), BlockedPush))
}

func TestPopFail(t *testing.T) {
	l := setupLogger()
	defer restoreLogger()

	c := WithExpiration[int](1)

	go func() {
		c.PopWithTimeout(time.Millisecond)
	}()

	// need a looong time on Windows to see the logs in the buffer
	time.Sleep(time.Millisecond * 100)
	require.True(t, strings.Contains(l.String(), BlockedPop))
}

func TestChannel(t *testing.T) {
	c := WithExpiration[int](1)
	channel := c.Channel()

	const data = 12345
	c.Push(data)

	value := <-channel
	require.Equal(t, data, value)
}

func TestLen(t *testing.T) {
	c := WithExpiration[bool](3)
	require.Equal(t, 0, c.Len())

	c.Push(true)
	c.Push(false)
	require.Equal(t, 2, c.Len())

	c.Pop()
	require.Equal(t, 1, c.Len())

	c.Pop()
	require.Equal(t, 0, c.Len())
}
