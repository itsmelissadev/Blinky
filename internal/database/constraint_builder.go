package database

type ConstraintBuilder struct {
	tableName string
	name      string
	check     string
}

func NewConstraint(tableName string, name string) *ConstraintBuilder {
	return &ConstraintBuilder{
		tableName: tableName,
		name:      name,
	}
}

func (b *ConstraintBuilder) Check(expression string) *ConstraintBuilder {
	b.check = expression
	return b
}

func (b *ConstraintBuilder) MinLength(column string, min int) *ConstraintBuilder {
	b.check = NewStatement(FuncCharLength).
		Add(NewStatement(column).Wrap(SQLQuote, SQLQuote).Wrap(SQLOpenParen, SQLCloseParen).String()).
		Space().
		Add(SQLGte).
		Space().
		Add(NewStatement().AddInt(min).String()).
		String()
	return b
}

func (b *ConstraintBuilder) MaxLength(column string, max int) *ConstraintBuilder {
	b.check = NewStatement(FuncCharLength).
		Add(NewStatement(column).Wrap(SQLQuote, SQLQuote).Wrap(SQLOpenParen, SQLCloseParen).String()).
		Space().
		Add(SQLLte).
		Space().
		Add(NewStatement().AddInt(max).String()).
		String()
	return b
}

func (b *ConstraintBuilder) MinValue(column string, min interface{}) *ConstraintBuilder {
	b.check = NewStatement(column).
		Wrap(SQLQuote, SQLQuote).
		Space().
		Add(SQLGte).
		Space().
		Add(ToVal(min)).
		String()
	return b
}

func (b *ConstraintBuilder) MaxValue(column string, max interface{}) *ConstraintBuilder {
	b.check = NewStatement(column).
		Wrap(SQLQuote, SQLQuote).
		Space().
		Add(SQLLte).
		Space().
		Add(ToVal(max)).
		String()
	return b
}

func (b *ConstraintBuilder) HasBody() bool {
	return b.check != ""
}

func (b *ConstraintBuilder) Build() string {
	return NewStatement(SQLConstraint).
		Space().
		Add(SQLQuote + b.name + SQLQuote).
		Space().
		Add(SQLCheck).
		Space().
		Add(NewStatement(b.check).Wrap(SQLOpenParen, SQLCloseParen).String()).
		String()
}

func (b *ConstraintBuilder) BuildAdd() string {
	return NewStatement(OpAddConstraint).
		Space().
		Add(SQLQuote + b.name + SQLQuote).
		Space().
		Add(SQLCheck).
		Space().
		Add(NewStatement(b.check).Wrap(SQLOpenParen, SQLCloseParen).String()).
		String()
}

func (b *ConstraintBuilder) BuildDrop() string {
	return NewStatement(OpDropConstraint).
		Space().
		Add(SQLQuote + b.name + SQLQuote).
		String()
}
