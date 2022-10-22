package domain

import "context"

// Any type implementing this interface can be used by BoardService to publish events.
// There are many possible types of adaptors for this interface:
//   - An event "broker" local to this application, that takes these events and makes them available for other modules to listen to,
//     e.g. a notification module.
//   - Publish the events to an external system like Apache Kafka or a cloud service like Google Pub/Sub.
type EventPublisher interface {
	PublishEvent(ctx context.Context, event interface{})
}

type BoardCreated struct {
	BoardId     string
	Name        string
	Description string
	CreatedTime int64
	CreatedBy   User
}

type BoardDeleted struct {
	BoardId   string
	DeletedBy User
}

// Implements EventPublisher by wrapping another EventPublisher (or nil)
// Forwards any calls to PublishEvent to the wrapped EventPublisher if it is not nil.
// This makes it possible to make event publishing optional.
type maybeEventPublisher struct {
	ep EventPublisher
}

func newMaybeEventPublisher(ep EventPublisher) EventPublisher {
	return &maybeEventPublisher{ep: ep}
}

func (m *maybeEventPublisher) PublishEvent(ctx context.Context, event interface{}) {
	if m.ep != nil {
		m.ep.PublishEvent(ctx, event)
	}
}
