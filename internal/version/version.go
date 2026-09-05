// Package version reports which tcping build this is, both to the user and to
// the services tcping talks to.
package version

// Current is the version of this build. It is the single place the version
// number is written down: the Makefile reads it from here for the .deb
// package, so bumping a release means editing this line only.
var Current = "3.0.0-rc1"

// UserAgent is the User-Agent header tcping sends on every HTTP request it
// makes, so the receiving side can tell which tcping it is talking to.
var UserAgent = "pouriyajamshidi/tcping/" + Current
