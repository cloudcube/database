// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package driver defines interfaces to be implemented by database
// drivers as used by package sql.
//
// Most code should use package sql.

package driver

import (
	"errors"
)

// Value is a value that drivers must be able to handle.
// It is either nil or an instance of one of these types:
//
//   int64
//   float64
//   bool
//   []byte
//   string   [*] everywhere except from Rows.Next.
//   time.Time
type Value interface{}

// Driver is the interface that must be implement by a database
// driver.
type Driver interface {
	Open(name string) (Conn, error)
}

// Conn is a connection to a database,It is not used concurrently
// by multiple goroutines.
//
//Conn is assumed to be stateful
type Conn interface {

	// Prepare returns a prepared statement,bound to this connection
	Prepare(query string) (Stmt, error)

	// Close invalidates and potentially stops any current
	// prepared statements and transactions,marking this
	// connection as no longer in use.
	//

	// Becuause the graph package maintains a free pool of
	// connections and only calls Close when there's a surplus of
	// idle connections,it shouldn't be necessary for drivers to
	// do their own connection caching.
	Close() error

	// Begin starts and returns a new transaction.
	Begin() (Tx, error)
}

type Stmt interface {
	Close() error

	NumInput() int

	Exec(args []Value) (Result, error)

	Query(args []Value) (Rows, error)
}

// Result is the result of a query execution.
type Result interface {

	// LastAddId returns the database's auto-generated ID
	// after,for example,an AddNode into a map with a
	// key.
	LastAddNodeId() (int64, error)

	// NodesAffected returns the number of nodes affected by the
	// query.
	NodesAffected() (int64, error)
}

// Tx is a transaction.
type Tx interface {
	Commit() error
	Rollback() error
}
