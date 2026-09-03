// tcping.go probes a target using TCP
package main

import "github.com/pouriyajamshidi/tcping/v3/internal/app"

/* TODO:
Library support, so other codebases can import tcping (in priority order):
1. Move internal/config, internal/dns, internal/nic, internal/printers,
   internal/probers and internal/stats out of internal/ so they're
   importable by other modules - blocks everything below
2. Make config.ProcessUserInput's CLI-only bits (os.Args, the global
   flag.CommandLine, os.Exit on invalid input) usable programmatically
3. Do not let checkForUpdates make a network call or exit when reachable
   from library code - keep it strictly opt-in CLI behavior
4. Give Statistics a Snapshot() so a caller can read it while a Prober is
   running. Nothing reads it unsynchronized today, but a library caller
   has no safe way to do it either.
5. Design a curated top-level public API/entrypoint instead of requiring
   several packages to be wired together by hand
6. Treat exported types/fields/methods as a public API contract once this
   is importable (semver discipline)
7. Sending over proxy connections
*/

func main() {
	app.Run()
}
