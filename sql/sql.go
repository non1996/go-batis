package sql

import (
	"fmt"
	"strings"

	"github.com/non1996/go-jsonobj/container"
	"github.com/non1996/go-jsonobj/stream"

	"github.com/non1996/gobatis/errors"
)

// Elem sql元素，比如简单的sql片段、动态sql标签、sql引用等
type Elem interface {
	Evaluate(ctx *Context) (statement *Statement, err error)
}

// ConditionElem 动态sql标签
type ConditionElem interface {
	Elem
	Condition
}

// SQL sql定义，即select/update/insert/delete/sql中定义的可被引用的语句
type SQL interface {
	Elem
}

func Frag(stmt string) Elem {
	var (
		props  []string
		params []string
	)

	stmt, props = translateProps(stmt)
	stmt, params = translateParams(stmt)

	if len(props) == 0 && len(params) == 0 {
		return &pure{Stmt: stmt}
	}

	return &fragment{
		stmt:       stmt,
		properties: props,
		parameters: params,
	}
}

func Include(id string, idFromProp bool, props Parameters) Elem {
	if props == nil {
		props = emptyParam
	}

	return &_include{
		ID:         id,
		IDFromProp: idFromProp,
		Props:      props,
	}
}

func If(condition Condition, children ...any) ConditionElem {
	return &_if{
		Condition: condition,
		Children:  anySliceToElemSlice(children),
	}
}

func When(condition Condition, children ...any) ConditionElem {
	return If(condition, children...)
}

func OtherWise(children ...any) ConditionElem {
	return If(True(), children...)
}

func Choose(children ...ConditionElem) Elem {
	return &choose{
		Children: children,
	}
}

func Where(children ...Elem) Elem {
	return Trim("WHERE", []string{"AND", "OR"}, nil, children...)
}

func Set(children ...Elem) Elem {
	return Trim("SET", nil, []string{","}, children...)
}

func Trim(
	prefix string,
	prefixOverrides []string,
	suffixOverrides []string,
	children ...Elem,
) Elem {
	return &trim{
		Prefix:          prefix,
		PrefixOverrides: prefixOverrides,
		SuffixOverrides: suffixOverrides,
		Children: stream.Map(children, func(e Elem) ConditionElem {
			if ce, ok := e.(ConditionElem); ok {
				return ce
			}
			return If(True(), e)
		}),
	}
}

func Composite(children ...any) Elem {
	return &composite{
		Children: anySliceToElemSlice(children),
	}
}

var (
	_ Elem = (*pure)(nil)
	_ Elem = (*fragment)(nil)
	_ Elem = (*_include)(nil)
	_ Elem = (*_if)(nil)
	_ Elem = (*choose)(nil)
	_ Elem = (*trim)(nil)
)

// pure 纯文本sql，不用做任何处理
type pure struct {
	Stmt string
}

func (s *pure) Evaluate(ctx *Context) (statement *Statement, err error) {
	return NewStatement(s.Stmt, nil), nil
}

// fragment 带参数或属性的sql片段
type fragment struct {
	stmt       string
	properties []string
	parameters []string
}

func (s *fragment) Evaluate(ctx *Context) (statement *Statement, err error) {
	var stmt string
	var args []string

	stmt = s.stmt
	stmt, err = s.evaluateProps(ctx, stmt)
	if err != nil {
		return nil, err
	}
	stmt, args, err = s.evaluateParams(ctx, stmt)
	if err != nil {
		return nil, err
	}

	return NewStatement(stmt, args), nil
}

func (s *fragment) evaluateProps(ctx *Context, stmt string) (string, error) {
	for idx, prop := range s.properties {
		exist := ctx.params.Exist(prop)
		if !exist {
			return "", errors.MissingParameter(prop)
		}
		value := ctx.params.Get(prop)
		stmt = strings.ReplaceAll(stmt, fmt.Sprintf("${%d}", idx), String(value))
	}

	return stmt, nil
}

func (s *fragment) evaluateParams(ctx *Context, stmt string) (string, []string, error) {
	for idx, param := range s.parameters {
		if ctx.named {
			stmt = strings.Replace(stmt, fmt.Sprintf("${%d}", idx), ":"+param, 1)
		} else {
			exist := ctx.params.Exist(param)
			if !exist {
				return "", nil, errors.MissingParameter(param)
			}
			stmt = strings.Replace(stmt, fmt.Sprintf("#{%d}", idx), "?", 1)
		}
	}

	return stmt, s.parameters, nil
}

// _include sql中的引用标签，引用另一个sql片段，Evaluate时调用其指向片段的Evaluate方法
type _include struct {
	ID         string     // 被引用的sql片段id
	IDFromProp bool       // 标识id由外部传入的属性赋值
	Props      Parameters // 传递给被引用的sql片段的属性
}

func (s *_include) prepareID(ctx *Context) (id string, err error) {
	if !s.IDFromProp {
		return s.ID, nil
	}

	if !ctx.params.Exist(s.ID) {
		return "", errors.MissingParameter(s.ID)
	}
	return String(ctx.params.Get(s.ID)), nil
}

func (s *_include) prepareProps(ctx *Context) (Parameters, error) {
	props := MapParameters{}

	for _, k := range s.Props.Keys() {
		vs := String(s.Props.Get(k))
		if vs[0] == '$' {
			vs = vs[1:]
			if !ctx.params.Exist(vs) {
				return nil, errors.MissingParameter(vs)
			}
			props[k] = ctx.params.Get(vs)
		} else {
			props[k] = s.Props.Get(k)
		}
	}

	return props, nil
}

func (s *_include) Evaluate(ctx *Context) (statement *Statement, err error) {
	id, err := s.prepareID(ctx)
	if err != nil {
		return nil, err
	}
	props, err := s.prepareProps(ctx)
	if err != nil {
		return nil, err
	}

	refed := ctx.GetSQL(id)
	return refed.Evaluate(ctx.Next(props))
}

// _if 动态sql中的if标签，根据参数判断是否添加sql片段
type _if struct {
	Condition
	Children []Elem
}

func (s *_if) Evaluate(ctx *Context) (statement *Statement, err error) {
	satisfy, err := s.Satisfy(ctx)
	if err != nil {
		return nil, err
	}
	if !satisfy {
		return emptyStatement, nil
	}

	childStatements, err := stream.MapWithError(s.Children, func(e Elem) (*Statement, error) {
		return e.Evaluate(ctx)
	})
	if err != nil {
		return nil, err
	}

	return StatementMerge(childStatements), nil
}

type choose struct {
	Children []ConditionElem
}

func (s *choose) Evaluate(ctx *Context) (statement *Statement, err error) {
	for _, child := range s.Children {
		satisfy, err := child.Satisfy(ctx)
		if err != nil {
			return nil, err
		}
		if satisfy {
			return child.Evaluate(ctx)
		}
	}
	return emptyStatement, nil
}

// trim
// 特例 where/set
type trim struct {
	Prefix          string
	PrefixOverrides []string
	SuffixOverrides []string
	Children        []ConditionElem
}

func (s *trim) trimPrefix(statement *Statement) {
	for _, prefix := range s.PrefixOverrides {
		if strings.HasPrefix(statement.Stmt, prefix) {
			statement.Stmt, _ = strings.CutPrefix(statement.Stmt, prefix)
			statement.Stmt = strings.TrimSpace(statement.Stmt)
			break
		}
	}
}

func (s *trim) trimSuffix(statement *Statement) {
	for _, suffix := range s.SuffixOverrides {
		if strings.HasSuffix(statement.Stmt, suffix) {
			statement.Stmt, _ = strings.CutSuffix(statement.Stmt, suffix)
			statement.Stmt = strings.TrimSpace(statement.Stmt)
			break
		}
	}
}

func (s *trim) Evaluate(ctx *Context) (statement *Statement, err error) {
	var childStatements []*Statement
	for _, child := range s.Children {
		satisfy, err := child.Satisfy(ctx)
		if err != nil {
			return nil, err
		}
		if !satisfy {
			continue
		}

		childStatement, err := child.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		childStatements = append(childStatements, childStatement)
	}

	if len(childStatements) == 0 {
		return emptyStatement, nil
	}

	s.trimPrefix(container.SliceGetFirst(childStatements))
	s.trimSuffix(container.SliceGetLast(childStatements))
	return StatementMerge(childStatements, s.Prefix), nil
}

// composite 复合 sql
type composite struct {
	Children []Elem
}

func (s *composite) Evaluate(ctx *Context) (statement *Statement, err error) {
	childStatements, err := stream.MapWithError(s.Children, func(e Elem) (*Statement, error) {
		return e.Evaluate(ctx)
	})
	if err != nil {
		return nil, err
	}
	return StatementMerge(childStatements), nil
}

func String(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func anySliceToElemSlice(elem []any) []Elem {
	if len(elem) == 0 {
		return nil
	}

	res := make([]Elem, 0, len(elem))

	for _, e := range elem {
		if s, ok := e.(string); ok {
			res = append(res, Frag(s))
		} else {
			res = append(res, e.(Elem))
		}
	}

	return res
}
