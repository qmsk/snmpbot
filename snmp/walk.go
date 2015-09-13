package snmp

import (
    wapsnmp "github.com/cdevr/WapSNMP"
)

// Walk through an unstructured tree, starting from the given root.
// Calls handler with each returned object, returning any error.
// Stops when walking outside of the given root, or the end of the MIB view.
func (self *Client) WalkTree(walkOID OID, handler func (oid OID, value interface{})) error {
    nextOID := walkOID.Copy()

    for {
        if varBinds, err := self.GetNext(nextOID); err != nil {
            return err
        } else {
            varBind := varBinds[0]
            varOID := OID(varBind.Name)

            self.log.Printf("WalkTree %v: %v\n", walkOID, varOID)

            if varBind.Value == wapsnmp.EndOfMibView {
                break
            } else if varOID.Equals(nextOID) {
                break
            } else if walkOID.Index(varOID) == nil {
                break
            } else {
                nextOID = varOID

                handler(varOID, varBind.Value)
            }
        }
    }

    return nil
}

// Walk through a table with the given list of entry field OIDs
// Call given handler with each returned row, returning any error.
// Any missing columns in the table will be given as empty VarBind's (Name == nil), as long as there is at least one column value.
func (self *Client) WalkTable(tableOID []OID, handler func ([]VarBind) error) error {
    var nextOID []OID

    for _, oid := range tableOID {
        nextOID = append(nextOID, oid)
    }

    // continue walking as long as we have something
    for {
        varBinds, err := self.GetNext(nextOID...)
        if err != nil {
            return err
        }

        // check result, update nextOIDs
        valid := false

        for i, varBind := range varBinds {
            oid := OID(varBind.Name)

            if varBind.Value == wapsnmp.EndOfMibView { // XXX: NoSuchObject, NoSuchInstance
                varBinds[i] = VarBind{}

                continue
            } else if oid.Equals(nextOID[i]) {
                varBinds[i] = VarBind{}

                // not making progress
                continue
            } else if tableOID[i].Index(oid) == nil {
                varBinds[i] = VarBind{}

                // walked out of table
                continue
            }

            nextOID[i] = oid
            valid = true
        }

        if !valid {
            break
        }

        // valid row
        if err := handler(varBinds); err != nil {
            return err
        }
    }

    return nil
}
