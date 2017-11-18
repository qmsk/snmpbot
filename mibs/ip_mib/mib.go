// IP-MIB
package ip_mib

import (
	"github.com/qmsk/snmpbot/mibs"
)

var IP = mibs.RegisterMIB("ip", 1, 3, 6, 1, 2, 1, 4)

var (
	ipAddrEntry = IP.MakeID("ipAddrEntry", 20, 1)
	ipAdEntAddr = IP.RegisterObject(ipAddrEntry.MakeID("ipAdEntAddr", 1), mibs.Object{
		Syntax: mibs.IPAddressSyntax{},
	})
	ipAddrIndexSyntax = mibs.IndexSyntax{
		ipAdEntAddr,
	}
	ipAddrTable = IP.RegisterTable(IP.MakeID("ipAddrTable", 20), mibs.Table{
		IndexSyntax: ipAddrIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			ipAdEntAddr,
			IP.RegisterObject(ipAddrEntry.MakeID("ipAdEntIfIndex", 2), mibs.Object{
				IndexSyntax: ipAddrIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			IP.RegisterObject(ipAddrEntry.MakeID("ipAdEntNetMask", 3), mibs.Object{
				IndexSyntax: ipAddrIndexSyntax,
				Syntax:      mibs.IPAddressSyntax{},
			}),
			IP.RegisterObject(ipAddrEntry.MakeID("ipAdEntBcastAddr", 4), mibs.Object{
				IndexSyntax: ipAddrIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
			IP.RegisterObject(ipAddrEntry.MakeID("ipAdEntReasmMaxSize", 5), mibs.Object{
				IndexSyntax: ipAddrIndexSyntax,
				Syntax:      mibs.IntegerSyntax{},
			}),
		},
	})
)

var (
	ipNetToMediaEntry   = IP.MakeID("ipNetToMediaEntry", 22, 1)
	ipNetToMediaIfIndex = IP.RegisterObject(ipNetToMediaEntry.MakeID("ipNetToMediaIfIndex", 1), mibs.Object{
		Syntax: mibs.IntegerSyntax{},
	})
	ipNetToMediaNetAddress = IP.RegisterObject(ipNetToMediaEntry.MakeID("ipNetToMediaNetAddress", 3), mibs.Object{
		Syntax: mibs.IPAddressSyntax{},
	})
	ipNetToMediaIndexSyntax = mibs.IndexSyntax{
		ipNetToMediaIfIndex,
		ipNetToMediaNetAddress,
	}

	ipNetToMediaTable = IP.RegisterTable(IP.MakeID("ipNetToMediaTable", 22), mibs.Table{
		IndexSyntax: ipNetToMediaIndexSyntax,
		EntrySyntax: mibs.EntrySyntax{
			ipNetToMediaIfIndex,
			IP.RegisterObject(ipNetToMediaEntry.MakeID("ipNetToMediaPhysAddress", 2), mibs.Object{
				IndexSyntax: ipNetToMediaIndexSyntax,
				Syntax:      mibs.PhysAddressSyntax{},
			}),
			ipNetToMediaNetAddress,
			IP.RegisterObject(ipNetToMediaEntry.MakeID("ipNetToMediaType", 4), mibs.Object{
				IndexSyntax: ipNetToMediaIndexSyntax,
				Syntax: mibs.EnumSyntax{
					{1, "other"},
					{2, "invalid"},
					{3, "dynamic"},
					{4, "static"},
				},
			}),
		},
	})
)

func init() {
	ipAdEntAddr.IndexSyntax = ipAddrIndexSyntax

	ipNetToMediaIfIndex.IndexSyntax = ipNetToMediaIndexSyntax
	ipNetToMediaNetAddress.IndexSyntax = ipNetToMediaIndexSyntax
}
