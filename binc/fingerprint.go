package binc

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"blue/code"
	"blue/consts"
	"blue/object"
)

// This file implements the build fingerprint: a compact description of the
// environment an image was compiled for. Loaders refuse mismatched images
// because bytecode and constants are only valid for the exact opcode set,
// reserved-constant layout, builtin surface and std sources they were
// compiled against (build tags like `static`/`rgfw` swap those out).

// Fingerprint returns the fingerprint of the RUNNING build. Images encode
// the fingerprint of their producing build; CheckEnvironment compares them.
//
// It covers:
//   - blue version (consts.VERSION, itself flavor-aware for `-static`)
//   - opcode set hash (names + operand widths of every opcode)
//   - reserved constant count
//   - go build tags (`-tags` from build info, empty when unset), with the
//     structural `minivm` tag filtered out: it selects which main package
//     is built, not what the runtime can do
//   - GOOS/GOARCH
func Fingerprint() string {
	tags := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "-tags" {
				tags = setting.Value
				break
			}
		}
	}
	tags = NormalizeTags(tags)
	return fmt.Sprintf("v%s|ops:%#016x|rc:%d|tags:%s|%s/%s",
		consts.VERSION,
		code.OpcodeSetFingerprint(),
		len(object.OBJECT_CONSTANTS),
		tags,
		runtime.GOOS, runtime.GOARCH,
	)
}

// minivmTag is the structural build tag of the minimal runner main package.
// It does not change the builtin/opcode surface so it is excluded from the
// fingerprint.
const minivmTag = "minivm"

// NormalizeTags sorts a comma-separated -tags value and drops minivm so
// that `blue` and `bluerun` built with the same flavor compare equal.
func NormalizeTags(tags string) string {
	parts := strings.Split(tags, ",")
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || p == minivmTag {
			continue
		}
		kept = append(kept, p)
	}
	sort.Strings(kept)
	return strings.Join(kept, ",")
}

// BlueVersion returns consts.VERSION of the running build.
func BlueVersion() string {
	return consts.VERSION
}

// CheckEnvironment verifies a container's recorded fingerprint/blueVersion
// against the running build, returning actionable errors on mismatch.
func CheckEnvironment(fingerprint, blueVersion string) error {
	if blueVersion != consts.VERSION {
		return fmt.Errorf("%w: image was compiled with blue v%s, running v%s", ErrFingerprintMismatch, blueVersion, consts.VERSION)
	}
	if fingerprint != Fingerprint() {
		return fmt.Errorf(`%w: image fingerprint
  image: %s
  build: %s
The image was compiled for a different build flavor (build tags like static/rgfw change the builtin and opcode surface). Recompile the program with this same binary/flavor.`, ErrFingerprintMismatch, fingerprint, Fingerprint())
	}
	return nil
}

// DescribeFingerprintMismatch is a helper for tests/tools that want to know
// whether two fingerprints differ in a specific component.
func DescribeFingerprintMismatch(a, b string) string {
	if a == b {
		return ""
	}
	for _, part := range []string{"ops:", "rc:", "tags:"} {
		pa := componentOf(a, part)
		pb := componentOf(b, part)
		if pa != pb {
			return part + strconv.Quote(pa) + " vs " + strconv.Quote(pb)
		}
	}
	return "version or platform"
}

func componentOf(fp, prefix string) string {
	i := indexOf(fp, prefix)
	if i < 0 {
		return ""
	}
	start := i + len(prefix)
	end := start
	for end < len(fp) && fp[end] != '|' {
		end++
	}
	return fp[start:end]
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
