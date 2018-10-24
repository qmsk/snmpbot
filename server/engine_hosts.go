package server

import (
	"sync"
)

type engineHosts struct {
	mutex sync.Mutex

	Hosts
}

// Returns false if host already exists
func (hosts *engineHosts) Add(host *Host) bool {
	hosts.mutex.Lock()
	defer hosts.mutex.Unlock()

	if _, exists := hosts.Hosts[host.id]; !exists {
		hosts.Hosts[host.id] = host
		return true
	} else {
		return false
	}
}

func (hosts *engineHosts) Set(host *Host) {
	hosts.mutex.Lock()
	defer hosts.mutex.Unlock()

	hosts.Hosts[host.id] = host
}

// Returns false if host does not exist
func (hosts *engineHosts) Del(host *Host) bool {
	hosts.mutex.Lock()
	defer hosts.mutex.Unlock()

	if _, exists := hosts.Hosts[host.id]; exists {
		delete(hosts.Hosts, host.id)
		return true
	} else {
		return false
	}
}

func (hosts *engineHosts) Copy() Hosts {
	var copy = make(Hosts)

	hosts.mutex.Lock()
	defer hosts.mutex.Unlock()

	for hostID, host := range hosts.Hosts {
		copy[hostID] = host
	}

	return copy
}
