package snmp

import (
    "fmt"
    "strings"
)

var (
    // global registry of MIBs
    mibs []*MIB
)

func registerMIB(name string, ids... int) *MIB {
    mib := &MIB{OID: OID(ids), Name: name}

    mibs = append(mibs, mib)

    return mib
}

// Return MIB by OID
func LookupMIB(oid OID) *MIB {
    for _, mib := range mibs {
        if mibIndex := mib.Index(oid); mibIndex != nil {
            return mib
        }
    }
    return nil
}

// Return Object by OID
func LookupObject(oid OID) *Object {
    if mib := LookupMIB(oid); mib == nil {
        return nil
    } else if object := mib.LookupObject(oid); object == nil {
        return nil
    } else {
        return object
    }
}

// Return MIB by Name
func ResolveMIB(name string) *MIB {
    for _, mib := range mibs {
        if mib.String() == name {
            return mib
        }
    }
    return nil
}

// Lookup a full OID against registered MIBs and their Objects.
// Returns MIB, Object, OID{...} if the given oid indexes an object
// Returns MIB, Object, nil if the given oid matches an object exactly
// Returns MIB, nil, OID{...} if the given oid matches to a MIB, but is not a registered object
// Returns nil, nil, nil if the given OID is unknown
func Lookup(oid OID) (*MIB, *Object, OID) {
    if mib := LookupMIB(oid); mib != nil {
        if object, objectIndex := mib.Lookup(oid); object != nil {
            return mib, object, objectIndex
        } else {
            return mib, nil, mib.Index(oid)
        }
    }

    return nil, nil, nil
}

// Lookup OID to a MIB-Object, and return a human-readable string
func LookupString(oid OID) string {
    mib, object, index := Lookup(oid)

    if object != nil && index != nil {
        return fmt.Sprintf("%s::%s%s", mib, object, index)
    } else if object != nil {
        return fmt.Sprintf("%s::%s", mib, object)
    } else if mib != nil {
        return fmt.Sprintf("%s%s", mib, index)
    } else {
        return fmt.Sprintf("%s", oid)
    }
}

func LookupTable(oid OID) *Table {
    if mib := LookupMIB(oid); mib == nil {
        return nil
    } else if table := mib.LookupTable(oid); table == nil {
        return nil
    } else {
        return table
    }
}

// Return Object by OID, MIB::Object Name
// Return nil if not found
func ResolveObject(name string) *Object {
    if name[0] == '.' {
        return LookupObject(ParseOID(name))

    } else if sepIndex := strings.Index(name, "::"); sepIndex != -1 {
        mibName := name[:sepIndex]
        objectName := name[sepIndex+2:]

        if mib := ResolveMIB(mibName); mib == nil {
            return nil
        } else if object := mib.ResolveObject(objectName); object == nil {
            return nil
        } else {
            return object
        }

    } else {
        panic(fmt.Errorf("Invalid name: %v", name))
    }
}

// Return OID from OID, MIB.Name::Object.Name
// Return nil if not found
func Resolve(name string) OID {
    if name[0] == '.' {
        return ParseOID(name)

    } else if sepIndex := strings.Index(name, "::"); sepIndex != -1 {
        mibName := name[:sepIndex]
        objectName := name[sepIndex+2:]

        if mib := ResolveMIB(mibName); mib == nil {
            return nil
        } else if object := mib.ResolveObject(objectName); object == nil {
            return nil
        } else {
            return object.OID
        }

    } else if sepIndex := strings.Index(name, "."); sepIndex != -1 {
        mibName := name[:sepIndex]
        mibIndex := name[sepIndex+1:]

        if mib := ResolveMIB(mibName); mib == nil {
            return nil
        } else {
            return mib.define(ParseOID(mibIndex)...)
        }

    } else if mib := ResolveMIB(name); mib != nil {
        return mib.OID

    } else {
        panic(fmt.Errorf("Invalid name: %v", name))
    }
}

func ResolveTable(name string) *Table {
    if sepIndex := strings.Index(name, "::"); sepIndex == -1 {
        return LookupTable(ParseOID(name))
    } else {
        mibName := name[:sepIndex]
        tableName := name[sepIndex+2:]

        if mib := ResolveMIB(mibName); mib == nil {
            return nil
        } else if table := mib.ResolveTable(tableName); table == nil {
            return nil
        } else {
            return table
        }
    }
}

// Registry of OIDs within a MIB
type MIB struct {
    OID

    Name        string

    objects     []*Object
    tables      []*Table
}

func (self MIB) String() string {
    return self.Name
}

// Lookup a full OID within this MIB
func (self *MIB) Lookup(oid OID) (*Object, OID) {
    for _, object := range self.objects {
        if object.Equals(oid) {
            return object, nil
        } else if objectIndex := object.Index(oid); objectIndex != nil {
            return object, objectIndex
        }
    }

    return nil, nil
}

// Return Object by OID, or nil
func (self *MIB) LookupObject(oid OID) *Object {
    for _, object := range self.objects {
        if objectIndex := object.Index(oid); objectIndex != nil {
            return object
        }
    }

    return nil
}

// Return Table by OID, or nil
func (self *MIB) LookupTable(oid OID) *Table {
    for _, table := range self.tables {
        if table.Equals(oid) {
            return table
        }
    }
    return nil
}

// Return Object by Name, or nil
func (self *MIB) ResolveObject(name string) *Object {
    for _, object := range self.objects {
        if object.String() == name {
            return object
        }
    }
    return nil
}

// Return Table by Name, or nil
func (self *MIB) ResolveTable(name string) *Table {
    for _, table := range self.tables {
        if table.String() == name {
            return table
        }
    }
    return nil
}

func (self *MIB) register(object *Object) *Object {
    self.objects = append(self.objects, object)

    return object
}

// Build and register Object
func (self *MIB) registerObject(name string, syntax Syntax, ids... int) *Object {
    object := &Object{
        OID:    self.define(ids...),
        Name:   name,
        Syntax: syntax,
    }

    self.objects = append(self.objects, object)

    return object
}

func (self *MIB) registerNotificationType(name string, ids... int) *NotificationType {
    notificationType := &NotificationType{Object: Object{OID: self.define(ids...), Name: name}}

    self.objects = append(self.objects, &notificationType.Object) // XXX: srsly?

    return notificationType
}

func (self *MIB) registerTable(table *Table) *Table {
    self.tables = append(self.tables, table)

    return table
}

// Object registered within a MIB
type Object struct {
    OID

    Name        string
    Syntax      Syntax
}

func (self Object) String() string {
    return self.Name
}

// Parse a raw SNMP value per its Syntax
func (self Object) ParseValue(snmpValue interface{}) (interface{}, error) {
    if self.Syntax == nil {
        return snmpValue, nil
    } else {
        return self.Syntax.parseValue(snmpValue)
    }
}

// XXX: just use Object with additional fields instead?
type NotificationType struct {
    Object
}
