package api

type MIBIndex struct {
	ID string
}

type MIB struct {
	MIBIndex

	Objects []ObjectIndex
	Tables  []TableIndex
}
