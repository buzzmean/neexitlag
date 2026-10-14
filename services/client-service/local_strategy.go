package clientservice

import (
	"encoding/json"
	"errors"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/security"
	"os"
	"path/filepath"
)

func (c *Controller) loadLocal() error {
	b, err := os.ReadFile(filepath.Join(filepath.Dir(c.customPath), "local-strategies.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) > 24000 {
		return errors.New("local config too large")
	}
	var v localstrategy.Config
	if err = security.Parse(b, &v); err != nil {
		return err
	}
	return c.Local.Reload(v)
}
func (c *Controller) handleLocal(method string, raw json.RawMessage) (any, error) {
	if method == "local.compare" {
		return compareLocalReports(raw)
	}
	if method == "local.network" {
		if n := c.network.Load(); n != nil {
			return n, nil
		}
		return nil, errors.New("network inspection pending")
	}
	if method == "local.record" || method == "local.recommend" {
		return c.handleLocalEvidence(method, raw)
	}
	if method == "local.get" {
		return c.Local.Config(), nil
	}
	if method == "local.metrics" {
		return c.Local.Metrics(), nil
	}
	c.localMu.Lock()
	defer c.localMu.Unlock()
	var v localstrategy.Config
	if err := security.Parse(raw, &v); err != nil {
		return nil, err
	}
	if len(v.Builtin) > 0 {
		return nil, errors.New("built-in strategies are reserved; use user profiles")
	}
	if err := c.Local.Validate(v); err != nil {
		return nil, err
	}
	if c.customPath != "" {
		dir := filepath.Dir(c.customPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
		f, err := os.CreateTemp(dir, ".local-strategies-*")
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		f.Chmod(0600)
		b, _ := json.Marshal(v)
		_, err = f.Write(b)
		if err == nil {
			err = f.Sync()
		}
		ce := f.Close()
		if err != nil {
			return nil, err
		}
		if ce != nil {
			return nil, ce
		}
		if err = os.Rename(f.Name(), filepath.Join(dir, "local-strategies.json")); err != nil {
			return nil, err
		}
	}
	if err := c.Local.Reload(v); err != nil {
		return nil, err
	}
	return c.Local.Config(), nil
}
