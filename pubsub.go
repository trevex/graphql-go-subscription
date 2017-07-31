package subscription

type Subscription interface{}

type PubSub interface {
	Subscribe(topic string, options interface{}, callback func(interface{}) error) (Subscription, error)
	Unsubscribe(sub Subscription) error
}
