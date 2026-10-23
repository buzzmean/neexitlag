package localcatalog

import (
	"crypto/ed25519"
	"crypto/rand"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/security"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSignedCatalogExpiryDomainTamperRollbackAndPersistentCache(t *testing.T) {
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	keys := map[string]ed25519.PublicKey{"fixture": pub}
	now := time.Now().UTC()
	doc := Document{Version: 2, GeneratedAt: now, ExpiresAt: now.Add(time.Hour), Catalog: localstrategy.BundledCatalog()}
	sign := func(d Document) []byte {
		raw, err := security.Sign(Domain, "fixture", key, d)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	raw := sign(doc)
	path := filepath.Join(t.TempDir(), "catalog.json")
	cache := New(keys, path)
	if err := cache.Apply(raw, now); err != nil {
		t.Fatal(err)
	}
	again := New(keys, path)
	if err := again.Load(now); err != nil {
		t.Fatal(err)
	}
	old := doc
	old.Version = 1
	if err := again.Apply(sign(old), now); err == nil {
		t.Fatal("rollback accepted")
	}
	other := doc
	other.Catalog.Reviewed = "other"
	if err := again.Apply(sign(other), now); err == nil {
		t.Fatal("same-version equivocation accepted")
	}
	renewed := doc
	renewed.GeneratedAt = now.Add(time.Minute)
	renewed.ExpiresAt = now.Add(2 * time.Hour)
	if err := again.Apply(sign(renewed), now.Add(time.Minute)); err != nil {
		t.Fatal("same data refresh failed", err)
	}
	wrong, _ := security.Sign("config", "fixture", key, doc)
	if err := again.Apply(wrong, now); err == nil {
		t.Fatal("wrong signature domain accepted")
	}
	bad := append([]byte(nil), raw...)
	bad[len(bad)-15] ^= 1
	if _, err := Validate(keys, bad, now); err == nil {
		t.Fatal("tamper accepted")
	}
	if _, err := again.Current(now.Add(3 * time.Hour)); err == nil {
		t.Fatal("expired catalog returned")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(sign(renewed)) {
		t.Fatal("rejected update changed disk cache")
	}
}
