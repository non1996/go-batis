package implement

import (
	"context"
)

type Session context.Context

func SetReader(session Session, r Reader) Session {
	return context.WithValue(session, txKey, r)
}

func SetWriter(session Session, w Writer) Session {
	return context.WithValue(session, txKey, w)
}

func Set(session Session, rw ReaderWriter) Session {
	return context.WithValue(session, txKey, rw)
}

func getR(session Session) Reader {
	return session.Value(txKey).(Reader)
}

func getW(session Session) Writer {
	return session.Value(txKey).(Writer)
}
