package domain

import "context"

// Any type implementing this interface can be used by BoardService to publish events.
// There are many possible types of adapters for this interface, e.g.:
//   - An event "broker" local to this application, that takes these events and makes them available for other components/modules to listen to,
//     e.g. a notification module.
//   - Publish the events to an external system like Apache Kafka or a cloud service like Google Pub/Sub.
//
// Implementations can import this package and use a type switch over the "event" argument.
// Instead we could have defined the interface with one method for every type of event.
type EventPublisher interface {
	PublishEvent(ctx context.Context, event interface{})
}

// TODO could add more events like InviteCreated, InviteAccepted etc.

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

/*
Implements EventPublisher by wrapping another EventPublisher or nil.
Forwards any calls to PublishEvent to the wrapped EventPublisher if it is not nil.
This is just syntactic sugar for optional event publishing.
We can safely call the PublishEvent method without having to check if the publisher is not nil.
*/
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
