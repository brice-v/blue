//go:build wasm

package object

// psutilUnavailable is the error returned by every psutil builtin on wasm
// builds. The gopsutil dependency has no js/wasm support and host process
// info is not reachable from a sandboxed browser context anyway.
func psutilUnavailable(name string) Object {
	return newError("`%s` error: the psutil module is not available in the wasm build of blue", name)
}

// PsutilBuiltins for wasm builds.
var PsutilBuiltins = []*Builtin{
	{Name: "_cpu_usage_percent", Fun: func(args ...Object) Object { return psutilUnavailable("cpu_usage_percent") }},
	{Name: "_cpu_info", Fun: func(args ...Object) Object { return psutilUnavailable("cpu_info") }},
	{Name: "_cpu_time_info", Fun: func(args ...Object) Object { return psutilUnavailable("cpu_time_info") }},
	{Name: "_cpu_count", Fun: func(args ...Object) Object { return psutilUnavailable("cpu_count") }},
	{Name: "_mem_virt_info", Fun: func(args ...Object) Object { return psutilUnavailable("mem_virt_info") }},
	{Name: "_mem_swap_info", Fun: func(args ...Object) Object { return psutilUnavailable("mem_swap_info") }},
	{Name: "_host_info", Fun: func(args ...Object) Object { return psutilUnavailable("host_info") }},
	{Name: "_host_temps_info", Fun: func(args ...Object) Object { return psutilUnavailable("host_temps_info") }},
	{Name: "_net_connections", Fun: func(args ...Object) Object { return psutilUnavailable("net_connections") }},
	{Name: "_net_io_info", Fun: func(args ...Object) Object { return psutilUnavailable("net_io_info") }},
	{Name: "_disk_partitions", Fun: func(args ...Object) Object { return psutilUnavailable("disk_partitions") }},
	{Name: "_disk_io_info", Fun: func(args ...Object) Object { return psutilUnavailable("disk_io_info") }},
	{Name: "_disk_usage", Fun: func(args ...Object) Object { return psutilUnavailable("disk_usage") }},
}
