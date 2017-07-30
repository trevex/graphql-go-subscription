package pubsub

import (
	"errors"
	"github.com/trevex/graphql-go-subscription"
)

type operation int

const (
	SUBSCRIBE operation = iota
	UNSUBSCRIBE
	PUBLISH
	SHUTDOWN
)

type command struct {
	op      operation
	topics  []string
	ch      chan interface{}
	payload interface{}
}

type Subscription struct {
	ch chan interface{}
}

type PubSub struct {
	cmds        chan command
	capacity    int
	topics      map[string]map[chan interface{}]bool
	subscribers map[chan interface{}]map[string]bool
}

func New(capacity int) *PubSub {
	ps := &PubSub{
		make(chan command),
		capacity,
		make(map[string]map[chan interface{}]bool),
		make(map[chan interface{}]map[string]bool),
	}
	go ps.run()
	return ps
}

func (ps *PubSub) Subscribe(topics ...string) (subscription.Subscription, error) {
	sub := &Subscription{
		make(chan interface{}, ps.capacity),
	}
	ps.cmds <- command{op: SUBSCRIBE, topics: topics, ch: sub.ch}
	return sub, nil
}

func (ps *PubSub) Unsubscribe(sub subscription.Subscription) error {
	if s, ok := sub.(*Subscription); ok {
		ps.cmds <- command{op: UNSUBSCRIBE, ch: s.ch}
		return nil
	}
	return errors.New("Subscription has wrong type.")
}

func (ps *PubSub) Publish(payload interface{}, topics ...string) {
	ps.cmds <- command{op: PUBLISH, payload: payload}
}

func (ps *PubSub) Shutdown() {
	ps.cmds <- command{op: SHUTDOWN}
}

func (ps *PubSub) run() {
	for cmd := range ps.cmds {
		switch cmd.op {
		case SHUTDOWN:
			break
		case SUBSCRIBE:
			for _, topic := range cmd.topics {
				ps.subscribe(topic, cmd.ch)
			}
		case UNSUBSCRIBE:
			for topic, _ := range ps.subscribers[cmd.ch] {
				ps.unsubscribe(topic, cmd.ch)
			}
		case PUBLISH:
			for _, topic := range cmd.topics {
				ps.publish(topic, cmd.payload)
			}
		}
	}
}

func (ps *PubSub) subscribe(topic string, ch chan interface{}) {
	if _, ok := ps.topics[topic]; !ok {
		ps.topics[topic] = make(map[chan interface{}]bool)
	}
	ps.topics[topic][ch] = true
	if _, ok := ps.subscribers[ch]; !ok {
		ps.subscribers[ch] = make(map[string]bool)
	}
	ps.subscribers[ch][topic] = true
}

func (ps *PubSub) unsubscribe(topic string, ch chan interface{}) {
	if _, ok := ps.topics[topic]; !ok {
		return
	}
	if _, ok := ps.topics[topic][ch]; !ok {
		return
	}
	delete(ps.topics[topic], ch)
	delete(ps.subscribers[ch], topic)
	if len(ps.topics[topic]) == 0 {
		delete(ps.topics, topic)
	}
	if len(ps.subscribers[ch]) == 0 {
		close(ch)
		delete(ps.subscribers, ch)
	}
}

func (ps *PubSub) publish(topic string, payload interface{}) {
	for ch, _ := range ps.topics[topic] {
		ch <- payload
	}
}
