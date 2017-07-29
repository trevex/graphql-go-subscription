package main

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/parser"
	// "github.com/trevex/graphql-go-subscription"
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
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:        rootQuery,
	Subscription: rootSubscription,
})

var ps = pubsub.New(4)

// var subscriptionManager = subscriptions.NewSubscriptionManager(schema, ps)

func main() {
	query := `
        subscription {
            newMessage
        }
    `

	// subscriptionManager.subscribe(query, func(result graphql.Result) {
	//     str, _ := json.Marshal(result)
	//     fmt.Println(str)
	// })

	// newMsg := "Hello, world!"
	// messages = append(messages, newMsg)
	// ps.publish("newMessage", newMsg)

	doc, err := parser.Parse(parser.ParseParams{Source: query})
	if err != nil {
		log.Fatalf("failed to parse query: %+v", err)
	}
	o, _ := json.Marshal(doc)
	fmt.Printf("%s \n", o)

	ps.Shutdown()
}
