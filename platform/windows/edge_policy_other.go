//go:build !windows

package windowsplatform

func EdgeProxyPolicyWrite(string, []string, string) error { return ErrUnavailable }
func EdgeProxyPolicyActive(string, string) bool           { return false }
