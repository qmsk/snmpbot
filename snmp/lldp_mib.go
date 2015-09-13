package snmp

var (
    LLDP_MIB                    = registerMIB("LLDP-MIB", OID{1,0,8802,1,1,2})

    LLDP_lldpObjects            = LLDP_MIB.define(1)
    LLDP_lldpLocalSystemData   = LLDP_lldpObjects.define(3)
    LLDP_lldpRemoteSystemsData  = LLDP_lldpObjects.define(4)

    LLDP_LldpChassisIdSubtypeSyntax = EnumSyntax{
        {1, "chassisComponent"},
        {2, "interfaceAlias"},
        {3, "portComponent"},
        {4, "macAddress"},
        {5, "networkAddress"},
        {6, "interfaceName"},
        {7, "local"},
    }
    LLDP_LldpPortIdSubtypeSyntax = EnumSyntax{
        {1, "interfaceAlias"},
        {2, "portComponent"},
        {3, "macAddress"},
        {4, "networkAddress"},
        {5, "interfaceName"},
        {6, "agentCircuitId"},
        {7, "local"},
    }

    LLDP_lldpLocChassisIdSubtype    = LLDP_MIB.registerObject("lldpLocChassisIdSubtype", LLDP_LldpChassisIdSubtypeSyntax, LLDP_lldpLocalSystemData.define(1))
    LLDP_lldpLocChassisId           = LLDP_MIB.registerObject("lldpLocChassisId", BinarySyntax, LLDP_lldpLocalSystemData.define(2))
    LLDP_lldpLocSysName             = LLDP_MIB.registerObject("lldpLocSysName", StringSyntax, LLDP_lldpLocalSystemData.define(3))
    LLDP_lldpLocSysDesc             = LLDP_MIB.registerObject("lldpLocSysDesc", StringSyntax, LLDP_lldpLocalSystemData.define(4))
    // LLDP_lldpLocSysCapSupported  // TODO: .5 BitsSyntax
    // LLDP_lldpLocSysCapEnabled    // TODO: .6 BitsSyntax

    LLDP_lldpLocPortEntry       = LLDP_lldpLocalSystemData.define(7, 1)
    LLDP_lldpLocalPortTable     = LLDP_MIB.registerTable(&Table{Node:Node{OID: LLDP_lldpLocalSystemData.define(7), Name: "lldpLocPortTable"},
        Index: []TableIndex{
            {"lldpLocPortNum",          IntegerSyntax},
        },
        Entry: []*Object{
            // lldpLocPortNum       .1  not-accessible
            LLDP_MIB.registerObject("lldpLocPortIdSubtype",     LLDP_LldpPortIdSubtypeSyntax,   LLDP_lldpLocPortEntry.define(2)),
            LLDP_MIB.registerObject("lldpLocPortId",            BinarySyntax,                   LLDP_lldpLocPortEntry.define(3)),
            LLDP_MIB.registerObject("lldpLocPortDesc",          StringSyntax,                   LLDP_lldpLocPortEntry.define(4)),
        },
    })

    LLDP_lldpRemEntry           = LLDP_lldpRemoteSystemsData.define(1, 1)
    LLDP_lldpRemTable           = LLDP_MIB.registerTable(&Table{Node:Node{OID: LLDP_lldpRemoteSystemsData.define(1), Name: "lldpRemTable"},
        Index: []TableIndex{
            {"lldpRemTimeMark",         TimeTicksSyntax},
            {"lldpRemLocalPortNum",     IntegerSyntax},
            {"lldpRemIndex",            IntegerSyntax},
        },
        Entry: []*Object{
            // lldpRemTimeMark      .1  not-accessible
            // lldpRemLocalPortNum  .2  not-accessible
            // lldpRemIndex         .3  not-accessible
            LLDP_MIB.registerObject("lldpRemChassisIdSubtype",  LLDP_LldpChassisIdSubtypeSyntax,    LLDP_lldpRemEntry.define(4)),
            LLDP_MIB.registerObject("lldpRemChassisId",         BinarySyntax,                       LLDP_lldpRemEntry.define(5)),
            LLDP_MIB.registerObject("lldpRemPortIdSubtype",     LLDP_LldpPortIdSubtypeSyntax,       LLDP_lldpRemEntry.define(6)),
            LLDP_MIB.registerObject("lldpRemPortId",            BinarySyntax,                       LLDP_lldpRemEntry.define(7)),
            LLDP_MIB.registerObject("lldpRemPortDesc",          StringSyntax,                       LLDP_lldpRemEntry.define(8)),
            LLDP_MIB.registerObject("lldpRemSysName",           StringSyntax,                       LLDP_lldpRemEntry.define(9)),
            LLDP_MIB.registerObject("lldpRemSysDesc",           StringSyntax,                       LLDP_lldpRemEntry.define(10)),
            //LLDP_MIB.registerObject("lldpRemSysCapSupported",   BitsSyntax,                         LLDP_lldpRemEntry.define(11)),  // TODO: BER + Syntax
            //LLDP_MIB.registerObject("lldpRemSysCapEnabled",     BitsSyntax,                         LLDP_lldpRemEntry.define(12)),  // TODO: BER + Syntax
        },
    })
)
