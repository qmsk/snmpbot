package server

import (
	"github.com/qmsk/snmpbot/mibs"
	"sync"
)

type Query struct {
	Hosts   Hosts
	Objects Objects
}

type query struct {
	resultChan chan Result
	waitGroup  sync.WaitGroup
}

func (q *query) executeHost(host *Host, objects Objects) {
	var mibsClient = mibs.Client{host.snmpClient}
	defer q.waitGroup.Done()

	if err := mibsClient.WalkObjects(func(object *mibs.Object, indexValues mibs.IndexValues, value mibs.Value, err error) error {
		q.resultChan <- Result{
			Host:        host,
			Object:      object,
			IndexValues: indexValues,
			Value:       value,
			Error:       err,
		}
		return nil
	}, objects.List()...); err != nil {
		q.resultChan <- Result{Host: host, Error: err}
	}
}

func (Q Query) execute(resultChan chan Result) {
	var q = query{
		resultChan: resultChan,
		waitGroup:  sync.WaitGroup{},
	}

	defer close(resultChan)

	for _, host := range Q.Hosts {
		q.waitGroup.Add(1)
		go q.executeHost(host, Q.Objects)
	}

	q.waitGroup.Wait()
}

// Object may not be set if Error
type Result struct {
	Host        *Host
	Object      *mibs.Object
	IndexValues mibs.IndexValues
	Value       interface{}
	Error       error
}
