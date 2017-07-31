package subscription

type Subscription interface {
}

type PubSub interface {
	Subscribe(topics string, config interface{}) (Subscription, error)
	Unsubscribe(sub Subscription) error
}
