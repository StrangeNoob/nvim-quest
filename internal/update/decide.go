package update

// ShouldCheck reports whether the launch-time update check should run. It is
// skipped when explicitly disabled (flag or env var) and for unversioned dev
// builds (which can't meaningfully compare against a release).
func ShouldCheck(noFlag bool, env, version string) bool {
	if noFlag || env != "" || version == "dev" {
		return false
	}
	return true
}
