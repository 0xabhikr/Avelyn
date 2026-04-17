package main

import (
	"unicode"
)

// TokenType represents the type of token
type TokenType int

const (
	TokenTypeNumber TokenType = iota
	TokenTypeIdentifier
	TokenTypeString
	TokenTypeLet
	TokenTypeConst
	TokenTypeFn
	TokenTypeVar
	TokenTypeReturn
	TokenTypeIf
	TokenTypeElse
	TokenTypeBinaryOperator
	TokenTypeEquals
	TokenTypeComma
	TokenTypeDot
	TokenTypeColon
	TokenTypeSemicolon
	TokenTypeAnd // &&
	TokenTypeOr  // ||
	TokenTypeNot // !
	TokenTypeOpenParen
	TokenTypeCloseParen
	TokenTypeOpenBrace
	TokenTypeCloseBrace
	TokenTypeOpenBracket
	TokenTypeCloseBracket
	TokenTypeEOF
)

// Token represents a single token
type Token struct {
	Value string
	Type  TokenType
	Line  int
}

var keywords = map[string]TokenType{
	"let":    TokenTypeLet,
	"const":  TokenTypeConst,
	"fn":     TokenTypeFn,
	"var":    TokenTypeVar,
	"return": TokenTypeReturn,
	"if":     TokenTypeIf,
	"else":   TokenTypeElse,
}

func isAlpha(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isInt(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isSkippable(ch rune) bool {
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r'
}

// Tokenize converts source code into a list of tokens
func Tokenize(sourceCode string) []Token {
	var tokens []Token
	src := []rune(sourceCode)
	line := 1

	for len(src) > 0 {
		ch := src[0]

		if ch == '\n' {
			line++
			src = src[1:]
			continue
		}

		// Handle comments
		if ch == '/' && len(src) > 1 {
			if src[1] == '/' {
				// Line comment: skip until end of line
				for len(src) > 0 && src[0] != '\n' {
					src = src[1:]
				}
				continue
			} else if src[1] == '*' {
				// Block comment: skip until */
				src = src[2:] // skip /*
				for len(src) > 1 {
					if src[0] == '*' && src[1] == '/' {
						src = src[2:] // skip */
						break
					}
					if src[0] == '\n' {
						line++
					}
					src = src[1:]
				}
				continue
			}
		}

		switch ch {
		case '(':
			tokens = append(tokens, Token{string(ch), TokenTypeOpenParen, line})
			src = src[1:]
		case ')':
			tokens = append(tokens, Token{string(ch), TokenTypeCloseParen, line})
			src = src[1:]
		case '{':
			tokens = append(tokens, Token{string(ch), TokenTypeOpenBrace, line})
			src = src[1:]
		case '}':
			tokens = append(tokens, Token{string(ch), TokenTypeCloseBrace, line})
			src = src[1:]
		case '[':
			tokens = append(tokens, Token{string(ch), TokenTypeOpenBracket, line})
			src = src[1:]
		case ']':
			tokens = append(tokens, Token{string(ch), TokenTypeCloseBracket, line})
			src = src[1:]
		case ';':
			tokens = append(tokens, Token{string(ch), TokenTypeSemicolon, line})
			src = src[1:]
		case ':':
			tokens = append(tokens, Token{string(ch), TokenTypeColon, line})
			src = src[1:]
		case ',':
			tokens = append(tokens, Token{string(ch), TokenTypeComma, line})
			src = src[1:]
		case '.':
			if len(src) > 1 && isInt(src[1]) {
				// It's a number like .5, handle below
				tokens = append(tokens, parseNumber(&src, &line, tokens))
			} else {
				tokens = append(tokens, Token{string(ch), TokenTypeDot, line})
				src = src[1:]
			}
		case '+', '-', '*', '/', '%':
			tokens = append(tokens, Token{string(ch), TokenTypeBinaryOperator, line})
			src = src[1:]
		case '=':
			if len(src) > 1 && src[1] == '=' {
				tokens = append(tokens, Token{"==", TokenTypeBinaryOperator, line})
				src = src[2:]
			} else {
				tokens = append(tokens, Token{string(ch), TokenTypeEquals, line})
				src = src[1:]
			}
		case '!':
			if len(src) > 1 && src[1] == '=' {
				tokens = append(tokens, Token{"!=", TokenTypeBinaryOperator, line})
				src = src[2:]
			} else {
				tokens = append(tokens, Token{string(ch), TokenTypeNot, line})
				src = src[1:]
			}
		case '>', '<':
			op := string(ch)
			src = src[1:]
			if len(src) > 0 && src[0] == '=' {
				op += "="
				src = src[1:]
			}
			tokens = append(tokens, Token{op, TokenTypeBinaryOperator, line})
		case '&':
			if len(src) > 1 && src[1] == '&' {
				tokens = append(tokens, Token{"&&", TokenTypeAnd, line})
				src = src[2:]
			} else {
				panic("Single '&' is not supported. Use '&&'")
			}
		case '|':
			if len(src) > 1 && src[1] == '|' {
				tokens = append(tokens, Token{"||", TokenTypeOr, line})
				src = src[2:]
			} else {
				panic("Single '|' is not supported. Use '||'")
			}
		case '"':
			src = src[1:]
			str := ""
			for len(src) > 0 && src[0] != '"' {
				// Handle escape sequences
				if src[0] == '\\' && len(src) > 1 {
					src = src[1:]
					switch src[0] {
					case 'n':
						str += "\n"
					case 't':
						str += "\t"
					case 'r':
						str += "\r"
					case '\\':
						str += "\\"
					case '"':
						str += "\""
					default:
						// Unknown escape, keep as is
						str += string(src[0])
					}
					src = src[1:]
				} else {
					if src[0] == '\n' {
						line++
					}
					str += string(src[0])
					src = src[1:]
				}
			}
			if len(src) > 0 {
				src = src[1:]
			}
			tokens = append(tokens, Token{str, TokenTypeString, line})
		default:
			if isInt(ch) {
				tokens = append(tokens, parseNumber(&src, &line, tokens))
			} else if isAlpha(ch) {
				ident := ""
				for len(src) > 0 && (isAlpha(src[0]) || isInt(src[0])) {
					ident += string(src[0])
					src = src[1:]
				}
				tokenType := TokenTypeIdentifier
				if t, ok := keywords[ident]; ok {
					tokenType = t
				}
				tokens = append(tokens, Token{ident, tokenType, line})
			} else if isSkippable(ch) {
				src = src[1:]
			} else {
				panic("Unrecognized character: " + string(ch))
			}
		}
	}

	tokens = append(tokens, Token{"EndOfFile", TokenTypeEOF, line})
	return tokens
}

func parseNumber(src *[]rune, line *int, tokens []Token) Token {
	num := ""
	dotCount := 0
	for len(*src) > 0 && (isInt((*src)[0]) || (*src)[0] == '.') {
		if (*src)[0] == '.' {
			dotCount++
			if dotCount > 1 {
				break
			}
		}
		num += string((*src)[0])
		*src = (*src)[1:]
	}
	return Token{num, TokenTypeNumber, *line}
}
