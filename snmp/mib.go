package snmp

import (
    "fmt"
    "log"
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

// Lookup a full OID against registered MIBs and their Objects.
func Lookup(oid OID) (*MIB, *Object, OID) {
    for _, mib := range mibs {
        if mibIndex := mib.Index(oid); mibIndex != nil {
            if object, objectIndex := mib.Lookup(oid); object != nil {
                return mib, object, objectIndex
            } else {
                return mib, nil, mibIndex
            }
        }
    }

    return nil, nil, nil
}

func LookupString(oid OID) string {
    mib, object, index := Lookup(oid)

    if object != nil {
        return fmt.Sprintf("%s::%s.%s", mib, object, index)
    } else if mib != nil {
        return fmt.Sprintf("%s::%s", mib, index)
    } else {
        return fmt.Sprintf("%s", oid)
    }
}

// Registry of OIDs within a MIB
type MIB struct {
    OID

    Name        string

    objects     []*Object
}

func (self MIB) String() string {
    return self.Name
}

// Lookup a full OID within this MIB
func (self *MIB) Lookup(oid OID) (*Object, OID) {
    for _, object := range self.objects {
        if objectIndex := object.Index(oid); objectIndex != nil {
            return object, objectIndex
        } else {
            log.Printf("MIB.Lookup %s: %s != %s\n", self, oid, object.OID)
        }
    }

    return nil, nil
}

func (self *MIB) registerObject(name string, ids... int) *Object {
    object := &Object{OID: self.define(ids...), Name: name}

    self.objects = append(self.objects, object)

    return object
}

func (self *MIB) registerNotificationType(name string, ids... int) *NotificationType {
    notificationType := &NotificationType{Object: Object{OID: self.define(ids...), Name: name}}

    self.objects = append(self.objects, &notificationType.Object) // XXX: srsly?

    return notificationType
}

// Object registered within a MIB
type Object struct {
    OID

    Name        string
}

func (self Object) String() string {
    return self.Name
}

// XXX: just use Object with additional fields instead?
type NotificationType struct {
    Object
}

