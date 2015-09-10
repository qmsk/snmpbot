package snmp

var (
    SNMPv2MIB               = MIB{OID{1,3,6,1,6,3,1}}
    SNMPv2_system           = MIB{OID{1,3,6,1,2,1,1}}

    SNMPv2_sysDescr         = SNMPv2_system.define(1)
    SNMPv2_sysObjectID      = SNMPv2_system.define(2)
    SNMPv2_sysUpTime        = SNMPv2_system.define(3)

    SNMPv2_snmpTrapOID      = SNMPv2MIB.define(1, 4, 1)                // 1.3.6.1.6.3.1.1.4.1
)


