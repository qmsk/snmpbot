package snmp

import (
    "fmt"
    "reflect"
)

type tableMeta struct {
    indexType   reflect.Type
    entryType   reflect.Type

    fields      []OID
}

// Use reflection on a table map to determine the necessary SNMP types and fields
func reflectTable(tableType reflect.Type) (meta tableMeta, err error) {
    meta.indexType = tableType.Key()
    meta.entryType = tableType.Elem().Elem()

    meta.fields = make([]OID, meta.entryType.NumField())

    for i := 0; i < meta.entryType.NumField(); i++ {
        field := meta.entryType.Field(i)
        snmpTag := field.Tag.Get("snmp")

        if snmpTag == "" {
            err = fmt.Errorf("Missing snmp tag for %v field %s", meta.entryType.Name(), field.Name)
            return
        }

        oid := ParseOID(snmpTag)

        // log.Printf("snmp.reflectTable: field %v:%v = %s %s\n", i, field.Name, oid, field.Type.Name())

        meta.fields[i] = oid
    }

    return
}

// Decode each OID to index the entry within the table map, and set the field to its snmp value
func loadTable(meta tableMeta, tableValue reflect.Value, varBinds []VarBind) error {
    if len(varBinds) != len(meta.fields) {
        // guaranteed by WalkTable()
        panic("snmp table row fields mismatch")
    }

    // load row
    for i, varBind := range varBinds {
        // index
        oid := OID(varBind.Name)
        fieldOid := meta.fields[i]

        oidIndex := fieldOid.Index(oid)

        if oidIndex == nil {
            // guaranteed by WalkTable()
            panic("snmp table row field mismatch")
        }

        // index
        indexValue := reflect.New(meta.indexType).Elem()
        index := indexValue.Addr().Interface().(Index)

        if err := index.setIndex(oidIndex); err != nil {
            return err
        }

        // entry
        entryValue := tableValue.MapIndex(indexValue)

        if !entryValue.IsValid() {
            entryValue = reflect.New(meta.entryType)
            tableValue.SetMapIndex(indexValue, entryValue)
        }

        // field
        fieldValue := entryValue.Elem().Field(i)
        fieldSyntax := fieldValue.Addr().Interface().(Syntax)

        if value, err := fieldSyntax.parseValue(varBind.Value); err != nil {
            return err
        } else {
            //log.Printf("snmp.Client.GetTable %v: %v[%v] = %v\n", meta.entryType.Name(), meta.entryType.Field(i).Name, index, value)

            fieldValue.Set(reflect.ValueOf(value))
        }
    }

    return nil
}

// Populate a map[IndexType]*struct{... Type `snmp:"oid"`} from SNMP
func (self *Client) GetTable(table interface{}) error {
    tableType := reflect.TypeOf(table)
    tableValue := reflect.ValueOf(table)

    tableMeta, err := reflectTable(tableType)
    if err != nil {
        return err
    }

    return self.WalkTable(tableMeta.fields, func(varBinds []VarBind) error {
        return loadTable(tableMeta, tableValue, varBinds)
    })
}

// Walk a table as supported by GetTable(); Calls the given handler func with each index and the entry struct
func WalkTable(table interface{}, handler func (string, interface{})) {
    tableValue := reflect.ValueOf(table)

    for _, mapKey := range tableValue.MapKeys() {
        index := fmt.Sprintf("%v", mapKey.Interface())
        entry := tableValue.MapIndex(mapKey).Interface()

        handler(index, entry)
    }
}
