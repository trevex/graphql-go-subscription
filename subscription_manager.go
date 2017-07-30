package subscription

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/parser"
)

type SetupFunction func()

type SetupFunctionMap map[string]SetupFunction

type SubscriptionManagerConfig struct {
	Schema         *graphql.Schema
	PubSub         PubSub
	SetupFunctions SetupFunctionMap
}

type SubscriptionManager struct {
	schema         *graphql.Schema
	pubsub         PubSub
	setupFunctions SetupFunctionMap
}

func NewSubscriptionManager(config SubscriptionManagerConfig) *SubscriptionManager {
	sm := &SubscriptionManager{config.Schema, config.PubSub, config.SetupFunctions}
	if sm.setupFunctions == nil {
		sm.setupFunctions = SetupFunctionMap{}
	}
	return sm
}

func (sm *SubscriptionManager) Subscribe(query string, callback func(graphql.Result)) error {
	doc, err := parser.Parse(parser.ParseParams{Source: query})
	if err != nil {
		return fmt.Errorf("Failed to parse query: %v", err)
	}
	result := graphql.ValidateDocument(sm.schema, doc, graphql.SpecifiedRules) // TODO: add single root subscription rule
	if !result.IsValid || len(result.Errors) > 0 {
		return fmt.Errorf("Validation failed, errors: %+v", result.Errors)
	}
	o, _ := json.Marshal(result)
	fmt.Printf("%s \n", o)
	return nil
}
