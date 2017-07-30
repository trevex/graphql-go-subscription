package subscription

type Subscription interface {
}

type PubSub interface {
	Subscribe(topics ...string) (Subscription, error)
	Unsubscribe(sub Subscription) error
}
