package main

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/trevex/graphql-go-subscription"
	"github.com/trevex/graphql-go-subscription/examples/pubsub"
	"time"
)

// GraphQL

var messages []string

var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"messages": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
	},
})

var rootSubscription = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"newMessage": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Info.RootValue, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:        rootQuery,
	Subscription: rootSubscription,
})

var ps = pubsub.New(4)

var subscriptionManager = subscription.NewSubscriptionManager(subscription.SubscriptionManagerConfig{
	Schema: schema,
	PubSub: ps,
})

func main() {
	query := `
        subscription {
            newMessage
        }
    `

	subId, _ := subscriptionManager.Subscribe(subscription.SubscriptionConfig{
		Query: query,
		Callback: func(result *graphql.Result) error {
			str, _ := json.Marshal(result)
			fmt.Printf("%s", str)
			return nil
		},
	})

	// Add a new message
	newMsg := "Hello, world!"
	// To the store
	messages = append(messages, newMsg)
	// And additionally publish it as well
	ps.Publish("newMessage", newMsg)

	// Dirty way to wait for goroutines
	time.Sleep(2 * time.Second)

	// To clean up a subscription unsubscribe
	subscriptionManager.Unsubscribe(subId)

	// Shutdown pubsub routines
	ps.Shutdown()
}
