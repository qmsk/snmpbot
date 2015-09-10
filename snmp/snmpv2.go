package snmp

var (
    SNMPv2MIB               = registerMIB("SNMPv2-MIB", 1,3,6,1,6,3,1)
    SNMPv2_system           = registerMIB("system", 1,3,6,1,2,1,1)

    SNMPv2_sysDescr         = SNMPv2_system.registerObject("sysDescr",          StringSyntax,       1)
    SNMPv2_sysObjectID      = SNMPv2_system.registerObject("sysObjectID",       OIDSyntax,          2)
    SNMPv2_sysUpTime        = SNMPv2_system.registerObject("sysUpTime",         TimeTicksSyntax,    3)
    SNMPv2_sysContact       = SNMPv2_system.registerObject("sysContact",        StringSyntax,       4)
    SNMPv2_sysName          = SNMPv2_system.registerObject("sysName",           StringSyntax,       5)
    SNMPv2_sysLocation      = SNMPv2_system.registerObject("sysLocation",       StringSyntax,       6)
    SNMPv2_sysServices      = SNMPv2_system.registerObject("sysServices",       IntegerSyntax,      7)
    SNMPv2_sysORLastChange  = SNMPv2_system.registerObject("sysORLastChange",   TimeTicksSyntax,    8)

    SNMPv2_snmpTrapOID              = SNMPv2MIB.registerObject("trapOID",       OIDSyntax,          1, 4, 1)        // 1.3.6.1.6.3.1.1.4.1
    SNMPv2_coldStart                = SNMPv2MIB.registerNotificationType("coldStart", 1, 5, 1)    // 1.3.6.1.6.3.1.1.5.1
    SNMPv2_warmStart                = SNMPv2MIB.registerNotificationType("warmStart", 1, 5, 2)
    SNMPv2_authenticationFailure    = SNMPv2MIB.registerNotificationType("authenticationFailure", 1, 5, 5)
)
