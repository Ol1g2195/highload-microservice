package worker

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func newTestLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	return l
}

func TestPool_StartStop(t *testing.T) {
	p := NewPool(3, newTestLogger())

	done := make(chan struct{})
	var executed int32

	p.Start()

	p.AddJob(func() {
		atomic.AddInt32(&executed, 1)
		close(done)
	})

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("job was not executed in time")
	}

	p.Stop()

	if c := atomic.LoadInt32(&executed); c != 1 {
		t.Fatalf("expected 1 executed job, got %d", c)
	}
}

func TestPool_QueueOverflow(t *testing.T) {
	p := NewPool(1, newTestLogger())
	p.Start()
	defer p.Stop()

	// Fill internal buffer (capacity 100 in implementation). We won't assert logs, just ensure no panic.
	for i := 0; i < 150; i++ {
		p.AddJob(func() { time.Sleep(1 * time.Millisecond) })
	}
}
