package main

// NodeType represents the type of AST node
type NodeType int

const (
	NodeTypeProgram NodeType = iota
	NodeTypeVarDeclaration
	NodeTypeFunctionDeclaration
	NodeTypeReturnStmt
	NodeTypeAssignmentExpr
	NodeTypeMemberExpr
	NodeTypeCallExpr
	NodeTypeProperty
	NodeTypeObjectLiteral
	NodeTypeNumericLiteral
	NodeTypeStringLiteral
	NodeTypeIdentifier
	NodeTypeBinaryExpr
	NodeTypeIfStmt
	NodeTypeUnaryExpr
	NodeTypeBlock
)

// Stmt is the base interface for all statements
type Stmt interface {
	stmt()
}

// Expr is the base interface for all expressions
type Expr interface {
	Stmt
	expr()
}

// ============ STATEMENTS ============

// Program represents the root of the AST
type Program struct {
	Body []Stmt
}

func (*Program) stmt() {}

// VarDeclaration represents variable declaration (let, const, var)
type VarDeclaration struct {
	Constant   bool
	Identifier string
	Value      Expr
}

func (*VarDeclaration) stmt() {}

// FunctionDeclaration represents function declaration
type FunctionDeclaration struct {
	Name       string
	Parameters []string
	Body       []Stmt
}

func (*FunctionDeclaration) stmt() {}

// ReturnStatement represents return statement
type ReturnStatement struct {
	Value Expr
}

func (*ReturnStatement) stmt() {}

// IfStatement represents if/else statement
type IfStatement struct {
	Test      Expr
	Body      []Stmt
	Alternate Stmt
}

func (*IfStatement) stmt() {}

// BlockStatement represents a block of statements
type BlockStatement struct {
	Body []Stmt
}

func (*BlockStatement) stmt() {}

// ============ EXPRESSIONS ============

// BinaryExpr represents binary operations
type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator string
}

func (*BinaryExpr) stmt() {}
func (*BinaryExpr) expr() {}

// UnaryExpr represents unary operations
type UnaryExpr struct {
	Operator string
	Argument Expr
}

func (*UnaryExpr) stmt() {}
func (*UnaryExpr) expr() {}

// AssignmentExpr represents assignment
type AssignmentExpr struct {
	Assignee Expr
	Value    Expr
}

func (*AssignmentExpr) stmt() {}
func (*AssignmentExpr) expr() {}

// MemberExpr represents member access (obj.prop or obj[prop])
type MemberExpr struct {
	Object   Expr
	Property Expr
	Computed bool
}

func (*MemberExpr) stmt() {}
func (*MemberExpr) expr() {}

// CallExpr represents function call
type CallExpr struct {
	Caller Expr
	Args   []Expr
}

func (*CallExpr) stmt() {}
func (*CallExpr) expr() {}

// ============ LITERALS ============

// Identifier represents identifier
type Identifier struct {
	Symbol string
}

func (*Identifier) stmt() {}
func (*Identifier) expr() {}

// NumericLiteral represents numeric literal
type NumericLiteral struct {
	Value float64
}

func (*NumericLiteral) stmt() {}
func (*NumericLiteral) expr() {}

// StringLiteral represents string literal
type StringLiteral struct {
	Value string
}

func (*StringLiteral) stmt() {}
func (*StringLiteral) expr() {}

// Property represents object property
type Property struct {
	Key   string
	Value Expr
}

func (*Property) stmt() {}
func (*Property) expr() {}

// ObjectLiteral represents object literal
type ObjectLiteral struct {
	Properties []Property
}

func (*ObjectLiteral) stmt() {}
func (*ObjectLiteral) expr() {}
