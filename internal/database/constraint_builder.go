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
		Add(NewStatement(Quote(column)).Wrap(SQLOpenParen, SQLCloseParen).String()).
		Add(SQLGte).
		Add(NewStatement().AddInt(min).String()).
		String()
	return b
}

func (b *ConstraintBuilder) MaxLength(column string, max int) *ConstraintBuilder {
	b.check = NewStatement(FuncCharLength).
		Add(NewStatement(Quote(column)).Wrap(SQLOpenParen, SQLCloseParen).String()).
		Add(SQLLte).
		Add(NewStatement().AddInt(max).String()).
		String()
	return b
}

func (b *ConstraintBuilder) MinValue(column string, min interface{}) *ConstraintBuilder {
	b.check = NewStatement(Quote(column)).
		Add(SQLGte).
		Add(ToVal(min)).
		String()
	return b
}

func (b *ConstraintBuilder) MaxValue(column string, max interface{}) *ConstraintBuilder {
	b.check = NewStatement(Quote(column)).
		Add(SQLLte).
		Add(ToVal(max)).
		String()
	return b
}

func (b *ConstraintBuilder) HasBody() bool {
	return b.check != ""
}

func (b *ConstraintBuilder) Build() string {
	return NewStatement(SQLConstraint).
		Add(Quote(b.name)).
		Add(SQLCheck).
		Add(NewStatement(b.check).Wrap(SQLOpenParen, SQLCloseParen).String()).
		String()
}

func (b *ConstraintBuilder) BuildAdd() string {
	return NewStatement(OpAddConstraint).
		Add(Quote(b.name)).
		Add(SQLCheck).
		Add(NewStatement(b.check).Wrap(SQLOpenParen, SQLCloseParen).String()).
		String()
}

func (b *ConstraintBuilder) BuildDrop() string {
	return NewStatement(OpDropConstraint).
		Add(Quote(b.name)).
		String()
}
