// Package options provides generic functional options pattern utilities.
package options

// Option represents a functional option that configures a value of type T.
type Option[T any] func(*T)
