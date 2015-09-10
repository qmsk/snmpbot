package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
    "reflect"
    "strings"
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
func loadTable(meta tableMeta, tableValue reflect.Value, snmpRow []gosnmp.SnmpPDU) error {
    if len(snmpRow) != len(meta.fields) {
        panic("snmp table row fields mismatch")
    }

    // load row
    for i, snmpVar := range snmpRow {
        // index
        oid := ParseOID(snmpVar.Name)
        fieldOid := meta.fields[i]

        oidIndex := fieldOid.Index(oid)

        if oidIndex == nil {
            panic("snmp table row field mismatch")
        }

        // index
        indexValue := reflect.New(meta.indexType).Elem()
        index := indexValue.Addr().Interface().(IndexType)

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
        field := fieldValue.Addr().Interface().(Type)

        if err := field.set(snmpVar.Type, snmpVar.Value); err != nil {
            return err
        }

        // log.Printf("snmp.Client.GetTable %v: %v[%v] = %v\n", meta.entryType.Name(), meta.entryType.Field(i).Name, index, field)
    }

    return nil
}

// Walk through a table with the given list of entry field OIDs
// Call given handler with each returned row, returning any error.
func (self *Client) getTable(oids []OID, handler func ([]gosnmp.SnmpPDU) error) error {
    var getRoot []string
    var getNext []string

    for _, oid := range oids {
        getRoot = append(getRoot, oid.String())
        getNext = append(getNext, oid.String())
    }

    for row := 0; getNext != nil; row++ {
        if self.log != nil { self.log.Printf("snmp.GetNext %v...\n", getNext) }

        response, err := self.gosnmp.GetNext(getNext)
        if err != nil {
            return err
        }

        if len(response.Variables) != len(getNext) {
            return fmt.Errorf("snmp variable count mismatch: %v should be %v\n", row, len(response.Variables), len(getNext))
        }

        // load getNext
        for col, snmpVar := range response.Variables {
            if snmpVar.Type == gosnmp.EndOfMibView || snmpVar.Type == gosnmp.NoSuchObject || snmpVar.Type == gosnmp.NoSuchInstance {
                continue
            } else if !strings.HasPrefix(snmpVar.Name, getRoot[col]) || snmpVar.Name == getNext[col] {
                // stop
                getNext = nil
                break
            } else {
                getNext[col] = snmpVar.Name
            }
        }

        if getNext == nil {
            break
        } else {
            if err := handler(response.Variables); err != nil {
                return err
            }
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

    return self.getTable(tableMeta.fields, func(snmpRow []gosnmp.SnmpPDU) error {
        return loadTable(tableMeta, tableValue, snmpRow)
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
