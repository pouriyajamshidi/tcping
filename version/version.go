// Package version reports which tcping build this is, both to the user and to
// the services tcping talks to.
package version

// Current is set at compile time via the Makefile
var Current = "beta"

// UserAgent is the User-Agent header tcping sends on every HTTP request it
// makes, so the receiving side can tell which tcping it is talking to.
var UserAgent = "pouriyajamshidi/tcping/" + Current
