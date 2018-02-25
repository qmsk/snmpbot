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

func (q *ObjectQuery) fail(host *Host, err error) {
	for _, object := range q.Objects {
		q.resultChan <- ObjectResult{Host: host, Object: object, Error: err}
	}
}

func (q *ObjectQuery) queryHost(host *Host) error {
	if client, err := host.getClient(); err != nil {
		return err
	} else if err := client.WalkObjects(func(object *mibs.Object, indexValues mibs.IndexValues, value mibs.Value, err error) error {
		q.resultChan <- ObjectResult{
			Host:        host,
			Object:      object,
			IndexValues: indexValues,
			Value:       value,
			Error:       err,
		}
		return nil
	}, q.Objects.List()...); err != nil {
		return err
	}

	return nil
}

func (q *ObjectQuery) query() {
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

	resultChan chan TableResult
	waitGroup  sync.WaitGroup
}

func (q *TableQuery) fail(host *Host, table *mibs.Table, err error) {
	q.resultChan <- TableResult{Host: host, Table: table, Error: err}
}

func (q *TableQuery) queryHostTable(host *Host, table *mibs.Table) error {
	if client, err := host.getClient(); err != nil {
		return err
	} else if err := client.WalkTable(table, func(indexValues mibs.IndexValues, entryValues mibs.EntryValues) error {
		q.resultChan <- TableResult{
			Host:        host,
			Table:       table,
			IndexValues: indexValues,
			EntryValues: entryValues,
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (q *TableQuery) query() {
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
