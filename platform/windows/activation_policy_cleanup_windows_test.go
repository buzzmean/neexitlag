//go:build windows

package windowsplatform

import (
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os"
	"testing"
	"time"
)

func TestActivationOwnedPolicyCleanupPreservesForeignSettings(t *testing.T) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf(`Software\VELIQTest\ActivationCleanup-%d-%d`, os.Getpid(), time.Now().UnixNano())
	k, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.ALL_ACCESS)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.DeleteKey(registry.CURRENT_USER, path)
	defer k.Close()
	owned := `{"ProxyMode":"pac_script","ProxyPacUrl":"data:application/x-ns-proxy-autoconfig;base64,ZnVuY3Rpb24gRmluZFByb3h5Rm9yVVJMKCl7cmV0dXJuICdESVJFQ1QnO30=","ProxyPacMandatory":true}`
	if err = k.SetStringValue("ProxySettings", owned); err != nil {
		t.Fatal(err)
	}
	k.SetStringValue("ProxyMode", "foreign")
	if err = browserProxyPolicyRemove(user.User.Sid.String(), []string{owned}, path); err != nil {
		t.Fatal(err)
	}
	if _, _, err = k.GetStringValue("ProxySettings"); err != registry.ErrNotExist {
		t.Fatal("owned value survived")
	}
	if v, _, _ := k.GetStringValue("ProxyMode"); v != "foreign" {
		t.Fatal("split foreign value changed")
	}
	if err = browserProxyPolicyRemove(user.User.Sid.String(), []string{owned}, path); err != nil {
		t.Fatal("repeat cleanup", err)
	}
	k.SetStringValue("ProxySettings", `{"ProxyMode":"direct"}`)
	if err = browserProxyPolicyRemove(user.User.Sid.String(), []string{owned}, path); err == nil {
		t.Fatal("foreign policy accepted")
	}
	if v, _, _ := k.GetStringValue("ProxySettings"); v != `{"ProxyMode":"direct"}` {
		t.Fatal("foreign policy changed")
	}
}
