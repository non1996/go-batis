package sql

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/non1996/go-batis/errors"
)

type Condition interface {
	Satisfy(ctx *Context) (satisfy bool, err error)
}

type ConstCondition struct {
	b bool
}

func (t *ConstCondition) Satisfy(ctx *Context) (satisfy bool, err error) {
	return t.b, nil
}

const tmplCondition = `{{if %s}}t{{else}}f{{end}}`

// templateCondition 使用go template实现的条件语句
type templateCondition struct {
	tmpl *template.Template
}

func Test(cond string) Condition {
	return &templateCondition{
		tmpl: template.Must(template.New("").Parse(fmt.Sprintf(tmplCondition, cond))),
	}
}

func (t *templateCondition) Satisfy(ctx *Context) (satisfy bool, err error) {
	b := bytes.NewBuffer(make([]byte, 0, 1))
	err = t.tmpl.Execute(b, ctx.params)
	if err != nil {
		return false, errors.TmplExecute(err)
	}

	return bytes.Equal(b.Bytes(), []byte("t")), nil
}

var tc = &trueCondition{}

type trueCondition struct {
}

func True() Condition {
	return tc
}

func (c *trueCondition) Satisfy(ctx *Context) (bool, error) {
	return true, nil
}
