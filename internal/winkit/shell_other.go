//go:build !windows

package winkit

// Shell returns no shell: the runtime is WinKit's, for Windows.
func Shell(string) (string, bool) { return "", false }
