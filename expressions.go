package main

import (
	"fmt"
	"math"
)

// evalNumericBinaryExpr handles numeric binary operations
func evalNumericBinaryExpr(lhs *NumberVal, rhs *NumberVal, operator string) (RuntimeVal, error) {
	var result float64
	switch operator {
	case "+":
		result = lhs.Value + rhs.Value
	case "-":
		result = lhs.Value - rhs.Value
	case "*":
		result = lhs.Value * rhs.Value
	case "/":
		if rhs.Value == 0 {
			return nil, fmt.Errorf("Division by zero")
		}
		result = lhs.Value / rhs.Value
	case "%":
		result = math.Mod(lhs.Value, rhs.Value)
	default:
		return nil, fmt.Errorf("Unknown numeric operator: %s", operator)
	}
	return &NumberVal{Value: result}, nil
}

// evalBinaryExpr evaluates binary expressions
func evalBinaryExpr(binop *BinaryExpr, env *Environment) (RuntimeVal, error) {
	lhs, err := Evaluate(binop.Left, env)
	if err != nil {
		return nil, err
	}

	// Logical short-circuiting for && and ||
	if binop.Operator == "&&" {
		// Use truthiness, not just boolean
		if !isTruthy(lhs) {
			return &BooleanVal{Value: false}, nil
		}
		rhs, err := Evaluate(binop.Right, env)
		if err != nil {
			return nil, err
		}
		return &BooleanVal{Value: isTruthy(rhs)}, nil
	}

	if binop.Operator == "||" {
		// Use truthiness, not just boolean
		if isTruthy(lhs) {
			return &BooleanVal{Value: true}, nil
		}
		rhs, err := Evaluate(binop.Right, env)
		if err != nil {
			return nil, err
		}
		return &BooleanVal{Value: isTruthy(rhs)}, nil
	}

	// Evaluate RHS for other operations
	rhs, err := Evaluate(binop.Right, env)
	if err != nil {
		return nil, err
	}

	// Comparison operators
	comparisonOps := map[string]bool{"==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true}
	if comparisonOps[binop.Operator] {
		lhsNum, lhsOk := lhs.(*NumberVal)
		rhsNum, rhsOk := rhs.(*NumberVal)
		if lhsOk && rhsOk {
			var result bool
			switch binop.Operator {
			case "==":
				result = lhsNum.Value == rhsNum.Value
			case "!=":
				result = lhsNum.Value != rhsNum.Value
			case ">":
				result = lhsNum.Value > rhsNum.Value
			case "<":
				result = lhsNum.Value < rhsNum.Value
			case ">=":
				result = lhsNum.Value >= rhsNum.Value
			case "<=":
				result = lhsNum.Value <= rhsNum.Value
			}
			return &BooleanVal{Value: result}, nil
		}

		// Support equality for other types
		if binop.Operator == "==" {
			return &BooleanVal{Value: valuesEqual(lhs, rhs)}, nil
		}
		if binop.Operator == "!=" {
			return &BooleanVal{Value: !valuesEqual(lhs, rhs)}, nil
		}
	}

	// Numeric math
	if lhsNum, ok := lhs.(*NumberVal); ok {
		if rhsNum, ok := rhs.(*NumberVal); ok {
			return evalNumericBinaryExpr(lhsNum, rhsNum, binop.Operator)
		}
	}

	// String concatenation
	if binop.Operator == "+" {
		if _, ok := lhs.(*StringVal); ok {
			return &StringVal{Value: stringify(lhs) + stringify(rhs)}, nil
		}
		if _, ok := rhs.(*StringVal); ok {
			return &StringVal{Value: stringify(lhs) + stringify(rhs)}, nil
		}
	}

	return MK_NULL(), nil
}

// valuesEqual compares two runtime values for equality
func valuesEqual(lhs, rhs RuntimeVal) bool {
	switch l := lhs.(type) {
	case *NumberVal:
		if r, ok := rhs.(*NumberVal); ok {
			return l.Value == r.Value
		}
	case *BooleanVal:
		if r, ok := rhs.(*BooleanVal); ok {
			return l.Value == r.Value
		}
	case *StringVal:
		if r, ok := rhs.(*StringVal); ok {
			return l.Value == r.Value
		}
	case *NullVal:
		_, ok := rhs.(*NullVal)
		return ok
	}
	return false
}

// evalIdentifier looks up an identifier in the environment
func evalIdentifier(ident *Identifier, env *Environment) (RuntimeVal, error) {
	return env.LookupVar(ident.Symbol)
}

// evalAssignment evaluates assignment expressions
func evalAssignment(node *AssignmentExpr, env *Environment) (RuntimeVal, error) {
	// Validate assignee
	_, isIdent := node.Assignee.(*Identifier)
	_, isMember := node.Assignee.(*MemberExpr)
	if !isIdent && !isMember {
		return nil, fmt.Errorf("Invalid LHS inside assignment expr")
	}

	value, err := Evaluate(node.Value, env)
	if err != nil {
		return nil, err
	}

	// Case 1: Simple variable assignment
	if ident, ok := node.Assignee.(*Identifier); ok {
		return env.AssignVar(ident.Symbol, value)
	}

	// Case 2: Member access assignment
	if member, ok := node.Assignee.(*MemberExpr); ok {
		obj, err := Evaluate(member.Object, env)
		if err != nil {
			return nil, err
		}

		objVal, ok := obj.(*ObjectVal)
		if !ok {
			return nil, fmt.Errorf("Cannot assign to property of non-object")
		}

		var propertyName string
		if member.Computed {
			evalProp, err := Evaluate(member.Property, env)
			if err != nil {
				return nil, err
			}
			// Convert number to string for property access
			switch prop := evalProp.(type) {
			case *StringVal:
				propertyName = prop.Value
			case *NumberVal:
				propertyName = fmt.Sprintf("%v", int(prop.Value))
			default:
				return nil, fmt.Errorf("Computed property must be string or number")
			}
		} else {
			ident, ok := member.Property.(*Identifier)
			if !ok {
				return nil, fmt.Errorf("Property must be identifier")
			}
			propertyName = ident.Symbol
		}

		objVal.Properties[propertyName] = value
		return value, nil
	}

	return value, nil
}

// evalObjectExpr evaluates object literals
func evalObjectExpr(obj *ObjectLiteral, env *Environment) (RuntimeVal, error) {
	properties := make(map[string]RuntimeVal)

	for _, prop := range obj.Properties {
		var runtimeVal RuntimeVal
		if prop.Value == nil {
			var err error
			runtimeVal, err = env.LookupVar(prop.Key)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			runtimeVal, err = Evaluate(prop.Value, env)
			if err != nil {
				return nil, err
			}
		}
		properties[prop.Key] = runtimeVal
	}

	return &ObjectVal{Properties: properties}, nil
}

// evalUnaryExpr evaluates unary expressions
func evalUnaryExpr(expr *UnaryExpr, env *Environment) (RuntimeVal, error) {
	result, err := Evaluate(expr.Argument, env)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "!":
		// Truthiness flip
		isTruthy := isTruthy(result)
		return &BooleanVal{Value: !isTruthy}, nil
	case "-":
		// Unary minus - handle null gracefully
		if _, ok := result.(*NullVal); ok {
			return MK_NULL(), nil
		}
		numVal, ok := result.(*NumberVal)
		if !ok {
			return nil, fmt.Errorf("Unary minus requires a number, got %T", result)
		}
		return &NumberVal{Value: -numVal.Value}, nil
	case "+":
		// Unary plus (no-op, just return the number)
		numVal, ok := result.(*NumberVal)
		if !ok {
			return nil, fmt.Errorf("Unary plus requires a number, got %T", result)
		}
		return numVal, nil
	default:
		return nil, fmt.Errorf("Unknown unary operator: %s", expr.Operator)
	}
}

// isTruthy determines if a value is truthy
func isTruthy(val RuntimeVal) bool {
	switch v := val.(type) {
	case *BooleanVal:
		return v.Value
	case *NumberVal:
		return v.Value != 0
	case *NullVal:
		return false
	default:
		return true
	}
}

// evalCallExpr evaluates function calls
func evalCallExpr(expr *CallExpr, env *Environment) (RuntimeVal, error) {
	args := make([]RuntimeVal, 0, len(expr.Args))
	for _, arg := range expr.Args {
		val, err := Evaluate(arg, env)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}

	fn, err := Evaluate(expr.Caller, env)
	if err != nil {
		return nil, err
	}

	if nativeFn, ok := fn.(*NativeFnValue); ok {
		return nativeFn.Call(args, env)
	}

	if funcVal, ok := fn.(*FunctionValue); ok {
		scope := NewEnvironment(funcVal.DeclarationEnv)

		for i, param := range funcVal.Parameters {
			var argValue RuntimeVal
			if i < len(args) {
				argValue = args[i]
			} else {
				argValue = MK_NULL()
			}
			scope.DeclareVar(param, argValue, false)
		}

		var lastResult RuntimeVal = MK_NULL()
		for _, stmt := range funcVal.Body {
			var err error
			lastResult, err = Evaluate(stmt, scope)
			if err != nil {
				// Check if this is a return value
				if retErr, ok := err.(*ReturnError); ok {
					return retErr.Value, nil
				}
				return nil, err
			}
		}
		return lastResult, nil
	}

	return nil, fmt.Errorf("Cannot call value that is not a function: %v", fn)
}

// evalMemberExpr evaluates member access
func evalMemberExpr(node *MemberExpr, env *Environment) (RuntimeVal, error) {
	obj, err := Evaluate(node.Object, env)
	if err != nil {
		return nil, err
	}

	// Handle null gracefully
	if _, ok := obj.(*NullVal); ok {
		return MK_NULL(), nil
	}

	objVal, ok := obj.(*ObjectVal)
	if !ok {
		return nil, fmt.Errorf("[Runtime Error]: Cannot access property of a non-object")
	}

	var propertyName string
	if node.Computed {
		evalProp, err := Evaluate(node.Property, env)
		if err != nil {
			return nil, err
		}

		// Convert number to string for property access
		switch prop := evalProp.(type) {
		case *StringVal:
			propertyName = prop.Value
		case *NumberVal:
			propertyName = fmt.Sprintf("%v", int(prop.Value))
		default:
			return nil, fmt.Errorf("Computed property must be string or number")
		}
	} else {
		ident, ok := node.Property.(*Identifier)
		if !ok {
			return nil, fmt.Errorf("[Runtime Error]: Dot operator requires identifier")
		}
		propertyName = ident.Symbol
	}

	if val, exists := objVal.Properties[propertyName]; exists {
		return val, nil
	}
	return MK_NULL(), nil
}
