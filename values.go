package main

import (
	"fmt"
	"math"
)

// ValueType represents the type of runtime value
type ValueType int

const (
	ValueTypeNull ValueType = iota
	ValueTypeNumber
	ValueTypeBoolean
	ValueTypeObject
	ValueTypeNativeFn
	ValueTypeFunction
	ValueTypeString
)

// RuntimeVal is the base interface for all runtime values
type RuntimeVal interface {
	Type() ValueType
	String() string
}

// ============ VALUE DEFINITIONS ============

// NullVal represents null
type NullVal struct{}

func (v *NullVal) Type() ValueType { return ValueTypeNull }
func (v *NullVal) String() string  { return "null" }

// BooleanVal represents a boolean value
type BooleanVal struct {
	Value bool
}

func (v *BooleanVal) Type() ValueType { return ValueTypeBoolean }
func (v *BooleanVal) String() string  { return fmt.Sprintf("%v", v.Value) }

// NumberVal represents a number value
type NumberVal struct {
	Value float64
}

func (v *NumberVal) Type() ValueType { return ValueTypeNumber }
func (v *NumberVal) String() string {
	if v.Value == math.Floor(v.Value) {
		return fmt.Sprintf("%.0f", v.Value)
	}
	return fmt.Sprintf("%v", v.Value)
}

// StringVal represents a string value
type StringVal struct {
	Value string
}

func (v *StringVal) Type() ValueType { return ValueTypeString }
func (v *StringVal) String() string  { return v.Value }

// ObjectVal represents an object/map
type ObjectVal struct {
	Properties map[string]RuntimeVal
}

func (v *ObjectVal) Type() ValueType { return ValueTypeObject }
func (v *ObjectVal) String() string {
	if len(v.Properties) == 0 {
		return "{ }"
	}
	str := "{ "
	i := 0
	for k, val := range v.Properties {
		str += fmt.Sprintf("%s: %s", k, stringify(val))
		i++
		if i < len(v.Properties) {
			str += ", "
		}
	}
	str += " }"
	return str
}

// FunctionCall represents the signature of native functions
type FunctionCall func(args []RuntimeVal, env *Environment) (RuntimeVal, error)

// NativeFnValue represents a built-in function
type NativeFnValue struct {
	Call FunctionCall
}

func (v *NativeFnValue) Type() ValueType { return ValueTypeNativeFn }
func (v *NativeFnValue) String() string  { return "[Native Function]" }

// FunctionValue represents a user-defined function
type FunctionValue struct {
	Name           string
	Parameters     []string
	DeclarationEnv *Environment
	Body           []Stmt
}

func (v *FunctionValue) Type() ValueType { return ValueTypeFunction }
func (v *FunctionValue) String() string  { return fmt.Sprintf("[Function: %s]", v.Name) }

// ============ HELPER FUNCTIONS ============

// MK_NULL creates a null value
func MK_NULL() RuntimeVal {
	return &NullVal{}
}

// MK_BOOL creates a boolean value
func MK_BOOL(b bool) RuntimeVal {
	return &BooleanVal{Value: b}
}

// MK_NUMBER creates a number value
func MK_NUMBER(n float64) RuntimeVal {
	return &NumberVal{Value: n}
}

// MK_NATIVE_FN creates a native function
func MK_NATIVE_FN(call FunctionCall) RuntimeVal {
	return &NativeFnValue{Call: call}
}

// stringify converts a runtime value to a user-friendly string
func stringify(val RuntimeVal) string {
	switch v := val.(type) {
	case *NumberVal:
		if v.Value == math.Floor(v.Value) {
			return fmt.Sprintf("%.0f", v.Value)
		}
		return fmt.Sprintf("%v", v.Value)
	case *StringVal:
		return v.Value
	case *BooleanVal:
		return fmt.Sprintf("%v", v.Value)
	case *NullVal:
		return "null"
	case *ObjectVal:
		return v.String()
	case *FunctionValue:
		return fmt.Sprintf("[Function: %s]", v.Name)
	case *NativeFnValue:
		return "[Native Function]"
	default:
		return val.String()
	}
}
