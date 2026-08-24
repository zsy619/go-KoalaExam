package event

import (
	"context"
	"testing"
	"time"
)

func TestEventBus_Subscribe_Publish(t *testing.T) {
	bus := NewBus()
	got := make(chan Event, 1)

	bus.Subscribe("test.event", HandlerFunc(func(ctx context.Context, e Event) error {
		got <- e
		return nil
	}))

	e := &testEvent{name: "test.event", t: time.Now()}
	if err := bus.Publish(context.Background(), e); err != nil {
		t.Fatalf("Publish error: %v", err)
	}

	select {
	case received := <-got:
		if received.EventName() != "test.event" {
			t.Errorf("expected test.event, got %s", received.EventName())
		}
	case <-time.After(time.Second):
		t.Fatal("handler not called within 1s")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewBus()
	count := 0

	for i := 0; i < 3; i++ {
		bus.Subscribe("multi", HandlerFunc(func(ctx context.Context, e Event) error {
			count++
			return nil
		}))
	}

	bus.Publish(context.Background(), &testEvent{name: "multi", t: time.Now()})
	if count != 3 {
		t.Errorf("expected 3 handlers called, got %d", count)
	}
}

func TestExamStartedEvent(t *testing.T) {
	e := &ExamStartedEvent{UserID: 1, ExamID: 2, RecordID: 3, StartedAt: time.Now()}
	if e.EventName() != "exam.started" {
		t.Errorf("expected exam.started, got %s", e.EventName())
	}
	if e.OccurredAt().IsZero() {
		t.Error("OccurredAt should not be zero")
	}
}

func TestFavoriteToggledEvent(t *testing.T) {
	e := &FavoriteToggledEvent{UserID: 1, Added: true, ToggledAt: time.Now()}
	if e.EventName() != "favorite.toggled" {
		t.Errorf("expected favorite.toggled, got %s", e.EventName())
	}
}

type testEvent struct {
	name string
	t    time.Time
}

func (e *testEvent) EventName() string  { return e.name }
func (e *testEvent) OccurredAt() time.Time { return e.t }
