package sql

import (
	"github.com/non1996/go-jsonobj/container"
)

type Parameters interface {
	Get(key string) any
	Exist(key string) bool
	Keys() []string
}

func MergeParameters(p1, p2 Parameters) Parameters {
	newP := MapParameters{}
	for _, k := range p1.Keys() {
		newP[k] = p1.Get(k)
	}
	for _, k := range p2.Keys() {
		newP[k] = p2.Get(k)
	}

	return newP
}

type MapParameters map[string]any

func (m MapParameters) Get(key string) any {
	return container.MapGet(m, key)
}

func (m MapParameters) Exist(key string) bool {
	return container.MapExist(m, key)
}

func (m MapParameters) Keys() []string {
	return container.MapKeys(m)
}
