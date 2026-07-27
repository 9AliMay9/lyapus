package health

import "testing"

func TestBranchProtectionProbe(t *testing.T) {
	t.Fatal("intentional failure: verify M0 branch protection")
}
