package service

import (
	"time"
)

type testClock struct {
	current time.Time
}

func newTestClock(start time.Time) *testClock {
	return &testClock{current: start.UTC()}
}

func (c *testClock) Now() time.Time {
	return c.current
}

func (c *testClock) Set(t time.Time) {
	c.current = t.UTC()
}

func (c *testClock) Advance(d time.Duration) {
	c.current = c.current.Add(d)
}
