package mibs

import (
	"fmt"
	"reflect"
)

// Used for config loading
type SyntaxMap map[string]reflect.Type

var syntaxMap = make(SyntaxMap)

func RegisterSyntax(name string, syntax Syntax) {
	var syntaxType = reflect.TypeOf(syntax)

	syntaxMap[name] = syntaxType
}

// Returns pointer-valued interface suitable for unmarshalling
func LookupSyntax(name string) (Syntax, error) {
	if syntaxType, ok := syntaxMap[name]; !ok {
		return nil, fmt.Errorf("Unknown Syntax %v", name)
	} else {
		var syntaxValue = reflect.New(syntaxType)

		return syntaxValue.Interface().(Syntax), nil
	}
}
