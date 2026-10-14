package clientservice

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/core/security"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"strings"
	"time"
)

func (c *Controller) watchNetwork() {
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		ctx, cancel := context.WithTimeout(c.ctx, 8*time.Second)
		n := windowsplatform.InspectNetwork(ctx)
		cancel()
		c.network.Store(&n)
		c.Local.Suspend(len(n.Conflicts) > 0)
		c.Learned.Cleanup(time.Now())
		select {
		case <-c.ctx.Done():
			return
		case <-tick.C:
		}
	}
}
func (c *Controller) localKey(app, dest, protocol string) (localstrategy.TargetKey, error) {
	n := c.network.Load()
	if n == nil || n.Error != "" || n.Fingerprint == "" || time.Since(n.ObservedAt) > 45*time.Second {
		return localstrategy.TargetKey{}, errors.New("fresh network fingerprint unavailable")
	}
	return localstrategy.TargetKey{ConfigRevision: c.Local.Revision(), Network: n.Fingerprint, Application: strings.ToLower(strings.ReplaceAll(app, "\\", "/")), Destination: dest, Protocol: protocol, Version: localstrategy.Version}, nil
}
func (c *Controller) applyLocalAuto(id windowsplatform.ProcessIdentity, t windowsplatform.Tuple, d *windowsplatform.Decision) {
	c.mu.Lock()
	automatic := c.settings.Automatic
	c.mu.Unlock()
	if !automatic {
		return
	}
	k, err := c.localKey(id.Path, t.Destination.String(), t.Protocol)
	if err != nil {
		return
	}
	now := time.Now()
	all := c.Learned.Candidates(k, now)
	var direct, relay optimizer.RouteMeasurement
	for _, v := range all {
		if v.Strategy == "DIRECT" {
			direct = v.Measurement
		}
		if v.Strategy == "RELAY" {
			relay = v.Measurement
		}
	}
	// Failed DIRECT must also remain available to policy (Candidates excludes failures).
	if v, ok := c.Learned.Evidence(k, "DIRECT", now); ok {
		direct = v.Measurement
	}
	winner := localstrategy.Recommend(direct, all, relay, now)
	if winner == "DIRECT" {
		d.Route = routing.Route{}
		return
	}
	if winner == "RELAY" {
		if relay.RouteID != "" && relay.RouteID != optimizer.Direct && relay.RouteID != "controlled-target" {
			d.Route = routing.Route{RelayID: relay.RouteID}
		}
		return
	}
	if winner == "" {
		return
	}
	for _, p := range c.Local.Config().User {
		if p.ID == winner && p.Enabled {
			d.Route = routing.Route{Kind: routing.DirectPlusRoute}
			d.Local = c.Local
			d.LocalStrategyID = winner
			d.BlockOnFailure = false
			return
		}
	}
}

type localRecord struct {
	Application string                     `json:"application"`
	Destination string                     `json:"destination"`
	Protocol    string                     `json:"protocol"`
	Network     string                     `json:"network"`
	Strategy    string                     `json:"strategy"`
	Success     bool                       `json:"success"`
	Measurement optimizer.RouteMeasurement `json:"measurement"`
}

func (c *Controller) handleLocalEvidence(method string, raw json.RawMessage) (any, error) {
	var v localRecord
	if err := security.Parse(raw, &v); err != nil {
		return nil, err
	}
	k, err := c.localKey(v.Application, v.Destination, v.Protocol)
	if err != nil {
		return nil, err
	}
	if k.Network != v.Network {
		return nil, errors.New("network changed during test; discard evidence")
	}
	if method == "local.record" {
		if v.Strategy == "RELAY" {
			c.mu.Lock()
			_, registered := c.acceptanceRelays[v.Measurement.RouteID]
			c.mu.Unlock()
			if !registered {
				return nil, errors.New("relay evidence must name a configured authenticated relay")
			}
		}

		if v.Strategy != "DIRECT" && v.Strategy != "RELAY" {
			found := false
			for _, p := range c.Local.Config().User {
				if p.ID == v.Strategy && p.Protocol == v.Protocol {
					found = true
				}
			}
			if !found {
				return nil, errors.New("strategy version no longer configured")
			}
		}
		if err = c.Learned.Record(k, v.Strategy, v.Measurement, v.Success, time.Now()); err != nil {
			return nil, err
		}
	}
	all := c.Learned.Candidates(k, time.Now())
	var d, r optimizer.RouteMeasurement
	if x, ok := c.Learned.Evidence(k, "DIRECT", time.Now()); ok {
		d = x.Measurement
	}
	if x, ok := c.Learned.Evidence(k, "RELAY", time.Now()); ok {
		r = x.Measurement
	}
	return map[string]any{"recommended": localstrategy.Recommend(d, all, r, time.Now()), "evidence": all}, nil
}

// Compare is report-only; cross-process A/B/C tests cannot populate learned
// strategy evidence for an application that did not perform that route's trial.
func compareLocalReports(raw json.RawMessage) (any, error) {
	var q struct {
		Measurements []localRecord `json:"measurements"`
	}
	if err := security.Parse(raw, &q); err != nil {
		return nil, err
	}
	if len(q.Measurements) < 2 || len(q.Measurements) > 6 {
		return nil, errors.New("bounded comparison requires 2..6 measurements")
	}
	now := time.Now()
	var d, r optimizer.RouteMeasurement
	var local []localstrategy.Working
	for _, v := range q.Measurements {
		if now.Sub(v.Measurement.ObservedAt) > 90*time.Second || v.Measurement.ObservedAt.After(now) {
			return nil, errors.New("stale/future comparison")
		}
		switch v.Strategy {
		case "DIRECT":
			d = v.Measurement
		case "RELAY":
			r = v.Measurement
		default:
			if v.Success {
				local = append(local, localstrategy.Working{Strategy: v.Strategy, Measurement: v.Measurement, Expires: now.Add(time.Minute)})
			}
		}
	}
	return map[string]string{"recommended": localstrategy.Recommend(d, local, r, now), "scope": "report only; no cross-exe learning"}, nil
}
