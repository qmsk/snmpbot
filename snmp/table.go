package snmp

import (
    "fmt"
    "github.com/soniah/gosnmp"
    "log"
    "reflect"
    "strings"
)

type tableField struct {
    name    string
    oid     OID
}

// Populate a map[IndexType]*struct{... Type `snmp:"oid"`} from SNMP
func (self *Client) GetTable(table interface{}) error {
    var fields []tableField

    tableType := reflect.TypeOf(table).Elem()
    indexType := tableType.Key()
    entryType := tableType.Elem().Elem()

    tableValue := reflect.ValueOf(table).Elem()

    for i := 0; i < entryType.NumField(); i++ {
        field := entryType.Field(i)
        snmpTag := field.Tag.Get("snmp")

        if snmpTag == "" {
            panic(fmt.Errorf("Missing snmp tag for field: %s.%s", entryType.Name(), field.Name))
        }

        oid := parseOID(snmpTag)

        log.Printf("snmp.Client.GetTable: field %v:%v = %s %s\n", i, field.Name, oid, field.Type.Name())

        fields = append(fields, tableField{
            name:   field.Name,
            oid:    oid,
        })
    }

    // snmp get
    var getRoot []string
    var getNext []string

    for _, fieldTab := range fields {
        getRoot = append(getRoot, fieldTab.oid.String())
        getNext = append(getNext, fieldTab.oid.String())
    }

    for row := 0; getNext != nil; row++ {
        response, err := self.snmp.GetNext(getNext)
        if err != nil {
            return err
        } else {
            log.Printf("snmp.Client.GetTable: snmp.GetNext %v: variables=%v\n", getNext, len(response.Variables))
        }

        if len(response.Variables) != len(getNext) {
            log.Printf("snmp.Client.GetTable: response row %v variable count mismatch: %v should be %v\n", row, len(response.Variables), len(getNext))
            break
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
        }

        // load row
        for i, snmpVar := range response.Variables {
            fieldTab := fields[i]

            oid := parseOID(snmpVar.Name)
            oidIndex := fieldTab.oid.Index(oid)

            // index
            indexValue := reflect.New(indexType).Elem()
            index := indexValue.Addr().Interface().(IndexType)

            if err := index.setIndex(oidIndex); err != nil {
                return err
            }

            // entry
            entryValue := tableValue.MapIndex(indexValue)

            if !entryValue.IsValid() {
                entryValue = reflect.New(entryType)
                tableValue.SetMapIndex(indexValue, entryValue)
            }

            // field
            fieldValue := entryValue.Elem().Field(i)
            field := fieldValue.Addr().Interface().(Type)

            if err := field.set(snmpVar.Type, snmpVar.Value); err != nil {
                return err
            }

            log.Printf("snmp.Client.GetTable: get %v.%v[%v] = %v\n", entryType.Name(), fieldTab.name, index, field)
        }
    }

    return nil
}
