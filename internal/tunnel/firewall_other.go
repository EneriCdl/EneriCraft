//go:build !windows

package tunnel

func AddFirewallRule(port int) error { return nil }
func RemoveFirewallRule(port int)  {}
