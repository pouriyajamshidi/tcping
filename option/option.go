// Package options provides generic functional options pattern utilities.
package option

// Option represents a functional option that configures a value of type T.
type Option[T any] func(*T)
