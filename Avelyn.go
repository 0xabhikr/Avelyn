package main

import (
	"fmt"
	"time"
)

// Environment manages variable scope
type Environment struct {
	parent    *Environment
	variables map[string]RuntimeVal
	constants map[string]bool
}

// NewEnvironment creates a new environment
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]RuntimeVal),
		constants: make(map[string]bool),
	}
}

// DeclareVar declares a new variable
func (env *Environment) DeclareVar(varname string, value RuntimeVal, constant bool) (RuntimeVal, error) {
	if _, exists := env.variables[varname]; exists {
		return nil, fmt.Errorf("Cannot declare variable '%s' as it is already defined", varname)
	}
	env.variables[varname] = value
	if constant {
		env.constants[varname] = true
	}
	return value, nil
}

// AssignVar assigns a value to an existing variable
func (env *Environment) AssignVar(varname string, value RuntimeVal) (RuntimeVal, error) {
	resolvedEnv, err := env.resolve(varname)
	if err != nil {
		return nil, err
	}
	if resolvedEnv.constants[varname] {
		return nil, fmt.Errorf("Cannot reassign to '%s' as it is constant", varname)
	}
	resolvedEnv.variables[varname] = value
	return value, nil
}

// LookupVar retrieves a variable value
func (env *Environment) LookupVar(varname string) (RuntimeVal, error) {
	resolvedEnv, err := env.resolve(varname)
	if err != nil {
		return nil, err
	}
	if val, exists := resolvedEnv.variables[varname]; exists {
		return val, nil
	}
	return MK_NULL(), nil
}

// resolve finds the environment that contains the variable
func (env *Environment) resolve(varname string) (*Environment, error) {
	if _, exists := env.variables[varname]; exists {
		return env, nil
	}
	if env.parent == nil {
		return nil, fmt.Errorf("Cannot resolve '%s'", varname)
	}
	return env.parent.resolve(varname)
}

// CreateGlobalEnv creates the global environment with built-in functions
func CreateGlobalEnv() *Environment {
	env := NewEnvironment(nil)

	// Core constants
	env.DeclareVar("true", MK_BOOL(true), true)
	env.DeclareVar("false", MK_BOOL(false), true)
	env.DeclareVar("null", MK_NULL(), true)

	// print function - clean output
	env.DeclareVar("print", MK_NATIVE_FN(func(args []RuntimeVal, _ *Environment) (RuntimeVal, error) {
		output := ""
		for i, arg := range args {
			if i > 0 {
				output += " "
			}
			output += stringify(arg)
		}
		fmt.Println(output)
		return MK_NULL(), nil
	}), true)

	// log function - debug output
	env.DeclareVar("log", MK_NATIVE_FN(func(args []RuntimeVal, _ *Environment) (RuntimeVal, error) {
		output := ""
		for i, arg := range args {
			if i > 0 {
				output += " "
			}
			output += arg.String()
		}
		fmt.Println(output)
		return MK_NULL(), nil
	}), true)

	// gettime function - returns current time in milliseconds
	env.DeclareVar("gettime", MK_NATIVE_FN(func(_ []RuntimeVal, _ *Environment) (RuntimeVal, error) {
		return MK_NUMBER(float64(time.Now().UnixNano() / 1_000_000)), nil
	}), true)

	return env
}
