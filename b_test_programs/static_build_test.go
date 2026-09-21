//go:build static

package b_program_test

// staticBuild reports whether the test binary is the static flavor. In static
// builds ui, gg and plot compile to error stubs (see lib/std/*-static.b), so
// their integration programs cannot run and carry a `# STATIC IGNORE` header.
const staticBuild = true
