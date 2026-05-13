package audit

import "context"

type Event struct {
	GameRunID   *string
	ActorUserID *string
	EventType   string
	EntityType  string
	EntityID    *string
	Payload     map[string]any
}

type Logger interface {
	Record(context.Context, Event) error
}

type NoopLogger struct{}

func (NoopLogger) Record(context.Context, Event) error {
	return nil
}

type Recorder interface {
	RecordAuditEvent(context.Context, Event) error
}

type StoreLogger struct {
	recorder Recorder
}

func NewStoreLogger(recorder Recorder) StoreLogger {
	return StoreLogger{recorder: recorder}
}

func (l StoreLogger) Record(ctx context.Context, event Event) error {
	if l.recorder == nil {
		return nil
	}

	return l.recorder.RecordAuditEvent(ctx, event)
}
