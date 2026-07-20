package folder

import "testing"

// INVARIANT: CanonicalPath must never contain ".." or a trailing slash
// when it reports clean. The fuzzer generates random path strings.
// Property: if ValidateCanonicalPathIsClean returns nil error, the path
// contains no ".." segment and no trailing slash (except root "/").

func FuzzCanonicalPath(f *testing.F) {
	seeds := []string{
		"/clean/path", "/home/user", "/a/b/c",
		"/bad/../escape", "/trailing/", "relative", "/..",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, path string) {
		f := &Folder{CanonicalPath: path}
		err := f.ValidateCanonicalPathIsClean()
		// Must not panic.
		// Property: if clean, then no SEGMENT equals ".." and no trailing slash.
		if err == nil {
			if hasTraversalSegment(path) {
				t.Fatalf("reported clean but has '..' segment: %q", path)
			}
			if len(path) > 1 && path[len(path)-1] == '/' {
				t.Fatalf("reported clean but has trailing slash: %q", path)
			}
		}
	})
}

// hasTraversalSegment returns true if any path segment is exactly "..".
// This is the real invariant — a substring like "0.." is a valid filename.
func hasTraversalSegment(path string) bool {
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if path[start:i] == ".." {
				return true
			}
			start = i + 1
		}
	}
	return false
}
