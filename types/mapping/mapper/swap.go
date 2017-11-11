package mapper

type Swap struct {
	Left, Right string
}

func (s Swap) Forward(data map[string]interface{}) {
	rightValue, rightOk := data[s.Right]
	leftValue, leftOk := data[s.Left]
	if rightOk {
		data[s.Left] = rightValue
	}
	if leftOk {
		data[s.Right] = leftValue
	}
}

func (s Swap) Back(data map[string]interface{}) {
	s.Forward(data)
}
