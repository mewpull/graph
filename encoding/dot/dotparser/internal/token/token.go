// Copyright ©2017 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// generated by gocc; DO NOT EDIT.

package token

import (
	"fmt"
)

type Token struct {
	Type
	Lit []byte
	Pos
}

type Type int

const (
	INVALID Type = iota
	EOF
)

type Pos struct {
	Offset int
	Line   int
	Column int
}

func (this Pos) String() string {
	return fmt.Sprintf("Pos(offset=%d, line=%d, column=%d)", this.Offset, this.Line, this.Column)
}

type TokenMap struct {
	typeMap []string
	idMap   map[string]Type
}

func (this TokenMap) Id(tok Type) string {
	if int(tok) < len(this.typeMap) {
		return this.typeMap[tok]
	}
	return "unknown"
}

func (this TokenMap) Type(tok string) Type {
	if typ, exist := this.idMap[tok]; exist {
		return typ
	}
	return INVALID
}

func (this TokenMap) TokenString(tok *Token) string {
	//TODO: refactor to print pos & token string properly
	return fmt.Sprintf("%s(%d,%s)", this.Id(tok.Type), tok.Type, tok.Lit)
}

func (this TokenMap) StringType(typ Type) string {
	return fmt.Sprintf("%s(%d)", this.Id(typ), typ)
}

var TokMap = TokenMap{
	typeMap: []string{
		"INVALID",
		"$",
		"{",
		"}",
		"empty",
		"strict",
		"graphx",
		"digraph",
		";",
		"--",
		"->",
		"node",
		"edge",
		"[",
		"]",
		",",
		"=",
		"subgraph",
		":",
		"id",
	},

	idMap: map[string]Type{
		"INVALID":  0,
		"$":        1,
		"{":        2,
		"}":        3,
		"empty":    4,
		"strict":   5,
		"graphx":   6,
		"digraph":  7,
		";":        8,
		"--":       9,
		"->":       10,
		"node":     11,
		"edge":     12,
		"[":        13,
		"]":        14,
		",":        15,
		"=":        16,
		"subgraph": 17,
		":":        18,
		"id":       19,
	},
}
