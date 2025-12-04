package main

import (
	"fmt"
	"sql-compiler/ast"
	. "sql-compiler/tokenizer"
	"strconv"
)

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) expect(tt TokenType) {
	if p.tokens[p.pos].Type != tt {
		panic("expected " + string(tt) + " but got " + string(p.tokens[p.pos].Type))
	}
	p.pos++
}
func (p *parser) expectIdent() string {
	if p.tokens[p.pos].Type != IDENT {
		panic(fmt.Sprintf("expected IDENT but got %s at %d", p.tokens[p.pos].Type, p.tokens[p.pos].Pos))
	}
	ident := p.tokens[p.pos].Literal
	p.pos++
	return ident
}
func (p *parser) expectIdentOf(ident string) {
	if p.tokens[p.pos].Type != IDENT {
		panic("expected IDENT")
	}
	if p.tokens[p.pos].Literal != ident {
		panic("expected IDENT " + ident)
	}
	p.pos++
}
func (p *parser) optionallyExpect(tt TokenType) bool {
	if !p.inrange() {
		return false
	}
	if p.tokens[p.pos].Type != tt {
		return false
	}
	p.pos++
	return true
}
func (p *parser) inrange() bool {
	return p.pos < len(p.tokens)
}
func (p *parser) parse_col_or_expr_lit() any {
	walk_back_pos := p.pos
	token := p.tokens[p.pos]
	p.pos += 1
	if token.Type == STRING {
		return token.Literal
	}
	if token.Type == TRUE || token.Type == FALSE {
		return token.Type == TRUE
	}
	if token.Type == INT {
		n, err := strconv.Atoi(token.Literal)
		if err != nil {
			panic(err)
		}
		return n
	}
	p.pos = walk_back_pos
	return p.parseCol()
}
func (p *parser) parse_simple_expr() ast.Where {
	Value1 := p.parse_col_or_expr_lit()
	operator := p.tokens[p.pos].Type
	if operator != LT && operator != GT && operator != EQ {
		panic("expected ASSIGN or LT or GT or LE or GE instead of " + string(operator))
	}
	p.pos++

	return ast.Where{
		Value1:   Value1,
		Operator: operator,
		Value2:   p.parse_col_or_expr_lit(),
	}
}
func (p *parser) parseCol() ast.Col {
	col_or_table_name := p.expectIdent()
	if p.optionallyExpect(DOT) {
		return ast.Table_access{
			Table_name: col_or_table_name,
			Col_name:   p.expectIdent(),
		}
	}
	return ast.Plain_col_name(col_or_table_name)
}
func (p *parser) parse_Select() ast.Select {
	s := ast.Select{}
	p.optionallyExpect(SELECT)
	var Value_to_select any
	for !p.optionallyExpect(FROM) {
		var alias string
		if p.optionallyExpect(LPAREN) {
			Value_to_select = p.parse_Select()
			p.expect(RPAREN)
			alias = Value_to_select.(ast.Select).Table
		} else {
			Value_to_select = p.parse_col_or_expr_lit()
		}
		if p.optionallyExpect(AS) {
			alias = p.expectIdent()
		}
		s.Selected_values = append(s.Selected_values, ast.Selected_value{Value_to_select: Value_to_select, Alias: alias})
		if !p.optionallyExpect(COMMA) {
			p.expect(FROM)
			break
		}
	}
	s.Table = p.expectIdent()
	p.expect(WHERE)
	for p.inrange() && (p.tokens[p.pos].Type != RPAREN) {
		where := p.parse_simple_expr()
		s.Wheres = append(s.Wheres, where)
		if !p.optionallyExpect(AND) {
			break
		}
	}

	return s
}
