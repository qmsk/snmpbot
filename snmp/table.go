package snmp

import (
    "fmt"
    "reflect"
)

/* MIB-based tables */
type Table struct {
    Node
    MIB             *MIB        // part of what MIB

    Index           []TableIndex
    Entry           []*Object
}

func (self Table) String() string {
    return fmt.Sprintf("%s::%s", self.MIB, self.Name)
}

type TableIndex struct {
    Name            string
    IndexSyntax     IndexSyntax
}

// Get an entire table, returning as a map
func (self *Client) GetTable(table *Table) (map[string]map[string]interface{}, error) {
    tableMap := make(map[string]map[string]interface{})

    var tableOID []OID

    for _, tableEntry := range table.Entry {
        tableOID = append(tableOID, tableEntry.OID)
    }

    err := self.WalkTable(tableOID, func(varBinds []VarBind) error {
        for i, varBind := range varBinds {
            oid := OID(varBind.Name)

            tableEntry := table.Entry[i]
            indexOID := tableEntry.OID.Index(oid)
            var index string

            for _, tableIndex := range table.Index {
                if indexLength := tableIndex.IndexSyntax.peekIndex(indexOID); indexLength <= 0 || indexLength > len(indexOID) {
                    return fmt.Errorf("invalid index for %s:%T: %v", tableIndex.Name, tableIndex.IndexSyntax, indexOID)

                } else if indexValue, err := tableIndex.IndexSyntax.parseIndex(indexOID[:indexLength]); err != nil {
                    return err

                } else {
                    indexOID = indexOID[indexLength:]
                    if index == "" {
                        index = indexValue.String()
                    } else {
                        index += "/" + indexValue.String()
                    }
                }
            }

            if _, found := tableMap[index]; !found {
                tableMap[index] = make(map[string]interface{})
            }

            if value, err := tableEntry.Syntax.parseValue(varBind.Value); err != nil {
                return err
            } else {
                tableMap[index][tableEntry.Name] = value
            }
        }

        return nil
    })

    if err != nil {
        return nil, err
    } else {
        return tableMap, nil
    }
}

// Probe for supported Tables
func (self *Client) ProbeTables(handler func(*Table)) error {
    for _, mib := range mibs {
        for _, table := range mib.tables {
            // TODO: probe...
            handler(table)
        }
    }

    return nil
}

/* Reflection-based table struct-maps */
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
        indexSyntax := indexValue.Addr().Interface().(IndexSyntax)

        if index, err := indexSyntax.parseIndex(oidIndex); err != nil {
            return err
        } else {
            indexValue.Set(reflect.ValueOf(index))
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

// Populate a map[IndexSyntax]*struct{... Syntax `snmp:"oid"`} from SNMP
func (self *Client) GetTableReflect(table interface{}) error {
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
