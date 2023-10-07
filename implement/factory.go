package implement

import (
	"context"
)

type Transaction func(session Session) (err error)

type SQLSessionFactory struct {
	t Transactionable
}

func NewSQLSessionFactory(t Transactionable) *SQLSessionFactory {
	return &SQLSessionFactory{t: t}
}

func (f *SQLSessionFactory) OpenSession(
	ctx context.Context,
	transactions ...Transaction,
) error {
	if len(transactions) == 0 {
		return nil
	}

	session, err := f.t.Begin(ctx)
	if err != nil {
		return err
	}

	//session := context.WithValue(ctx, txKey, tx)

	for _, transaction := range transactions {
		if err = transaction(session); err != nil {
			return f.t.Rollback(session)
		}
	}

	return f.t.Commit(session)
}

type keyType string

const txKey = keyType("K_GOBATIS_TX")
