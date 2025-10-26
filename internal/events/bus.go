package events

import "sync"

// Event represents a generic message passed between systems.
type Event interface{}

// Bus provides a lightweight pub/sub event system for ECS communication.
type Bus struct {
	mu   sync.RWMutex
	subs map[string][]chan Event
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		subs: make(map[string][]chan Event),
	}
}

// Subscribe registers a listener channel for a topic.
// Returned channel has a small buffer and should be read continuously.
func (b *Bus) Subscribe(topic string) <-chan Event {
	ch := make(chan Event, 8)

	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], ch)
	b.mu.Unlock()

	return ch
}

// Unsubscribe removes a specific channel from a topic.
func (b *Bus) Unsubscribe(topic string, target <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	chans := b.subs[topic]
	for i, ch := range chans {
		if ch == target {
			close(ch)
			b.subs[topic] = append(chans[:i], chans[i+1:]...)
			break
		}
	}

	// Cleanup empty topic
	if len(b.subs[topic]) == 0 {
		delete(b.subs, topic)
	}
}

// Publish sends an event to all subscribers of a topic.
// If a subscriber’s channel buffer is full, the event is dropped.
func (b *Bus) Publish(topic string, evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subs[topic] {
		select {
		case ch <- evt:
		default:
			// drop if slow subscriber
		}
	}
}

// Close gracefully closes all subscription channels and clears topics.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, chans := range b.subs {
		for _, ch := range chans {
			close(ch)
		}
	}
	b.subs = make(map[string][]chan Event)
}

