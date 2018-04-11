package types

type Log struct {
	Path   string
	Format string
	UUID   string
}

type Line struct {
	LineNo int64
	Value  string
}

type Field struct {
	UUID        string
	FieldType   FieldType
	ValueType   string
	ValueString string
	ValueFloat  float64
	StartTime   int64
}

type FieldType struct {
	Name string
	Typ  string
}
