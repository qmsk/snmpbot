package snmp

var (
    SNMPv2MIB               = registerMIB("SNMPv2-MIB", 1,3,6,1,6,3,1)
    SNMPv2_system           = registerMIB("system", 1,3,6,1,2,1,1)

    SNMPv2_sysDescr         = SNMPv2_system.define(1)
    SNMPv2_sysObjectID      = SNMPv2_system.define(2)
    SNMPv2_sysUpTime        = SNMPv2_system.define(3)

    SNMPv2_snmpTrapOID              = SNMPv2MIB.registerObject("trapOID", 1, 4, 1)                // 1.3.6.1.6.3.1.1.4.1
    SNMPv2_coldStart                = SNMPv2MIB.registerNotificationType("coldStart", 1, 5, 1)    // 1.3.6.1.6.3.1.1.5.1
    SNMPv2_warmStart                = SNMPv2MIB.registerNotificationType("warmStart", 1, 5, 2)
    SNMPv2_authenticationFailure    = SNMPv2MIB.registerNotificationType("authenticationFailure", 1, 5, 5)
)


