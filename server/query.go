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
}

type objectQuery struct {
	ObjectQuery
	resultChan chan ObjectResult
	waitGroup  sync.WaitGroup
}

func (q *objectQuery) fail(host *Host, err error) {
	for _, object := range q.Objects {
		q.resultChan <- ObjectResult{Host: host, Object: object, Error: err}
	}
}

func (q *objectQuery) queryHost(host *Host) error {
	if err := host.client.WalkObjects(q.Objects.List(), func(object *mibs.Object, indexValues mibs.IndexValues, value mibs.Value, err error) error {
		q.resultChan <- ObjectResult{
			Host:        host,
			Object:      object,
			IndexValues: indexValues,
			Value:       value,
			Error:       err,
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (q *objectQuery) query() {
	defer close(q.resultChan)

	for _, host := range q.Hosts {
		q.waitGroup.Add(1)
		go func(host *Host) {
			defer q.waitGroup.Done()

			if err := q.queryHost(host); err != nil {
				q.fail(host, err)
			}
		}(host)
	}

	q.waitGroup.Wait()
}

type TableQuery struct {
	Hosts  Hosts
	Tables Tables
}

type tableQuery struct {
	TableQuery
	resultChan chan TableResult
	waitGroup  sync.WaitGroup
}

func (q *tableQuery) fail(host *Host, table *mibs.Table, err error) {
	q.resultChan <- TableResult{Host: host, Table: table, Error: err}
}

func (q *tableQuery) queryHostTable(host *Host, table *mibs.Table) error {
	if err := host.client.WalkTable(table, func(indexValues mibs.IndexValues, entryValues mibs.EntryValues, err error) error {
		q.resultChan <- TableResult{
			Host:        host,
			Table:       table,
			IndexValues: indexValues,
			EntryValues: entryValues,
			Error:       err,
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (q *tableQuery) query() {
	defer close(q.resultChan)

	for _, host := range q.Hosts {
		for _, table := range q.Tables {
			q.waitGroup.Add(1)
			go func(host *Host, table *mibs.Table) {
				defer q.waitGroup.Done()

				if err := q.queryHostTable(host, table); err != nil {
					q.fail(host, table, err)
				}
			}(host, table)
		}
	}

	q.waitGroup.Wait()
}
