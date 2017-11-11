package mapper

import "strings"

type Copy struct {
	From, To string
}

func (c Copy) Forward(data map[string]interface{}) {
	val, ok := GetValue(data, strings.Split(c.From, "/")...)
	if !ok {
		return
	}

	PutValue(data, val, strings.Split(c.To, "/")...)
}

func (c Copy) Back(data map[string]interface{}) {
	val, ok := GetValue(data, strings.Split(c.To, "/")...)
	if !ok {
		return
	}

	PutValue(data, val, strings.Split(c.From, "/")...)
}
