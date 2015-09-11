package snmp

import (
    "fmt"
    "strings"
)

var (
    // global registry of MIBs
    mibs []*MIB
)

func registerMIB(name string, oid OID) *MIB {
    mib := &MIB{Node:Node{OID: oid, Name: name}}

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

// Return MIB by Name
func ResolveMIB(name string) *MIB {
    for _, mib := range mibs {
        if mib.String() == name {
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

func LookupTable(oid OID) *Table {
    if mib := LookupMIB(oid); mib== nil {
        return nil
    } else if table := mib.LookupTable(oid); table == nil {
        return nil
    } else {
        return table
    }
}

func LookupNotificationType(oid OID) *NotificationType {
    if mib := LookupMIB(oid); mib== nil {
        return nil
    } else if notificationType := mib.LookupNotificationType(oid); notificationType == nil {
        return nil
    } else {
        return notificationType
    }
}

// Return a human-readble string representation of the OID, including an MIB, Object and Index
func FormatObject(oid OID) string {
    if mib := LookupMIB(oid); mib == nil {
        return oid.String()
    } else if object := mib.LookupObject(oid); object == nil {
        return mib.Format(oid)
    } else {
        return object.Format(oid)
    }
}

// Return a human-readble string representation of the OID, including an MIB, Object and Index
func FormatNotificationType(oid OID) string {
    if mib := LookupMIB(oid); mib == nil {
        return oid.String()
    } else if notificationType := mib.LookupNotificationType(oid); notificationType == nil {
        return mib.Format(oid)
    } else {
        return notificationType.Format(oid)
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

type Node struct {
    OID

    Name        string
}

func (self Node) String() string {
    return self.Name
}

// Registry of OIDs within a MIB
type MIB struct {
    Node

    objects             []*Object
    tables              []*Table
    notificationTypes   []*NotificationType
}

func (self MIB) String() string {
    return self.Name
}

// Format a full OID per this Object
func (self MIB) Format(oid OID) string {
    if index := self.Index(oid); index == nil {
        return fmt.Sprintf("%s", self)
    } else {
        return fmt.Sprintf("%s%s", self, index)
    }
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

// Return NotificationType by OID, or nil
func (self *MIB) LookupNotificationType(oid OID) *NotificationType {
    for _, notificationType := range self.notificationTypes {
        if notificationType.Equals(oid) {
            return notificationType
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

// Build and register Object
func (self *MIB) registerObject(name string, syntax Syntax, oid OID) *Object {
    object := &Object{
        Node:   Node{OID: oid, Name:name},
        MIB:    self,
        Syntax: syntax,
    }

    self.objects = append(self.objects, object)


    return object
}

func (self *MIB) registerNotificationType(name string, oid OID) *NotificationType {
    notificationType := &NotificationType{
        Node:   Node{OID: oid, Name: name},
        MIB:    self,
    }

    self.notificationTypes = append(self.notificationTypes, notificationType)

    return notificationType
}

func (self *MIB) registerTable(table *Table) *Table {
    table.MIB = self

    self.tables = append(self.tables, table)

    // register objects as belonging to a Table
    for _, tableEntry := range table.Entry {
        tableEntry.Table = table
    }

    return table
}

// SNMPv2-SMI OBJECT-TYPE
// Object contained within a MIB.
// Can be scalar, or with a Table-Entry.
type Object struct {
    Node
    MIB         *MIB        // part of what MIB

    Syntax      Syntax

    Table       *Table      // optional, if part of a table
}

func (self Object) String() string {
    return fmt.Sprintf("%s::%s", self.MIB, self.Name)
}

// Format a full OID per this Object
func (self Object) Format(oid OID) string {
    if index := self.Index(oid); index == nil {
        return fmt.Sprintf("%s", self)
    } else {
        return fmt.Sprintf("%s%s", self, index)
    }
}

// Parse a raw SNMP value per its Syntax
func (self Object) ParseValue(snmpValue interface{}) (interface{}, error) {
    if self.Syntax == nil {
        return snmpValue, nil
    } else {
        return self.Syntax.parseValue(snmpValue)
    }
}

// SNMPv2-SMI NOTIFICATION-TYPE
// Used in SNMPv2 Trap-PDU SNMPv2::snmpTrapOID as an OID value to determine the trapped event. Does not carry any value.
type NotificationType struct {
    Node
    MIB         *MIB
}

func (self NotificationType) String() string {
    return fmt.Sprintf("%s::%s", self.MIB, self.Name)
}

func (self NotificationType) Format(oid OID) string {
    if index := self.Index(oid); index == nil {
        return fmt.Sprintf("%s", self)
    } else {
        return fmt.Sprintf("%s%s", self, index)
    }
}
