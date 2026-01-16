package utils

import (
	"blue/code"
	"blue/object"
	"log"
	"math"
)

func CheckOverflow(leftVal, rightVal int64) bool {
	result := leftVal + rightVal
	return result-leftVal != rightVal
}

func CheckUnderflow(leftVal, rightVal int64) bool {
	result := leftVal - rightVal
	return result+rightVal != leftVal
}

func CheckOverflowMul(leftVal, rightVal int64) bool {
	if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
		return false
	}
	if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
		return true
	}
	result := leftVal * rightVal
	return result/rightVal != leftVal
}

func CheckOverflowPow(leftVal, rightVal int64) bool {
	if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
		return false
	}
	if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
		return true
	}
	if rightVal > 63 && leftVal > 1 {
		return true
	}
	return false
}

func IfNameInMapSetEnv(env *object.Environment, m object.OrderedMap2[object.HashKey, object.MapPair], name string) bool {
	for _, k := range m.Keys {
		mp, _ := m.Get(k)
		if mp.Key.Type() == object.STRING_OBJ {
			s := mp.Key.(*object.Stringo).Value
			if name == s {
				env.Set(name, mp.Value)
				return true
			}
		}
	}
	return false
}

func GetNextOpCallPos(ins code.Instructions, ip int) int {
	i := ip
	for i < len(ins) {
		def, err := code.Lookup(ins[i])
		if err != nil {
			log.Fatalf("UNREACHABLE - failed to lookup instruction")
		}
		if def.Name == "OpCall" {
			return i
		}
		_, read := code.ReadOperands(def, ins[i+1:])
		i += 1 + read
	}
	return -1
}
