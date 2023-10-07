package implement

import (
	"context"
	"database/sql"
)

type Writer interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExec(query string, arg any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
}

type Reader interface {
	Get(dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Transactionable interface {
	Begin(context.Context) (Session, error)
	Rollback(Session) error
	Commit(Session) error
}

type Statement interface {
	Prepare() (string, []any, error)
}

type NamedStatement interface {
	Prepare() (string, any, error)
}
