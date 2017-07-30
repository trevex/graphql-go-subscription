package subscription

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
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

type SubscriptionConfig struct {
	Query          string
	VariableValues map[string]interface{}
	Callback       func(graphql.Result)
}

func NewSubscriptionManager(config SubscriptionManagerConfig) *SubscriptionManager {
	sm := &SubscriptionManager{config.Schema, config.PubSub, config.SetupFunctions}
	if sm.setupFunctions == nil {
		sm.setupFunctions = SetupFunctionMap{}
	}
	return sm
}

func (sm *SubscriptionManager) Subscribe(config SubscriptionConfig) error {
	if config.VariableValues == nil {
		config.VariableValues = make(map[string]interface{})
	}
	doc, err := parser.Parse(parser.ParseParams{Source: config.Query})
	if err != nil {
		return fmt.Errorf("Failed to parse query: %v", err)
	}
	result := graphql.ValidateDocument(sm.schema, doc, graphql.SpecifiedRules) // TODO: add single root subscription rule
	if !result.IsValid || len(result.Errors) > 0 {
		return fmt.Errorf("Validation failed, errors: %+v", result.Errors)
	}

	var subscriptionName string
	var args map[string]interface{}
	for _, node := range doc.Definitions {
		if node.GetKind() == "OperationDefinition" {
			def, _ := node.(*ast.OperationDefinition)
			rootField, _ := def.GetSelectionSet().Selections[0].(*ast.Field)
			subscriptionName = rootField.Name.Value

			fields := sm.schema.SubscriptionType().Fields()
			args, err = getArgumentValues(fields[subscriptionName].Args, rootField.Arguments, config.VariableValues)
			break
		}
	}

	o, _ := json.Marshal(args)
	fmt.Printf("%s \n", o)
	return nil
}
