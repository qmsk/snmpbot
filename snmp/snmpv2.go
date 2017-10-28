package snmp

var (
	SNMPv2MIB     = registerMIB("SNMPv2-MIB", OID{1, 3, 6, 1, 6, 3, 1})
	SNMPv2_system = registerMIB("system", OID{1, 3, 6, 1, 2, 1, 1})

	SNMPv2_snmpTrapOID = SNMPv2MIB.define(1, 4, 1)

	SNMPv2_sysDescr        = SNMPv2_system.registerObject("sysDescr", StringSyntax, SNMPv2_system.define(1))
	SNMPv2_sysObjectID     = SNMPv2_system.registerObject("sysObjectID", OIDSyntax, SNMPv2_system.define(2))
	SNMPv2_sysUpTime       = SNMPv2_system.registerObject("sysUpTime", TimeTicksSyntax, SNMPv2_system.define(3))
	SNMPv2_sysContact      = SNMPv2_system.registerObject("sysContact", StringSyntax, SNMPv2_system.define(4))
	SNMPv2_sysName         = SNMPv2_system.registerObject("sysName", StringSyntax, SNMPv2_system.define(5))
	SNMPv2_sysLocation     = SNMPv2_system.registerObject("sysLocation", StringSyntax, SNMPv2_system.define(6))
	SNMPv2_sysServices     = SNMPv2_system.registerObject("sysServices", IntegerSyntax, SNMPv2_system.define(7))
	SNMPv2_sysORLastChange = SNMPv2_system.registerObject("sysORLastChange", TimeTicksSyntax, SNMPv2_system.define(8))

	SNMPv2_coldStart             = SNMPv2MIB.registerNotificationType("coldStart", SNMPv2MIB.define(1, 5, 1)) // 1.3.6.1.6.3.1.1.5.1
	SNMPv2_warmStart             = SNMPv2MIB.registerNotificationType("warmStart", SNMPv2MIB.define(1, 5, 2))
	SNMPv2_authenticationFailure = SNMPv2MIB.registerNotificationType("authenticationFailure", SNMPv2MIB.define(1, 5, 5))
)
