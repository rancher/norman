package query

var (
	COND_EQ      = ConditionType("eq")
	COND_NE      = ConditionType("ne")
	COND_NULL    = ConditionType("null")
	COND_NOTNULL = ConditionType("notnull")
	COND_IN      = ConditionType("in")
	COND_NOTIN   = ConditionType("notin")
	COND_OR      = ConditionType("or")
	COND_AND     = ConditionType("AND")

	mods = map[ConditionType]bool{
		COND_EQ:      true,
		COND_NE:      true,
		COND_NULL:    true,
		COND_NOTNULL: true,
		COND_IN:      true,
		COND_NOTIN:   true,
		COND_OR:      true,
		COND_AND:     true,
	}
)

type ConditionType string

type Condition struct {
	values        []interface{}
	conditionType ConditionType
	left, right   *Condition
}

func ValidMod(mod string) bool {
	return mods[ConditionType(mod)]
}

func NewCondition(conditionType ConditionType, values ...interface{}) *Condition {
	return &Condition{
		values:        values,
		conditionType: conditionType,
	}
}

func NE(value interface{}) *Condition {
	return NewCondition(COND_NE, value)
}

func EQ(value interface{}) *Condition {
	return NewCondition(COND_EQ, value)
}

func NULL(value interface{}) *Condition {
	return NewCondition(COND_NULL)
}

func NOTNULL(value interface{}) *Condition {
	return NewCondition(COND_NOTNULL)
}

func IN(values ...interface{}) *Condition {
	return NewCondition(COND_IN, values...)
}

func NOTIN(values ...interface{}) *Condition {
	return NewCondition(COND_NOTIN, values...)
}

func (c *Condition) AND(right *Condition) *Condition {
	return &Condition{
		conditionType: COND_AND,
		left:          c,
		right:         right,
	}
}

func (c *Condition) OR(right *Condition) *Condition {
	return &Condition{
		conditionType: COND_OR,
		left:          c,
		right:         right,
	}
}
