package mibs

import (
	"fmt"
	"strings"
)

type IndexSyntax []*Object
type IndexValues []Value
type IndexMap map[IDKey]Value

func (indexSyntax IndexSyntax) UnpackIndex(index []int) (IndexValues, error) {
	if indexSyntax == nil {
		if len(index) == 1 && index[0] == 0 {
			return IndexValues{}, nil
		} else {
			return nil, fmt.Errorf("Unexpected leaf index: %v", index)
		}
	}

	var values = make(IndexValues, len(indexSyntax))

	for i, indexObject := range indexSyntax {
		if indexValue, indexRemaining, err := indexObject.Syntax.UnpackIndex(index); err != nil {
			return values, fmt.Errorf("Invalid index for %v: %v", indexObject, err)
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

func (indexSyntax IndexSyntax) MapIndex(index []int) (IndexMap, error) {
	var indexMap = make(IndexMap)

	for _, indexObject := range indexSyntax {
		if indexValue, indexRemaining, err := indexObject.Syntax.UnpackIndex(index); err != nil {
			return nil, fmt.Errorf("Invalid index for %v: %v", indexObject, err)
		} else {
			indexMap[indexObject.ID.Key()] = indexValue
			index = indexRemaining
		}
	}

	return indexMap, nil
}

func (indexSyntax IndexSyntax) FormatIndex(index []int) (string, error) {
	if indexValues, err := indexSyntax.UnpackIndex(index); err != nil {
		return "", err
	} else {
		return indexSyntax.FormatValues(indexValues), nil
	}
}

func (indexSyntax IndexSyntax) FormatValues(indexValues IndexValues) string {
	var indexStrings = make([]string, len(indexSyntax))

	for i, indexValue := range indexValues {
		indexStrings[i] = fmt.Sprintf("[%v]", indexValue)
	}

	return strings.Join(indexStrings, "")
}
