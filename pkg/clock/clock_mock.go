package clock

import "time"

type MockClock struct {
	Time time.Time
}

func NewMockClock() *MockClock {
	return &MockClock{}
}

func NewMockClockAt(timestamp string) *MockClock {
	dt, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		panic(err)
	}
	return &MockClock{Time: dt}
}

func (c *MockClock) NowUTC() time.Time {
	return c.Time.UTC()
}

func (c *MockClock) NowMonotonic() time.Time {
	return c.Time
}

func (c *MockClock) AdvanceSeconds(seconds int) {
	c.Time = c.Time.Add(time.Duration(seconds) * time.Second)
}
