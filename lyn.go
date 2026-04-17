package main

import (
	"fmt"
	"os"
)

// ReturnError is used to break out of function execution
type ReturnError struct {
	Value RuntimeVal
}

func (e *ReturnError) Error() string {
	return "return"
}

// evalProgram evaluates a program (list of statements)
func evalProgram(program *Program, env *Environment) (RuntimeVal, error) {
	var lastEvaluated RuntimeVal = MK_NULL()
	for _, statement := range program.Body {
		var err error
		lastEvaluated, err = Evaluate(statement, env)
		if err != nil {
			return nil, err
		}
	}
	return lastEvaluated, nil
}

// evalVarDeclaration evaluates variable declarations
func evalVarDeclaration(declaration *VarDeclaration, env *Environment) (RuntimeVal, error) {
	var value RuntimeVal
	if declaration.Value != nil {
		var err error
		value, err = Evaluate(declaration.Value, env)
		if err != nil {
			return nil, err
		}
	} else {
		value = MK_NULL()
	}
	return env.DeclareVar(declaration.Identifier, value, declaration.Constant)
}

// evalFunctionDeclaration evaluates function declarations
func evalFunctionDeclaration(declaration *FunctionDeclaration, env *Environment) (RuntimeVal, error) {
	fn := &FunctionValue{
		Name:           declaration.Name,
		Parameters:     declaration.Parameters,
		DeclarationEnv: env,
		Body:           declaration.Body,
	}
	return env.DeclareVar(declaration.Name, fn, true)
}

// evalBlockStatement evaluates a block statement
func evalBlockStatement(block *BlockStatement, env *Environment) (RuntimeVal, error) {
	scope := NewEnvironment(env)
	var lastEvaluated RuntimeVal = MK_NULL()
	for _, statement := range block.Body {
		var err error
		lastEvaluated, err = Evaluate(statement, scope)
		if err != nil {
			return nil, err
		}
	}
	return lastEvaluated, nil
}

// evalIfStatement evaluates if statements
func evalIfStatement(declaration *IfStatement, env *Environment) (RuntimeVal, error) {
	test, err := Evaluate(declaration.Test, env)
	if err != nil {
		return nil, err
	}

	// Krisp truthiness: non-zero numbers and boolean 'true' are truthy
	truthy := isTruthy(test)

	if truthy {
		scope := NewEnvironment(env)
		var lastResult RuntimeVal = MK_NULL()
		for _, stat := range declaration.Body {
			var err error
			lastResult, err = Evaluate(stat, scope)
			if err != nil {
				return nil, err
			}
		}
		return lastResult, nil
	} else if declaration.Alternate != nil {
		return Evaluate(declaration.Alternate, env)
	}

	return MK_NULL(), nil
}

// evalReturnStatement evaluates return statements
func evalReturnStatement(stmt *ReturnStatement, env *Environment) (RuntimeVal, error) {
	var returnValue RuntimeVal
	if stmt.Value != nil {
		var err error
		returnValue, err = Evaluate(stmt.Value, env)
		if err != nil {
			return nil, err
		}
	} else {
		returnValue = MK_NULL()
	}
	return nil, &ReturnError{Value: returnValue}
}

// Evaluate is the main evaluation function
// It dispatches AST nodes to their specific evaluation functions
func Evaluate(astNode Stmt, env *Environment) (RuntimeVal, error) {
	switch node := astNode.(type) {
	// ============ LITERALS ============
	case *NumericLiteral:
		return &NumberVal{Value: node.Value}, nil

	case *StringLiteral:
		return &StringVal{Value: node.Value}, nil

	case *Identifier:
		return evalIdentifier(node, env)

	case *ObjectLiteral:
		return evalObjectExpr(node, env)

	// ============ EXPRESSIONS ============
	case *CallExpr:
		return evalCallExpr(node, env)

	case *AssignmentExpr:
		return evalAssignment(node, env)

	case *BinaryExpr:
		return evalBinaryExpr(node, env)

	case *UnaryExpr:
		return evalUnaryExpr(node, env)

	case *MemberExpr:
		return evalMemberExpr(node, env)

	// ============ STATEMENTS ============
	case *Program:
		return evalProgram(node, env)

	case *VarDeclaration:
		return evalVarDeclaration(node, env)

	case *FunctionDeclaration:
		return evalFunctionDeclaration(node, env)

	case *ReturnStatement:
		return evalReturnStatement(node, env)

	case *IfStatement:
		return evalIfStatement(node, env)

	case *BlockStatement:
		return evalBlockStatement(node, env)

	// ============ ERROR HANDLING ============
	default:
		fmt.Fprintf(os.Stderr, "Runtime Error: This AST Node has not yet been setup for interpretation.\n")
		fmt.Fprintf(os.Stderr, "Node Type: %T\n", node)
		os.Exit(1)
		return nil, fmt.Errorf("unknown node type")
	}
}
