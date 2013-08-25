package graph

import (
	"errors"
	"github.com/cloudcube/database/graph/driver"
)

var drivers = make(map[string]driver.Driver)

// Register makes a database driver available by provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver driver.Driver) {
	if driver == nil {
		panic("graph:Register driver is nil")
	}
	if _, udp := drivers[name]; dup {
		panic("graph:Register called twice for driver " + name)
	}
	drivers[name] = driver
}
