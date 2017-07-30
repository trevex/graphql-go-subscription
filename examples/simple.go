package main

import (
	"encoding/json"
	"github.com/graphql-go/graphql"
	"github.com/trevex/graphql-go-subscription"
	"github.com/trevex/graphql-go-subscription/examples/pubsub"
	"log"
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
	Name: "RootSubscription",
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
	Schema: &schema,
	PubSub: ps,
})

func main() {
	query := `
        subscription {
            newMessage
        }
    `

	subscriptionManager.Subscribe(query, func(result graphql.Result) {
		str, _ := json.Marshal(result)
		log.Println(str)
	})

	// Add a new message
	newMsg := "Hello, world!"
	// To the store
	messages = append(messages, newMsg)
	// And additionally publish it as well
	ps.Publish(newMsg, "newMessage")

	// Shutdown subscription routines
	ps.Shutdown()
}
