package broker

// ISubscription is an interface for pubsub and stream subscriptions
type ISubscription interface {
	GetTopicEventChan() chan TopicEvent
	GetTopics() []string
}

// BaseSubscription ...
type BaseSubscription struct {
	TopicEventChan chan TopicEvent
	Topics         []string
}

// TopicEvent ...
type TopicEvent struct {
	Event []byte
	Topic string
}

// GetTopicEventChan returns topic with event channel
func (baseSubscription BaseSubscription) GetTopicEventChan() chan TopicEvent {
	return baseSubscription.TopicEventChan
}

// GetTopics returns topics for pubsub or streams
func (baseSubscription BaseSubscription) GetTopics() []string {
	return baseSubscription.Topics
}
