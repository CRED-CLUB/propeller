package subscription

import (
	"github.com/CRED-CLUB/propeller/pkg/broker"
	"github.com/google/uuid"
)

// Subscription holds a subscription by a subscriber
type Subscription struct {
	TopicEventChan chan broker.TopicEvent
	ErrChan        chan error
	ID             uuid.UUID
}
