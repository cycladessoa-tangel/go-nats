package nats

import (
	"testing"
	"time"
)

// More advanced tests on subscriptions

func TestServerAutoUnsub(t *testing.T) {
	nc := newConnection(t)
	defer nc.Close()
	received := 0
	max := 10
	sub, err := nc.Subscribe("foo", func(_, _ string, _ []byte, _ *Subscription) {
		received += 1
	})
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	sub.AutoUnsubscribe(max)
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()
	if (received != max) {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
}

func TestClientSyncAutoUnsub(t *testing.T) {
	nc := newConnection(t)
	defer nc.Close()
	received := 0
	max := 10
	sub, _ := nc.SubscribeSync("foo")
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	sub.AutoUnsubscribe(max)
	nc.Flush()
	for {
		_, err := sub.NextMsg(1*time.Millisecond)
		if err != nil { break }
		received += 1
	}
	if (received != max) {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
}

func TestClientASyncAutoUnsub(t *testing.T) {
	nc := newConnection(t)
	defer nc.Close()
	received := 0
	max := 10
	sub, err := nc.Subscribe("foo", func(_, _ string, _ []byte, _ *Subscription) {
		received += 1
	})
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	sub.AutoUnsubscribe(max)
	nc.Flush()
	if (received != max) {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
}

func TestCloseSubRelease(t *testing.T) {
	nc := newConnection(t)
	sub, _ := nc.SubscribeSync("foo")
	start := time.Now()
	go func() {
		time.Sleep(5*time.Millisecond)
		nc.Close()
	}()
	_, err := sub.NextMsg(50*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected an error from NextMsg")
	}
	elapsed := time.Since(start)
	if elapsed > 10*time.Millisecond {
		t.Fatalf("Too much time elapsed to release NextMsg: %dms",
			(elapsed/time.Millisecond))
	}
}

func TestSlowSubscriber(t *testing.T) {
	nc := newConnection(t)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	for i := 0; i < (maxChanLen + 10); i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	timeout := 500*time.Millisecond
	start := time.Now()
	nc.FlushTimeout(timeout)
	elapsed := time.Since(start)
	if elapsed >= timeout {
		t.Fatalf("Flush did not return before timeout")
	}
	// Make sure NextMsg returns an error to indicate slow consumer
	_, err := sub.NextMsg(100*time.Millisecond)
	if err == nil {
		t.Fatalf("NextMsg did not return an error")
	}
}

func TestSlowAsyncSubscriber(t *testing.T) {
	nc := newConnection(t)
	defer nc.Close()

	nc.Subscribe("foo", func(_, _ string, _ []byte, _ *Subscription) {
		time.Sleep(100*time.Second)
	})
	for i := 0; i < (maxChanLen + 10); i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	timeout := 500*time.Millisecond
	start := time.Now()
	err := nc.FlushTimeout(timeout)
	elapsed := time.Since(start)
	if elapsed >= timeout {
		t.Fatalf("Flush did not return before timeout")
	}
	if err == nil {
		t.Fatal("Expected an error indicating slow consumer")
	}
}

// FIXME Hack, make this better
func TestStopServer(t *testing.T) {
	s.stopServer()
}
