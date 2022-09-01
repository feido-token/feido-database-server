package main

import (
	"database/sql"
	feido_proto "github.com/feido-token/feido-database-server/feido-proto"
	_ "github.com/mattn/go-sqlite3"
)

type eidDatabase interface {
	Open(path string) error
	Close() error

	isRevoked(quote *feido_proto.ICheckitQuery) (bool, error)
}

type sqltDB struct {
	db *sql.DB
}

func (sdb *sqltDB) Open(path string) (err error) {
	sdb.db, err = sql.Open("sqlite3", path)
	return
}

func (sdb *sqltDB) Close() error {
	return sdb.db.Close()
}

func (sdb *sqltDB) isRevoked(query *feido_proto.ICheckitQuery) (bool, error) {
	const checkIfRevoked = "SELECT id FROM revoked_eids WHERE doc_num == ? AND doc_country == ? AND doc_type == ?;"
	stmt, err := sdb.db.Prepare(checkIfRevoked)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(query.TravelDocumentNumber[:], query.CountryOfIssuance[:], query.DocumentType[:])
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if !rows.Next() {
		return false, nil
	}

	// eID occurs in database
	return true, nil
}
