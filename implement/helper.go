package implement

import (
	"database/sql"
	errors2 "errors"

	"github.com/non1996/go-jsonobj/function"
)

type Entity interface {
	SetID(int64)
}

func Insert(
	session Session,
	statement string,
	entity Entity,
) (err error) {
	res, err := getW(session).NamedExecContext(session, statement, entity)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	entity.SetID(id)
	return nil
}

func InsertWithIDSetter[T any](
	session Session,
	statement string,
	entity T,
	idSetter func(T, int64),
) (err error) {
	res, err := getW(session).NamedExecContext(session, statement, entity)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	idSetter(entity, id)
	return nil
}

func Exec(
	session Session,
	statement Statement,
) (affected int64, err error) {
	stmt, args, err := statement.Prepare()
	if err != nil {
		return 0, err
	}

	res, err := getW(session).ExecContext(session, stmt, args...)
	if err != nil {
		return 0, err
	}
	affected, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func NamedExec(
	session Session,
	statement Statement,
) (affected int64, err error) {
	stmt, arg, err := statement.Prepare()
	if err != nil {
		return 0, err
	}

	res, err := getW(session).NamedExecContext(session, stmt, arg)
	if err != nil {
		return 0, err
	}
	affected, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func List[T any](
	session Session,
	statement Statement,
) (list []T, err error) {
	stmt, args, err := statement.Prepare()
	if err != nil {
		return nil, err
	}

	err = getR(session).SelectContext(session, &list, stmt, args...)
	if NoRow(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return list, nil
}

func Get[T any](
	session Session,
	statement Statement,
) (res *T, err error) {
	stmt, args, err := statement.Prepare()
	if err != nil {
		return nil, err
	}

	res = new(T)
	err = getR(session).GetContext(session, res, stmt, args...)
	if NoRow(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GetNoPtr[T any](
	session Session,
	statement Statement,
) (res T, err error) {
	stmt, args, err := statement.Prepare()
	if err != nil {
		return function.Zero[T](), err
	}

	err = getR(session).GetContext(session, &res, stmt, args...)
	if NoRow(err) {
		return function.Zero[T](), nil
	}
	if err != nil {
		return function.Zero[T](), err
	}

	return res, nil
}

func Exist(
	session Session,
	statement Statement,
) (exist bool, err error) {
	return GetNoPtr[bool](session, statement)
}

func Count(
	session Session,
	statement Statement,
) (count int64, err error) {
	return GetNoPtr[int64](session, statement)
}

func NoRow(err error) bool {
	return errors2.Is(err, sql.ErrNoRows)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustV[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func Err[T any](_ T, err error) error {
	return err
}
