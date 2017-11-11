package mapper

//type LabelField struct {
//	Fields []string
//}
//
//func (l LabelField) Forward(data map[string]interface{}) {
//	for _, field := range l.Fields {
//		moveForLabel(field).Forward(data)
//	}
//
//}
//
//func (l LabelField) Back(data map[string]interface{}) {
//	for _, field := range l.Fields {
//		moveForLabel(field).Back(data)
//	}
//}
//
//func moveForLabel(field string) *Enum {
//	return &Enum{
//		From: field,
//		To:   "metadata/labels/io.cattle.field." + strings.ToLower(field),
//	}
//}
