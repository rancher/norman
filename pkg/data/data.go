package data

import (
	"github.com/rancher/norman/pkg/types/convert"
	"github.com/rancher/norman/pkg/types/values"
)

type Object map[string]interface{}

func New() Object {
	return map[string]interface{}{}
}

func (o Object) Map(names ...string) Object {
	v := values.GetValueN(o, names...)
	m := convert.ToMapInterface(v)
	return Object(m)
}

func (o Object) String(names ...string) string {
	v := values.GetValueN(o, names...)
	return convert.ToString(v)
}
