package main

import (
	"fmt"
	"strconv"
)

// Parser converts tokens into an AST
type Parser struct {
	tokens []Token
	idx    int
}

// NewParser creates a new parser
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, idx: 0}
}

func (p *Parser) notEof() bool {
	return p.idx < len(p.tokens) && p.tokens[p.idx].Type != TokenTypeEOF
}

func (p *Parser) at() Token {
	if p.idx >= len(p.tokens) {
		return Token{"", TokenTypeEOF, 0}
	}
	return p.tokens[p.idx]
}

func (p *Parser) eat() Token {
	token := p.at()
	p.idx++
	return token
}

func (p *Parser) expect(tokenType TokenType, err string) Token {
	if p.idx >= len(p.tokens) {
		fmt.Printf("Parser Error: Unexpected end of input. Expected %v\n", tokenType)
		panic("Parse error: " + err)
	}
	token := p.eat()
	if token.Type != tokenType {
		fmt.Printf("--- Parser Error ---\n")
		fmt.Printf("Line: %d\n", token.Line)
		fmt.Printf("Message: %s\n", err)
		fmt.Printf("Found: '%s' (Type: %v)\n", token.Value, token.Type)
		fmt.Printf("Expected: %v\n", tokenType)
		fmt.Printf("--------------------\n")
		panic("Parse error: " + err)
	}
	return token
}

// ProduceAST parses the token stream and returns a Program AST
func (p *Parser) ProduceAST() *Program {
	var body []Stmt
	for p.notEof() {
		body = append(body, p.parseStmt())
	}
	return &Program{Body: body}
}

func (p *Parser) parseStmt() Stmt {
	tokenType := p.at().Type

	switch tokenType {
	case TokenTypeLet, TokenTypeConst, TokenTypeVar:
		return p.parseVarDeclaration()
	case TokenTypeFn:
		return p.parseFnDeclaration()
	case TokenTypeReturn:
		return p.parseReturnStatement()
	case TokenTypeIf:
		return p.parseIfStatement()
	case TokenTypeSemicolon:
		p.eat()
		return p.parseStmt()
	default:
		expr := p.parseExpr()
		// Global semicolon eater
		for p.notEof() && p.at().Type == TokenTypeSemicolon {
			p.eat()
		}
		return expr
	}
}

func (p *Parser) parseBlockStatement() []Stmt {
	p.expect(TokenTypeOpenBrace, "Expected opening brace.")
	var body []Stmt
	for p.notEof() && p.at().Type != TokenTypeCloseBrace {
		body = append(body, p.parseStmt())
	}
	p.expect(TokenTypeCloseBrace, "Expected closing brace.")
	return body
}

func (p *Parser) parseIfStatement() Stmt {
	p.eat() // consume 'if'
	p.expect(TokenTypeOpenParen, "Expected '(' after if")
	condition := p.parseExpr()
	p.expect(TokenTypeCloseParen, "Expected ')' after condition")

	var body []Stmt
	if p.at().Type == TokenTypeOpenBrace {
		body = p.parseBlockStatement()
	} else {
		body = append(body, p.parseStmt())
	}

	var alternate Stmt
	if p.at().Type == TokenTypeElse {
		p.eat()
		if p.at().Type == TokenTypeIf {
			alternate = p.parseIfStatement()
		} else if p.at().Type == TokenTypeOpenBrace {
			alternate = &BlockStatement{Body: p.parseBlockStatement()}
		} else {
			alternate = p.parseStmt()
		}
	}

	return &IfStatement{Test: condition, Body: body, Alternate: alternate}
}

func (p *Parser) parseReturnStatement() Stmt {
	p.eat() // consume 'return'

	// Check for empty return or return before closing brace/semicolon
	if !p.notEof() || p.at().Type == TokenTypeSemicolon || p.at().Type == TokenTypeCloseBrace {
		// Consume optional semicolon after return
		if p.notEof() && p.at().Type == TokenTypeSemicolon {
			p.eat()
		}
		return &ReturnStatement{Value: nil}
	}

	value := p.parseExpr()
	// Consume optional semicolon after return value
	if p.notEof() && p.at().Type == TokenTypeSemicolon {
		p.eat()
	}
	return &ReturnStatement{Value: value}
}

func (p *Parser) parseFnDeclaration() Stmt {
	p.eat() // fn
	name := p.expect(TokenTypeIdentifier, "Expected function name").Value

	p.expect(TokenTypeOpenParen, "Expected open parenthesis")
	var params []string
	if p.at().Type != TokenTypeCloseParen {
		params = append(params, p.expect(TokenTypeIdentifier, "Expected parameter name").Value)
		for p.at().Type == TokenTypeComma {
			p.eat()
			params = append(params, p.expect(TokenTypeIdentifier, "Expected parameter name").Value)
		}
	}
	p.expect(TokenTypeCloseParen, "Missing closing parenthesis")

	body := p.parseBlockStatement()
	return &FunctionDeclaration{Name: name, Parameters: params, Body: body}
}

func (p *Parser) parseVarDeclaration() Stmt {
	keyword := p.eat().Type
	isConstant := keyword == TokenTypeConst
	identifier := p.expect(TokenTypeIdentifier, "Expected identifier name").Value

	if !p.notEof() || p.at().Type == TokenTypeSemicolon {
		if isConstant {
			panic("Constants must be initialized")
		}
		return &VarDeclaration{Constant: isConstant, Identifier: identifier, Value: nil}
	}

	p.expect(TokenTypeEquals, "Expected equals token")
	value := p.parseExpr()
	return &VarDeclaration{Constant: isConstant, Identifier: identifier, Value: value}
}

// ============ EXPRESSION PARSING ============

func (p *Parser) parseExpr() Expr {
	return p.parseAssignmentExpr()
}

func (p *Parser) parseAssignmentExpr() Expr {
	left := p.parseLogicalOrExpr()

	if p.at().Type == TokenTypeEquals {
		p.eat()
		value := p.parseAssignmentExpr()
		return &AssignmentExpr{Assignee: left, Value: value}
	}

	return left
}

func (p *Parser) parseLogicalOrExpr() Expr {
	left := p.parseLogicalAndExpr()
	for p.notEof() && p.at().Value == "||" {
		operator := p.eat().Value
		right := p.parseLogicalAndExpr()
		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}
	return left
}

func (p *Parser) parseLogicalAndExpr() Expr {
	left := p.parseEqualityExpr()
	for p.notEof() && p.at().Value == "&&" {
		operator := p.eat().Value
		right := p.parseEqualityExpr()
		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}
	return left
}

func (p *Parser) parseEqualityExpr() Expr {
	left := p.parseAdditiveExpr()
	comparisonOps := map[string]bool{"==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true}
	for p.notEof() && comparisonOps[p.at().Value] {
		operator := p.eat().Value
		right := p.parseAdditiveExpr()
		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}
	return left
}

func (p *Parser) parseAdditiveExpr() Expr {
	left := p.parseMultiplicativeExpr()
	for p.notEof() && (p.at().Value == "+" || p.at().Value == "-") {
		operator := p.eat().Value
		right := p.parseMultiplicativeExpr()
		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}
	return left
}

func (p *Parser) parseMultiplicativeExpr() Expr {
	left := p.parseCallMemberExpr()
	for p.notEof() && (p.at().Value == "/" || p.at().Value == "*" || p.at().Value == "%") {
		operator := p.eat().Value
		right := p.parseCallMemberExpr()
		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}
	return left
}

func (p *Parser) parseCallMemberExpr() Expr {
	member := p.parseMemberExpr()
	if p.at().Type == TokenTypeOpenParen {
		return p.parseCallExpr(member)
	}
	return member
}

func (p *Parser) parseCallExpr(caller Expr) Expr {
	callExpr := &CallExpr{Caller: caller, Args: p.parseArgs()}
	if p.at().Type == TokenTypeOpenParen {
		return p.parseCallExpr(callExpr)
	}
	return callExpr
}

func (p *Parser) parseArgs() []Expr {
	p.expect(TokenTypeOpenParen, "Expected open parenthesis")
	var args []Expr
	if p.at().Type != TokenTypeCloseParen {
		args = p.parseArgumentsList()
	}
	p.expect(TokenTypeCloseParen, "Missing closing parenthesis")
	return args
}

func (p *Parser) parseArgumentsList() []Expr {
	var args []Expr
	args = append(args, p.parseExpr())
	for p.at().Type == TokenTypeComma {
		p.eat()
		args = append(args, p.parseExpr())
	}
	return args
}

func (p *Parser) parseMemberExpr() Expr {
	obj := p.parsePrimaryExpr()

	for p.at().Type == TokenTypeDot || p.at().Type == TokenTypeOpenBracket {
		operator := p.eat()
		var property Expr
		computed := false

		if operator.Type == TokenTypeDot {
			computed = false
			token := p.expect(TokenTypeIdentifier, "Dot operator requires identifier")
			property = &Identifier{Symbol: token.Value}
		} else {
			computed = true
			property = p.parseExpr()
			p.expect(TokenTypeCloseBracket, "Missing closing bracket")
		}
		obj = &MemberExpr{Object: obj, Property: property, Computed: computed}
	}
	return obj
}

func (p *Parser) parseObjectExpr() Expr {
	p.eat() // {
	var properties []Property
	for p.notEof() && p.at().Type != TokenTypeCloseBrace {
		key := p.expect(TokenTypeIdentifier, "Object key expected").Value

		if p.at().Type == TokenTypeComma || p.at().Type == TokenTypeCloseBrace {
			if p.at().Type == TokenTypeComma {
				p.eat()
			}
			properties = append(properties, Property{Key: key, Value: nil})
			continue
		}

		p.expect(TokenTypeColon, "Missing colon after object key")
		value := p.parseExpr()
		properties = append(properties, Property{Key: key, Value: value})

		if p.at().Type != TokenTypeCloseBrace {
			p.expect(TokenTypeComma, "Expected comma between object properties")
		}
	}
	p.expect(TokenTypeCloseBrace, "Object missing closing brace.")
	return &ObjectLiteral{Properties: properties}
}

func (p *Parser) parsePrimaryExpr() Expr {
	tokenType := p.at().Type

	switch tokenType {
	case TokenTypeNot:
		op := p.eat().Value
		return &UnaryExpr{Operator: op, Argument: p.parsePrimaryExpr()}
	case TokenTypeBinaryOperator:
		// Handle unary minus (and plus for completeness)
		op := p.at().Value
		if op == "-" || op == "+" {
			p.eat()
			return &UnaryExpr{Operator: op, Argument: p.parsePrimaryExpr()}
		}
		token := p.at()
		panic(fmt.Sprintf("Unexpected token on line %d: '%s' (Type: %v)", token.Line, token.Value, token.Type))
	case TokenTypeString:
		return &StringLiteral{Value: p.eat().Value}
	case TokenTypeIdentifier:
		return &Identifier{Symbol: p.eat().Value}
	case TokenTypeNumber:
		val, _ := strconv.ParseFloat(p.eat().Value, 64)
		return &NumericLiteral{Value: val}
	case TokenTypeOpenBrace:
		return p.parseObjectExpr()
	case TokenTypeOpenParen:
		p.eat()
		value := p.parseExpr()
		p.expect(TokenTypeCloseParen, "Expected closing parenthesis.")
		return value
	default:
		token := p.at()
		panic(fmt.Sprintf("Unexpected token on line %d: '%s' (Type: %v)", token.Line, token.Value, token.Type))
	}
}
