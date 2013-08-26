package graph

import (
	//"errors"
	"errors"
	"fmt"
	"github.com/cloudcube/database/graph/driver"
	"sync"
)

var drivers = make(map[string]driver.Driver)

type DB struct {
	driver driver.Driver
	dsn    string
	mu     sync.Mutex

	freeConn []*driverConn
	closed   bool
	dep      map[finalCloser]depSet //stacktrace of last conn's put;debug only
	lastPut  map[*driverConn]string //zero means defaultMaxIdleConns;negative means 0
	maxIdle  int
}

type driverConn struct {
	db *DB

	sync.Mutex  //guards following
	ci          driver.Conn
	finalClosed bool //ci.Close has been called
	openStmt    map[driver.Stmt]bool

	// guarded by db.mu
	inUse      bool
	onPut      []func() // code(with db.mu help) run when conn is next returned
	dbmuClosed bool     // same as closed,but guarded by db.mu,for connIfFree
}

// Register makes a database driver available by provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver driver.Driver) {
	if driver == nil {
		panic("graph:Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("graph:Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// Opens a database specified by its database driver name and a
// driver-specific data source name,usually consisting of at least a
// database name and connection information.
//
// Most users will open a database via a driver-specific connection
// helper function that return a *DB.No database drivers are included
// in the Go graph library.See http://github.com/graphdrivers for
// a list of third-party drivers.
//
// Open may just validata its arguments without creating a connection
// to the database.To verify that the data source name is valid,call
// Ping.

func Open(driverName, dataSourceName string) (*DB, error) {
	driveri, ok := driver[driverName]
	if !ok {
		return nil, fmt.Errorf("graph:unknown driver %q (forgotten import?", driverName)
	}
	db := &DB{
		driver:  driveri,
		dsn:     dataSourceName,
		lastPut: make(map[*driverConn]string),
	}
	return db, nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping() error {
	dc, err := db.conn()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) conn(*driverConn, error) {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil, errors.New("graph:database is closed")
	}
	if n := len(db.freeConn); n > 0 {
		conn := db.freeConn[n-1]
		db.freeConn = db.freeConn[:n-1]
		conn.inUse = true
		db.mu.Unlock()
		return conn, nil
	}
	db.mu.Unlock()

	ci, err := db.driver.Open(db.dsn)
	if err != nil {
		return nil, err
	}
	dc := &driverConn{
		db: db,
		ci: ci,
	}
	db.mu.Lock()
	db.addDepLocked(dc, dc)
	dc.inUse = true
	db.mu.Unlock()
	return dc, nil

}

type finalCloser interface {
	finalClose() error
}

func (dc *driverConn) finalClose() error {
	dc.Lock()

	for si := range dc.openStmt {
		si.Close()
	}
	dc.openStmt = nil

	err := dc.ci.Close()
	dc.ci = nil
	dc.finalClosed = true

	dc.Unlock()
	return err

}

// driverStmt associates a driver.Stmt with the
// *driverConn from which it came,so the driverConn's lock can be
// held during calls.
type driverStmt struct {
	sync.Locker //the *driverConn
	si          driver.Stmt
}

func (ds *driverStmt) Close() error {
	ds.Lock()
	defer ds.Unlock()
	return ds.si.Close()
}

// depSet is a finalCloser's outstanding dependencies
type depSet map[interface{}]bool //set of true bools

func (db *DB) addDepLocked(x finalCloser, dep interface{}) {
	if db.dep == nil {
		db.dep = make(map[finalCloser]depSet)
	}
	xdep := db.dep[x]
	if xdep == nil {
		xdep = make(depSet)
		db.dep[x] = xdep
	}
	xdep[dep] = true
}
