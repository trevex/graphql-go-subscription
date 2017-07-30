package subscription

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"math"
)

// Taken from graphql-go/values.go

func isNullish(value interface{}) bool {
	if value, ok := value.(string); ok {
		return value == ""
	}
	if value, ok := value.(*string); ok {
		if value == nil {
			return true
		}
		return *value == ""
	}
	if value, ok := value.(int); ok {
		return math.IsNaN(float64(value))
	}
	if value, ok := value.(*int); ok {
		if value == nil {
			return true
		}
		return math.IsNaN(float64(*value))
	}
	if value, ok := value.(float32); ok {
		return math.IsNaN(float64(value))
	}
	if value, ok := value.(*float32); ok {
		if value == nil {
			return true
		}
		return math.IsNaN(float64(*value))
	}
	if value, ok := value.(float64); ok {
		return math.IsNaN(value)
	}
	if value, ok := value.(*float64); ok {
		if value == nil {
			return true
		}
		return math.IsNaN(*value)
	}
	return value == nil
}

func valueFromAST(valueAST ast.Value, ttype graphql.Input, variables map[string]interface{}) interface{} {

	if ttype, ok := ttype.(*graphql.NonNull); ok {
		val := valueFromAST(valueAST, ttype.OfType, variables)
		return val
	}

	if valueAST == nil {
		return nil
	}

	if valueAST, ok := valueAST.(*ast.Variable); ok && valueAST.Kind == kinds.Variable {
		if valueAST.Name == nil {
			return nil
		}
		if variables == nil {
			return nil
		}
		variableName := valueAST.Name.Value
		variableVal, ok := variables[variableName]
		if !ok {
			return nil
		}
		return variableVal
	}

	if ttype, ok := ttype.(*graphql.List); ok {
		itemType := ttype.OfType
		if valueAST, ok := valueAST.(*ast.ListValue); ok && valueAST.Kind == kinds.ListValue {
			values := []interface{}{}
			for _, itemAST := range valueAST.Values {
				v := valueFromAST(itemAST, itemType, variables)
				values = append(values, v)
			}
			return values
		}
		v := valueFromAST(valueAST, itemType, variables)
		return []interface{}{v}
	}

	if ttype, ok := ttype.(*graphql.InputObject); ok {
		valueAST, ok := valueAST.(*ast.ObjectValue)
		if !ok {
			return nil
		}
		fieldASTs := map[string]*ast.ObjectField{}
		for _, fieldAST := range valueAST.Fields {
			if fieldAST.Name == nil {
				continue
			}
			fieldName := fieldAST.Name.Value
			fieldASTs[fieldName] = fieldAST

		}
		obj := map[string]interface{}{}
		for fieldName, field := range ttype.Fields() {
			fieldAST, ok := fieldASTs[fieldName]
			if !ok || fieldAST == nil {
				continue
			}
			fieldValue := valueFromAST(fieldAST.Value, field.Type, variables)
			if isNullish(fieldValue) {
				fieldValue = field.DefaultValue
			}
			if !isNullish(fieldValue) {
				obj[fieldName] = fieldValue
			}
		}
		return obj
	}

	switch ttype := ttype.(type) {
	case *graphql.Scalar:
		parsed := ttype.ParseLiteral(valueAST)
		if !isNullish(parsed) {
			return parsed
		}
	case *graphql.Enum:
		parsed := ttype.ParseLiteral(valueAST)
		if !isNullish(parsed) {
			return parsed
		}
	}
	return nil
}

func getArgumentValues(argDefs []*graphql.Argument, argASTs []*ast.Argument, variableVariables map[string]interface{}) (map[string]interface{}, error) {
	argASTMap := map[string]*ast.Argument{}
	for _, argAST := range argASTs {
		if argAST.Name != nil {
			argASTMap[argAST.Name.Value] = argAST
		}
	}
	results := map[string]interface{}{}
	for _, argDef := range argDefs {

		name := argDef.PrivateName
		var valueAST ast.Value
		if argAST, ok := argASTMap[name]; ok {
			valueAST = argAST.Value
		}
		value := valueFromAST(valueAST, argDef.Type, variableVariables)
		if isNullish(value) {
			value = argDef.DefaultValue
		}
		if !isNullish(value) {
			results[name] = value
		}
	}
	return results, nil
}
