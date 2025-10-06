package evaluator

import (
	"blue/object"
	"fmt"
)

type helpStrArgs struct {
	explanation string
	signature   string
	errors      string
	example     string
}

func (hsa helpStrArgs) String() string {
	return fmt.Sprintf("%s\n    Signature:  %s\n    Error(s):   %s\n    Example(s): %s\n", hsa.explanation, hsa.signature, hsa.errors, hsa.example)
}

type BuiltinMapType struct {
	*object.ConcurrentMap[string, *object.Builtin]
}

func NewBuiltinObjMap(input map[string]*object.Builtin) BuiltinMapType {
	return BuiltinMapType{&object.ConcurrentMap[string, *object.Builtin]{
		Kv: input,
	}}
}

type BuiltinMapTypeInternal map[string]*object.Builtin

var builtins = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"help":    object.GetBuiltinByName("help"),
	"new":     object.GetBuiltinByName("new"),
	"keys":    object.GetBuiltinByName("keys"),
	"values":  object.GetBuiltinByName("values"),
	"del":     object.GetBuiltinByName("del"),
	"len":     object.GetBuiltinByName("len"),
	"append":  object.GetBuiltinByName("append"),
	"prepend": object.GetBuiltinByName("prepend"),
	"push":    object.GetBuiltinByName("push"),
	"pop":     object.GetBuiltinByName("pop"),
	"unshift": object.GetBuiltinByName("unshift"),
	"shift":   object.GetBuiltinByName("shift"),
	"concat":  object.GetBuiltinByName("concat"),
	"reverse": object.GetBuiltinByName("reverse"),
	"println": object.GetBuiltinByName("println"),
	"print":   object.GetBuiltinByName("print"),
	"input":   object.GetBuiltinByName("input"),
	// Default headers seem to be host, user-agent, accept-encoding (not case sensitive for these check pictures)
	// deno also used accept: */* (not sure what that is)
	"_fetch": object.GetBuiltinByName("_fetch"),
	"_read":  object.GetBuiltinByName("_read"),
	"_write": object.GetBuiltinByName("_write"),
	"set":    object.GetBuiltinByName("set"),
	// This function is lossy
	"int": object.GetBuiltinByName("int"),
	// This function is lossy
	"float": object.GetBuiltinByName("float"),
	// This function is lossy
	"bigint": object.GetBuiltinByName("bigint"),
	// This function is lossy
	"bigfloat": object.GetBuiltinByName("bigfloat"),
	// This function is lossy
	"uint":          object.GetBuiltinByName("uint"),
	"eval_template": object.GetBuiltinByName("eval_template"),
	"error":         object.GetBuiltinByName("error"),
	"assert":        object.GetBuiltinByName("assert"),
	"type":          object.GetBuiltinByName("type"),
	"exec":          object.GetBuiltinByName("exec"),
	"is_alive":      object.GetBuiltinByName("is_alive"),
	"exit":          object.GetBuiltinByName("exit"),
	"cwd":           object.GetBuiltinByName("cwd"),
	"cd":            object.GetBuiltinByName("cd"),
	"_to_bytes":     object.GetBuiltinByName("_to_bytes"),
	"str":           object.GetBuiltinByName("str"),
	"is_file":       object.GetBuiltinByName("is_file"),
	"is_dir":        object.GetBuiltinByName("is_dir"),
	"find_exe":      object.GetBuiltinByName("find_exe"),
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	"rm": object.GetBuiltinByName("rm"),
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	"ls": object.GetBuiltinByName("ls"),
	// TODO: Eventually we need to support files better (and possibly, stdin, stderr, stdout) and then http stuff
	"is_valid_json": object.GetBuiltinByName("is_valid_json"),
	"wait":          object.GetBuiltinByName("wait"),
	"_publish":      object.GetBuiltinByName("_publish"),
	"_broadcast":    object.GetBuiltinByName("_broadcast"),
	// Functions for subscribers in pubsub
	"add_topic": object.GetBuiltinByName("add_topic"),
	// TODO: add_topics, remove_topics?
	"remove_topic":          object.GetBuiltinByName("remove_topic"),
	"_subscribe":            object.GetBuiltinByName("_subscribe"),
	"unsubscribe":           object.GetBuiltinByName("unsubscribe"),
	"_pubsub_sub_listen":    object.GetBuiltinByName("_pubsub_sub_listen"),
	"_get_subscriber_count": object.GetBuiltinByName("_get_subscriber_count"),
	"_kv_put":               object.GetBuiltinByName("_kv_put"),
	"_kv_get":               object.GetBuiltinByName("_kv_get"),
	"_kv_delete":            object.GetBuiltinByName("_kv_delete"),
	"_new_uuid":             object.GetBuiltinByName("_new_uuid"),
	// This is straight out of golang's example for runtime/metrics https://pkg.go.dev/runtime/metrics
	"_go_metrics":        object.GetBuiltinByName("_go_metrics"),
	"get_os":             object.GetBuiltinByName("get_os"),
	"get_arch":           object.GetBuiltinByName("get_arch"),
	"_gc":                object.GetBuiltinByName("_gc"),
	"_version":           object.GetBuiltinByName("_version"),
	"_num_cpu":           object.GetBuiltinByName("_num_cpu"),
	"_num_process":       object.GetBuiltinByName("_num_process"),
	"_num_max_cpu":       object.GetBuiltinByName("_num_max_cpu"),
	"_num_os_thread":     object.GetBuiltinByName("_num_os_thread"),
	"_set_max_cpu":       object.GetBuiltinByName("_set_max_cpu"),
	"_set_gc_percent":    object.GetBuiltinByName("_set_gc_percent"),
	"_get_mem_stats":     object.GetBuiltinByName("_get_mem_stats"),
	"_get_stack_trace":   object.GetBuiltinByName("_get_stack_trace"),
	"_free_os_mem":       object.GetBuiltinByName("_free_os_mem"),
	"_print_stack_trace": object.GetBuiltinByName("_print_stack_trace"),
	"_set_max_stack":     object.GetBuiltinByName("_set_max_stack"),
	"_set_max_threads":   object.GetBuiltinByName("_set_max_threads"),
	"_set_mem_limit":     object.GetBuiltinByName("_set_mem_limit"),
	"re":                 object.GetBuiltinByName("re"),
	"to_list":            object.GetBuiltinByName("to_list"),
	"abs_path":           object.GetBuiltinByName("abs_path"),
	"fmt":                object.GetBuiltinByName("fmt"),
	"save":               object.GetBuiltinByName("save"),
	"__hash":             object.GetBuiltinByName("__hash"),
})

func GetBuiltins(e *Evaluator) BuiltinMapType {
	b := builtins
	b.Put("to_num", createToNumBuiltin(e))
	b.Put("_sort", createSortBuiltin(e))
	b.Put("_sorted", createSortedBuiltin(e))
	b.Put("all", createAllBuiltin(e))
	b.Put("any", createAnyBuiltin(e))
	b.Put("map", createMapBuiltin(e))
	b.Put("filter", createFilterBuiltin(e))
	b.Put("load", createLoadBuiltin(e))
	return b
}
