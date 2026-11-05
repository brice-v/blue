package object

func init() {
	for _, module := range AllBuiltins {
		for _, builtin := range module.Builtins {
			if builtin == nil || builtin.Module != "" {
				continue
			}
			builtin.Module = module.Name
		}
	}

	for i := range BuiltinobjsList {
		entry := &BuiltinobjsList[i]
		if entry.Builtin == nil {
			continue
		}
		entry.Builtin.Module = "builtinobj"
		entry.Builtin.Name = entry.Name
	}
}

func builtinIdentity(obj Object) (module, name string, ok bool) {
	switch builtin := obj.(type) {
	case *Builtin:
		if builtin.Module == "" || builtin.Name == "" {
			return "", "", false
		}
		return builtin.Module, builtin.Name, true
	case *BuiltinObj:
		if builtin.Module == "" || builtin.Name == "" {
			return "", "", false
		}
		return builtin.Module, builtin.Name, true
	}
	return "", "", false
}
