package main

import (
	"errors"
	feidoProto "github.com/feido-token/feido-database-server/feido-proto"
)

type databaseManager struct {
	eidDB *sqltDB
}

func (dbm *databaseManager) Init(dbPath string) (err error) {
	if dbPath == "" {
		err = errors.New("no database path given")
		return
	}

	dbm.eidDB = &sqltDB{}
	if err = dbm.eidDB.Open(dbPath); err != nil {
		return
	}
	return
}

func (dbm *databaseManager) Fini() {
	if dbm.eidDB != nil {
		dbm.eidDB.Close()
	}
}

func (dbm *databaseManager) LookupEID(query *feidoProto.ICheckitQuery) (isRevoked bool, err error) {
	if dbm.eidDB == nil {
		err = errors.New("no database")
		return
	}

	isRevoked, err = dbm.eidDB.isRevoked(query)
	return
}
