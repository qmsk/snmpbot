package mibs

import (
  "fmt"
  "strings"
)

type IndexSyntax []*Object
type IndexMap map[IDKey]Value

func (indexSyntax IndexSyntax) UnpackIndex(index []int) ([]Value, error) {
	var values = make([]Value, len(indexSyntax))

	for i, indexObject := range indexSyntax {
		if indexValue, indexRemaining, err := indexObject.Syntax.UnpackIndex(index); err != nil {
			return nil, fmt.Errorf("Invalid index for %v: %v", indexValue, err)
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
			return nil, fmt.Errorf("Invalid index for %v: %v", indexValue, err)
		} else {
			indexMap[indexObject.ID.Key()] = indexValue
			index = indexRemaining
		}
	}

  return indexMap, nil
}

func (indexSyntax IndexSyntax) FormatIndex(index []int) (string, error) {
	var indexStrings = make([]string, len(indexSyntax))

	indexValues, err := indexSyntax.UnpackIndex(index)
	if err != nil {
		return "", err
	}

	for i, indexValue := range indexValues {
		indexStrings[i] = fmt.Sprintf("[%v]", indexValue)
	}

	return strings.Join(indexStrings, ""), nil
}
