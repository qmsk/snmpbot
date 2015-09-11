package snmp

// IP-MIB
var (
    IPMIB           = registerMIB("IP-MIB", OID{1,3,6,1,2,1,4})

    IP_ipNetToMediaEntry    = IPMIB.define(22, 1)
    IP_ipNetToMediaTable    = IPMIB.registerTable(&Table{Node: Node{OID:IPMIB.define(22), Name: "ipNetToMediaTable"},
        Index: []TableIndex{
            {"ipNetToMediaIfIndex",     IntegerSyntax},
            {"ipNetToMediaNetAddress",  IpAddressSyntax},
        },
        Entry: []*Object{
            IPMIB.registerObject("ipNetToMediaIfIndex",     IntegerSyntax,      IP_ipNetToMediaEntry.define(1)),
            IPMIB.registerObject("ipNetToMediaPhysAddress", PhysAddressSyntax,  IP_ipNetToMediaEntry.define(2)),
            IPMIB.registerObject("ipNetToMediaNetAddress",  IpAddressSyntax,    IP_ipNetToMediaEntry.define(3)),
            IPMIB.registerObject("ipNetToMediaType",        EnumSyntax{
                {1, "other"},
                {2, "invalid"},
                {3, "dynamic"},
                {4, "static"},
            }, IP_ipNetToMediaEntry.define(4)),
        },
    })
)
