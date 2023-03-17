package orm

import "database/sql"

type MsDb struct {
	db *sql.DB
}

func Open(driverName string, source string) *MsDb {
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	msDb := &MsDb{
		db: db,
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return msDb
}
