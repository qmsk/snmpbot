package mibs

type TableIndex struct {
	Name   string
	Syntax IndexSyntax
}

type Table struct {
	*ID

	Index []TableIndex
	Entry []*Object
}

func (table *Table) RegisterObject(id ID, objectBase Object) *Object {
	var object = table.MIB.RegisterObject(id, objectBase)

	object.Table = table

	return object
}
