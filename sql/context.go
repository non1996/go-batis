package sql

import (
	"github.com/non1996/go-batis/errors"
)

var (
	emptyParam      = MapParameters{}
	emptyCollection = Collection{}
)

// Context 构造动态语句时的上下文
type Context struct {
	params     Parameters
	named      bool
	collection Collection
}

func NewContext() *Context {
	return &Context{
		params:     emptyParam,
		named:      false,
		collection: emptyCollection,
	}
}

func (c *Context) WithParams(params Parameters) *Context {
	c.params = params
	return c
}

func (c *Context) Named() *Context {
	c.named = true
	return c
}

func (c *Context) WithCollection(collection Collection) *Context {
	c.collection = collection
	return c
}

func (c *Context) GetSQL(id string) SQL {
	return c.collection.MustGet(id)
}

func (c *Context) Next(params Parameters) *Context {
	return &Context{
		params:     MergeParameters(c.params, params),
		named:      c.named,
		collection: c.collection,
	}
}

// Collection sql定义集合
type Collection map[string]SQL

func (c Collection) MustGet(id string) SQL {
	sql, exist := c[id]
	if !exist {
		panic(errors.MissingSQL(id))
	}
	return sql
}
