package mibs

import (
	"fmt"
	"strings"
)

type TableIndexSyntax struct {
	Name string
	IndexSyntax
}

type Table struct {
	*ID

	IndexSyntax []TableIndexSyntax
	Entry       []*Object
}

func (table *Table) RegisterObject(id ID, objectBase Object) *Object {
	var object = table.MIB.RegisterObject(id, objectBase)

	object.Table = table

	return object
}

func (table *Table) UnpackIndex(index []int) ([]Value, error) {
	var values = make([]Value, len(table.IndexSyntax))

	for i, tableIndex := range table.IndexSyntax {
		if indexValue, indexRemaining, err := tableIndex.UnpackIndex(index); err != nil {
			return nil, fmt.Errorf("Invalid index for %v: %v", tableIndex.Name, err)
		} else {
			values[i] = indexValue
			index = indexRemaining
		}
	}

	if len(index) > 0 {
		return values, fmt.Errorf("Trailing index values: %v", index)
	}

	return values, nil
}

func (table *Table) FormatIndex(index []int) (string, error) {
	var indexStrings = make([]string, len(table.IndexSyntax))

	indexValues, err := table.UnpackIndex(index)
	if err != nil {
		return "", err
	}

	for i, indexValue := range indexValues {
		indexStrings[i] = fmt.Sprintf("[%v]", indexValue)
	}

	return strings.Join(indexStrings, ""), nil
}
