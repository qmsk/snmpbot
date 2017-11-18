package server

import (
	"github.com/qmsk/snmpbot/api"
	"github.com/qmsk/snmpbot/client"
	"github.com/qmsk/snmpbot/mibs"
	"log"
)

type mibsWrapper struct{}

type mibWrapper struct {
	*mibs.MIB
}

func (_ mibsWrapper) probeHost(client *client.Client, f func(mib mibWrapper)) error {
	var mibsClient = mibs.Client{client}
	var ids []mibs.ID

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		ids = append(ids, mib.ID)
	})

	log.Printf("Probing MIBs: %v", ids)

	if probed, err := mibsClient.ProbeMany(ids); err != nil {
		return err
	} else {
		for _, id := range ids {
			log.Printf("Probed %v = %v", id, probed[id.Key()])

			if probed[id.Key()] {
				f(mibWrapper{id.MIB})
			}
		}

		return nil
	}
}

func (_ mibsWrapper) makeAPIIndex() []api.MIBIndex {
	var index []api.MIBIndex

	mibs.WalkMIBs(func(mib *mibs.MIB) {
		var mibIndex = api.MIBIndex{
			ID:      mib.String(),
			Objects: []api.ObjectIndex{},
			Tables:  []api.TableIndex{},
		}

		mib.Walk(func(id mibs.ID) {
			if object := mib.Object(id); object != nil {
				mibIndex.Objects = append(mibIndex.Objects, objectView{object}.makeAPIIndex())
			}

			if table := mib.Table(id); table != nil {
				mibIndex.Tables = append(mibIndex.Tables, tableView{table}.makeAPIIndex())
			}
		})

		index = append(index, mibIndex)
	})

	return index
}
