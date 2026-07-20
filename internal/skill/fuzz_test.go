package skill

import "testing"

// INVARIANT: ValidateSemVer must accept exactly strings of the form
// MAJOR.MINOR.PATCH with no leading zeros, no prefixes, no pre-release.
// Fuzz property: for any string, ValidateSemVer must not panic, and
// the result must be stable (calling twice gives the same answer).

func FuzzSemVer(f *testing.F) {
	// Seed corpus: valid and invalid examples
	seeds := []string{
		"1.2.3", "0.0.1", "10.20.30", "255.255.255", // valid
		"v1.2.3", "1.2", "1", "1.2.3.4", "1.x.3",   // invalid
		"01.2.3", "1.2.3-beta", "1.2.3+meta", "",    // invalid
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, version string) {
		sk := &Skill{Version: version}
		// Must not panic.
		r1 := sk.ValidateSemVer()
		r2 := sk.ValidateSemVer()
		// Property: deterministic — same input, same output.
		if r1 != r2 {
			t.Fatalf("non-deterministic: %q gave %v then %v", version, r1, r2)
		}
	})
}
