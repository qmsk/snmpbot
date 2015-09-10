package snmp

import (
    "fmt"
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
// Returns MIB, Object, OID{...} if the given oid indexes an object
// Returns MIB, Object, nil if the given oid matches an object exactly
// Returns MIB, nil, OID{...} if the given oid matches to a MIB, but is not a registered object
// Returns nil, nil, nil if the given OID is unknown
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

// Resolve an OID to a human-readable string
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
        if object.Equals(oid) {
            return object, nil
        } else if objectIndex := object.Index(oid); objectIndex != nil {
            return object, objectIndex
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

