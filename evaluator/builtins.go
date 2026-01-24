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
	"help":    object.GetBuiltinByName(object.BuiltinBaseType, "help"),
	"new":     object.GetBuiltinByName(object.BuiltinBaseType, "new"),
	"keys":    object.GetBuiltinByName(object.BuiltinBaseType, "keys"),
	"values":  object.GetBuiltinByName(object.BuiltinBaseType, "values"),
	"del":     object.GetBuiltinByName(object.BuiltinBaseType, "del"),
	"len":     object.GetBuiltinByName(object.BuiltinBaseType, "len"),
	"append":  object.GetBuiltinByName(object.BuiltinBaseType, "append"),
	"prepend": object.GetBuiltinByName(object.BuiltinBaseType, "prepend"),
	"push":    object.GetBuiltinByName(object.BuiltinBaseType, "push"),
	"pop":     object.GetBuiltinByName(object.BuiltinBaseType, "pop"),
	"unshift": object.GetBuiltinByName(object.BuiltinBaseType, "unshift"),
	"shift":   object.GetBuiltinByName(object.BuiltinBaseType, "shift"),
	"concat":  object.GetBuiltinByName(object.BuiltinBaseType, "concat"),
	"reverse": object.GetBuiltinByName(object.BuiltinBaseType, "reverse"),
	"println": object.GetBuiltinByName(object.BuiltinBaseType, "println"),
	"print":   object.GetBuiltinByName(object.BuiltinBaseType, "print"),
	"input":   object.GetBuiltinByName(object.BuiltinBaseType, "input"),
	// Default headers seem to be host, user-agent, accept-encoding (not case sensitive for these check pictures)
	// deno also used accept: */* (not sure what that is)
	"_fetch": object.GetBuiltinByName(object.BuiltinBaseType, "_fetch"),
	"_read":  object.GetBuiltinByName(object.BuiltinBaseType, "_read"),
	"_write": object.GetBuiltinByName(object.BuiltinBaseType, "_write"),
	"set":    object.GetBuiltinByName(object.BuiltinBaseType, "set"),
	// This function is lossy
	"int": object.GetBuiltinByName(object.BuiltinBaseType, "int"),
	// This function is lossy
	"float": object.GetBuiltinByName(object.BuiltinBaseType, "float"),
	// This function is lossy
	"bigint": object.GetBuiltinByName(object.BuiltinBaseType, "bigint"),
	// This function is lossy
	"bigfloat": object.GetBuiltinByName(object.BuiltinBaseType, "bigfloat"),
	// This function is lossy
	"uint":          object.GetBuiltinByName(object.BuiltinBaseType, "uint"),
	"eval_template": object.GetBuiltinByName(object.BuiltinBaseType, "eval_template"),
	"error":         object.GetBuiltinByName(object.BuiltinBaseType, "error"),
	"assert":        object.GetBuiltinByName(object.BuiltinBaseType, "assert"),
	"type":          object.GetBuiltinByName(object.BuiltinBaseType, "type"),
	"exec":          object.GetBuiltinByName(object.BuiltinBaseType, "exec"),
	"is_alive":      object.GetBuiltinByName(object.BuiltinBaseType, "is_alive"),
	"exit":          object.GetBuiltinByName(object.BuiltinBaseType, "exit"),
	"cwd":           object.GetBuiltinByName(object.BuiltinBaseType, "cwd"),
	"cd":            object.GetBuiltinByName(object.BuiltinBaseType, "cd"),
	"_to_bytes":     object.GetBuiltinByName(object.BuiltinBaseType, "_to_bytes"),
	"str":           object.GetBuiltinByName(object.BuiltinBaseType, "str"),
	"is_file":       object.GetBuiltinByName(object.BuiltinBaseType, "is_file"),
	"is_dir":        object.GetBuiltinByName(object.BuiltinBaseType, "is_dir"),
	"find_exe":      object.GetBuiltinByName(object.BuiltinBaseType, "find_exe"),
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	"rm": object.GetBuiltinByName(object.BuiltinBaseType, "rm"),
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	"ls": object.GetBuiltinByName(object.BuiltinBaseType, "ls"),
	// TODO: Eventually we need to support files better (and possibly, stdin, stderr, stdout) and then http stuff
	"is_valid_json": object.GetBuiltinByName(object.BuiltinBaseType, "is_valid_json"),
	"from_json":     object.GetBuiltinByName(object.BuiltinBaseType, "from_json"),
	"wait":          object.GetBuiltinByName(object.BuiltinBaseType, "wait"),
	"_publish":      object.GetBuiltinByName(object.BuiltinBaseType, "_publish"),
	"_broadcast":    object.GetBuiltinByName(object.BuiltinBaseType, "_broadcast"),
	// Functions for subscribers in pubsub
	"add_topic": object.GetBuiltinByName(object.BuiltinBaseType, "add_topic"),
	// TODO: add_topics, remove_topics?
	"remove_topic":          object.GetBuiltinByName(object.BuiltinBaseType, "remove_topic"),
	"_subscribe":            object.GetBuiltinByName(object.BuiltinBaseType, "_subscribe"),
	"unsubscribe":           object.GetBuiltinByName(object.BuiltinBaseType, "unsubscribe"),
	"_pubsub_sub_listen":    object.GetBuiltinByName(object.BuiltinBaseType, "_pubsub_sub_listen"),
	"_get_subscriber_count": object.GetBuiltinByName(object.BuiltinBaseType, "_get_subscriber_count"),
	"_kv_put":               object.GetBuiltinByName(object.BuiltinBaseType, "_kv_put"),
	"_kv_get":               object.GetBuiltinByName(object.BuiltinBaseType, "_kv_get"),
	"_kv_delete":            object.GetBuiltinByName(object.BuiltinBaseType, "_kv_delete"),
	"_new_uuid":             object.GetBuiltinByName(object.BuiltinBaseType, "_new_uuid"),
	// This is straight out of golang's example for runtime/metrics https://pkg.go.dev/runtime/metrics
	"_go_metrics":        object.GetBuiltinByName(object.BuiltinBaseType, "_go_metrics"),
	"get_os":             object.GetBuiltinByName(object.BuiltinBaseType, "get_os"),
	"get_arch":           object.GetBuiltinByName(object.BuiltinBaseType, "get_arch"),
	"_gc":                object.GetBuiltinByName(object.BuiltinBaseType, "_gc"),
	"_version":           object.GetBuiltinByName(object.BuiltinBaseType, "_version"),
	"_num_cpu":           object.GetBuiltinByName(object.BuiltinBaseType, "_num_cpu"),
	"_num_process":       object.GetBuiltinByName(object.BuiltinBaseType, "_num_process"),
	"_num_max_cpu":       object.GetBuiltinByName(object.BuiltinBaseType, "_num_max_cpu"),
	"_num_os_thread":     object.GetBuiltinByName(object.BuiltinBaseType, "_num_os_thread"),
	"_set_max_cpu":       object.GetBuiltinByName(object.BuiltinBaseType, "_set_max_cpu"),
	"_set_gc_percent":    object.GetBuiltinByName(object.BuiltinBaseType, "_set_gc_percent"),
	"_get_mem_stats":     object.GetBuiltinByName(object.BuiltinBaseType, "_get_mem_stats"),
	"_get_stack_trace":   object.GetBuiltinByName(object.BuiltinBaseType, "_get_stack_trace"),
	"_free_os_mem":       object.GetBuiltinByName(object.BuiltinBaseType, "_free_os_mem"),
	"_print_stack_trace": object.GetBuiltinByName(object.BuiltinBaseType, "_print_stack_trace"),
	"_set_max_stack":     object.GetBuiltinByName(object.BuiltinBaseType, "_set_max_stack"),
	"_set_max_threads":   object.GetBuiltinByName(object.BuiltinBaseType, "_set_max_threads"),
	"_set_mem_limit":     object.GetBuiltinByName(object.BuiltinBaseType, "_set_mem_limit"),
	"re":                 object.GetBuiltinByName(object.BuiltinBaseType, "re"),
	"to_list":            object.GetBuiltinByName(object.BuiltinBaseType, "to_list"),
	"abs_path":           object.GetBuiltinByName(object.BuiltinBaseType, "abs_path"),
	"fmt":                object.GetBuiltinByName(object.BuiltinBaseType, "fmt"),
	"save":               object.GetBuiltinByName(object.BuiltinBaseType, "save"),
	"__hash":             object.GetBuiltinByName(object.BuiltinBaseType, "__hash"),
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
