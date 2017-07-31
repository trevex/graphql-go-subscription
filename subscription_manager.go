package subscription

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type FilterFunc func(ctx context.Context, payload interface{}) bool

type TriggerConfig struct {
	Options interface{}
	Filter  FilterFunc
}

type TriggerMap map[string]*TriggerConfig

var defaultTriggerConfig = TriggerConfig{
	Options: nil,
	Filter: func(ctx context.Context, payload interface{}) bool {
		return true
	},
}

type SetupFunction func(config *SubscriptionConfig, args map[string]interface{}, subscriptionName string) TriggerMap

type SetupFunctionMap map[string]SetupFunction

type SubscriptionId uint64

type SubscriptionManagerConfig struct {
	Schema         graphql.Schema
	PubSub         PubSub
	SetupFunctions SetupFunctionMap
}

type SubscriptionManager struct {
	schema         graphql.Schema
	pubsub         PubSub
	setupFunctions SetupFunctionMap
	subscriptions  map[SubscriptionId][]Subscription
	maxId          SubscriptionId
}

type SubscriptionConfig struct {
	Query          string
	Context        context.Context
	VariableValues map[string]interface{}
	OperationName  string
	Callback       func(*graphql.Result) error
}

func NewSubscriptionManager(config SubscriptionManagerConfig) *SubscriptionManager {
	sm := &SubscriptionManager{
		config.Schema,
		config.PubSub,
		config.SetupFunctions,
		make(map[SubscriptionId][]Subscription),
		0,
	}
	if sm.setupFunctions == nil {
		sm.setupFunctions = SetupFunctionMap{}
	}
	return sm
}

func (sm *SubscriptionManager) Subscribe(config SubscriptionConfig) (SubscriptionId, error) {
	if config.VariableValues == nil {
		config.VariableValues = make(map[string]interface{})
	}
	doc, err := parser.Parse(parser.ParseParams{Source: config.Query})
	if err != nil {
		return 0, fmt.Errorf("Failed to parse query: %v", err)
	}
	result := graphql.ValidateDocument(&sm.schema, doc, graphql.SpecifiedRules) // TODO: add single root subscription rule
	if !result.IsValid || len(result.Errors) > 0 {
		return 0, fmt.Errorf("Validation failed, errors: %+v", result.Errors)
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

	var triggerMap TriggerMap
	if setupFunc, ok := sm.setupFunctions[subscriptionName]; ok {
		triggerMap = setupFunc(&config, args, subscriptionName)
	} else {
		triggerMap = TriggerMap{
			subscriptionName: &defaultTriggerConfig,
		}
	}
	sm.maxId++
	subscriptionId := sm.maxId
	sm.subscriptions[subscriptionId] = []Subscription{}

	for triggerName, triggerConfig := range triggerMap {
		sub, err := sm.pubsub.Subscribe(triggerName, triggerConfig.Options, func(payload interface{}) error {
			if triggerConfig.Filter(config.Context, payload) {
				result := graphql.Execute(graphql.ExecuteParams{
					Schema:        sm.schema,
					Root:          payload,
					AST:           doc,
					OperationName: config.OperationName,
					Args:          args,
					Context:       config.Context,
				})
				err := config.Callback(result)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("Subscription of trigger %v failed, error: %v", triggerName, err)
		}
		sm.subscriptions[subscriptionId] = append(sm.subscriptions[subscriptionId], sub)
	}

	return subscriptionId, nil
}

func (sm *SubscriptionManager) Unsubscribe(id SubscriptionId) {
	for _, sub := range sm.subscriptions[id] {
		sm.pubsub.Unsubscribe(sub)
	}
	delete(sm.subscriptions, id)
}
