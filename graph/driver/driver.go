// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package driver defines interfaces to be implemented by database
// drivers as used by package sql.
//
// Most code should use package sql.

package driver

import (
// "errors"
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

	//CreateNode and return a node.
	CreateNode() (Node, error)

	// Node returns a node.
	Node() (Node, error)

	//CreateRelationship returns a relationship.
	CreateRelationship() (Relationship, error)

	//Get relationship types.
	RelationshipTypes() ([]string, error)

	// Close invalidates and potentially stops any current
	// prepared statements and transactions,marking this
	// connection as no longer in use.
	//
	// Because the graph package maintains a free pool of
	// connections and only calls Close when there's a surplus of
	// idle connections,it shouldn't be necessary for drivers to
	// do their own connection caching.
	Close() error
}

type Node interface {

	//SetProperty on node.
	SetProperty(propertyKey string, propertyVal interface{}) error

	//UpdateProperties.
	UpdateProperties(properties map[string]interface{}) error

	//Get properties for node.
	Properties() (map[string]interface{}, error)

	// Get all relationships.
	Relationships() ([]*Relationship, error)

	//Get Incoming relationships.
	IncomingRelationships() ([]*Relationship, error)

	//Get outgoing relationships.
	OutgoingRelationships() ([]*Relationship, error)

	//Get typed relationships.
	TypedRelationships(...string) ([]*Relationship, error)

	//Delete an node.
	Delete() error
}

type Relationship interface {
	//Update relationship properties.
	UpdataProperties(properties map[string]interface{}) error

	//RemoveProperties from a relationship.
	RemoveProperties() error

	// Remove property from a relationship.
	RemoveProperty(propertyKey string) error

	//Properties get all relationship's properties.
	Properties() (map[string]interface{}, error)

	//Property get single property on a relationship.
	Property(propertyKey string) (string, error)

	//SetProperties on a relationship.
	SetProperties(map[string]interface{}) error

	//SetProperty on a relationship.
	SetProperty(propertyKey string, propertyVal interface{}) error

	//Delete a relationship.
	Delete() error
}

// Stmt is a prepared statement.It is bound to a Conn and not
// used by multiple goroutines concurrently.
type Stmt interface {
	// Close closes the statement.
	//
	// As of Go 1.1,a Stmt will not be closed if it's in use
	// by any queries.
	Close() error
}

// Execer is an optional interface that may be implemented by a Conn.
//
// If a Conn does not implement Execer,the graph package's DB.Exec will
// first prepare a query,execute the statement,and then close the statement.
//
//Exec may return ErrSkip.
type Execer interface {
}

// Queryer is an Optional interface that may be implemented by a Conn.
//
// If a Conn does not implement Queryer,the graph package's DB.Query will
// first prepare a query,execute the statement,and then Close the statement.
//
// Query may return ErrSkip.
type Queryer interface {
}
