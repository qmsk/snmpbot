package snmp

import (
    wapsnmp "github.com/cdevr/WapSNMP"
)

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
