package snmp

// SNMP IF-MIB implementation

var (
	IfMIB      = registerMIB("IF-MIB", OID{1, 3, 6, 1, 2, 1, 31})
	Interfaces = registerMIB("interfaces", OID{1, 3, 6, 1, 2, 1, 2})

	Interfaces_linkDown = SNMPv2MIB.registerNotificationType("linkDown", Interfaces.define(1, 5, 3))
	Interfaces_linkUp   = SNMPv2MIB.registerNotificationType("linkUp", Interfaces.define(1, 5, 4))

	Interfaces_ifNumber = Interfaces.registerObject("ifNumber", IntegerSyntax, Interfaces.define(1))
	Interfaces_ifTable  = Interfaces.registerTable(&Table{Node: Node{OID: Interfaces.define(2), Name: "ifTable"},
		Index: []TableIndex{
			{"ifIndex", IntegerSyntax},
		},
		Entry: []*Object{
			Interfaces.registerObject("ifIndex", IntegerSyntax, Interfaces.define(2, 1, 1)),
			Interfaces.registerObject("ifDescr", StringSyntax, Interfaces.define(2, 1, 2)),
			Interfaces.registerObject("ifType", IntegerSyntax, Interfaces.define(2, 1, 3)),
			Interfaces.registerObject("ifMtu", IntegerSyntax, Interfaces.define(2, 1, 4)),
			Interfaces.registerObject("ifSpeed", GaugeSyntax, Interfaces.define(2, 1, 5)),
			Interfaces.registerObject("ifPhysAddress", PhysAddressSyntax, Interfaces.define(2, 1, 6)),
			Interfaces.registerObject("ifAdminStatus", EnumSyntax{
				{1, "up"},
				{2, "down"},
				{3, "testing"},
			}, Interfaces.define(2, 1, 7)),
			Interfaces.registerObject("ifOperStatus", EnumSyntax{
				{1, "up"},
				{2, "down"},
				{3, "testing"},
				{4, "unknown"},
				{5, "dormant"},
				{6, "notPresent"},
				{7, "lowerLayerDown"},
			}, Interfaces.define(2, 1, 8)),
			Interfaces.registerObject("ifLastChange", TimeTicksSyntax, Interfaces.define(2, 1, 9)),
		},
	})
)

type InterfaceIndex struct {
	Index Integer
}

func (self InterfaceIndex) parseIndex(oid OID) (interface{}, error) {
	if index, err := self.Index.parseIndex(oid); err != nil {
		return nil, err
	} else {
		return InterfaceIndex{Index: index.(Integer)}, nil
	}
}

func (self InterfaceIndex) String() string {
	return self.Index.String()
}

type InterfaceEntry struct {
	Index       Integer   `snmp:"1.3.6.1.2.1.2.2.1.1"`
	Descr       String    `snmp:"1.3.6.1.2.1.2.2.1.2"`
	Type        Integer   `snmp:"1.3.6.1.2.1.2.2.1.3"`
	Mtu         Integer   `snmp:"1.3.6.1.2.1.2.2.1.4"`
	Speed       Gauge     `snmp:"1.3.6.1.2.1.2.2.1.5"`
	PhysAddress Binary    `snmp:"1.3.6.1.2.1.2.2.1.6"`
	AdminStatus Integer   `snmp:"1.3.6.1.2.1.2.2.1.7"`
	OperStatus  Integer   `snmp:"1.3.6.1.2.1.2.2.1.8"`
	LastChange  TimeTicks `snmp:"1.3.6.1.2.1.2.2.1.9"`
}

type InterfaceTable map[InterfaceIndex]*InterfaceEntry
