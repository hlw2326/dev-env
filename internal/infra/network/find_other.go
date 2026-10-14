//go:build !windows

package network

func (c *Checker) FindProcess(_ int) (int, string, error) {
	return 0, "", nil
}
