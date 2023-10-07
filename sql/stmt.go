package sql

import (
	"strings"
)

var emptyStatement = &Statement{}

type Statement struct {
	Stmt     string
	ArgNames []string
}

func NewStatement(
	stmt string,
	args []string,
) *Statement {
	return &Statement{
		Stmt:     strings.TrimSpace(stmt),
		ArgNames: args,
	}
}

func (s Statement) Get() (string, []string) {
	return s.Stmt, s.ArgNames
}

func (s Statement) GetStmt() string {
	return s.Stmt
}

func (s Statement) GetArgNames() []string {
	return s.ArgNames
}

func StatementMerge(statements []*Statement, prefix ...string) *Statement {
	var stmts = make([]string, 0, len(statements)+len(prefix))
	var args = make([]string, 0, len(statements))

	if len(prefix) != 0 {
		stmts = append(stmts, prefix[0])
	}

	for _, s := range statements {
		stmts = append(stmts, s.Stmt)
		args = append(args, s.ArgNames...)
	}

	return &Statement{
		Stmt:     strings.Join(stmts, " "),
		ArgNames: args,
	}
}
