//go:build !windows

package skip

func shExecutable() (string, error) {
	return "sh", nil
}
