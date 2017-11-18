package api

type Host struct {
	HostIndex

	Objects []Object
}

type HostObjects struct {
	HostID  string
	Objects []Object
}
