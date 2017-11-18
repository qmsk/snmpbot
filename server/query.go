package server

import (
	"github.com/qmsk/snmpbot/mibs"
	"sync"
)

// Object may not be set if Error
type ObjectResult struct {
	Host        *Host
	Object      *mibs.Object
	IndexValues mibs.IndexValues
	Value       interface{}
	Error       error
}

type TableResult struct {
	Host        *Host
	Table       *mibs.Table
	IndexValues mibs.IndexValues
	EntryValues mibs.EntryValues
	Error       error
}

type ObjectQuery struct {
	Hosts   Hosts
	Objects Objects

	resultChan chan ObjectResult
	waitGroup  sync.WaitGroup
}

func (q *ObjectQuery) queryHost(host *Host) {
	var mibsClient = mibs.Client{host.snmpClient}
	defer q.waitGroup.Done()

	if err := mibsClient.WalkObjects(func(object *mibs.Object, indexValues mibs.IndexValues, value mibs.Value, err error) error {
		q.resultChan <- ObjectResult{
			Host:        host,
			Object:      object,
			IndexValues: indexValues,
			Value:       value,
			Error:       err,
		}
		return nil
	}, q.Objects.List()...); err != nil {
		q.resultChan <- ObjectResult{Host: host, Error: err}
	}
}

func (q *ObjectQuery) query() {
	defer close(q.resultChan)

	for _, host := range q.Hosts {
		q.waitGroup.Add(1)
		go q.queryHost(host)
	}

	q.waitGroup.Wait()
}

type TableQuery struct {
	Hosts  Hosts
	Tables Tables

	resultChan chan TableResult
	waitGroup  sync.WaitGroup
}

func (q *TableQuery) queryHostTable(host *Host, table *mibs.Table) {
	var mibsClient = mibs.Client{host.snmpClient}
	defer q.waitGroup.Done()

	if err := mibsClient.WalkTable(table, func(indexValues mibs.IndexValues, entryValues mibs.EntryValues) error {
		q.resultChan <- TableResult{
			Host:        host,
			Table:       table,
			IndexValues: indexValues,
			EntryValues: entryValues,
		}
		return nil
	}); err != nil {
		q.resultChan <- TableResult{Host: host, Table: table, Error: err}
	}
}

func (q *TableQuery) query() {
	defer close(q.resultChan)

	for _, host := range q.Hosts {
		for _, table := range q.Tables {
			q.waitGroup.Add(1)
			go q.queryHostTable(host, table)
		}
	}

	q.waitGroup.Wait()
}
