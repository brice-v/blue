package vm

import (
	"blue/blueutil"
	"blue/code"
	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/token"
	"testing"
)

// benchPrograms are representative workloads used for profiling.
var benchPrograms = map[string]string{
	// Deep recursion: function call overhead + int arithmetic + comparisons
	"fib20": `
fun fib(n) {
    if n < 2 {
        return n;
    }
    return fib(n-1) + fib(n-2);
}
fib(20)
`,
	// Tight loop: local var arithmetic, comparison, jump (range-based for)
	"sum-loop": `
fun main() {
    var total = 0;
    for i in 0..20000 {
        total = total + i;
    }
    return total;
}
main()
`,
	// Same loop written C-style: isolates for-in range re-evaluation cost
	"c-for-loop": `
fun main() {
    var total = 0;
    for (var i = 0; i <= 20000; i += 1) {
        total = total + i;
    }
    return total;
}
main()
`,
	// While-style loop via `for cond`
	"while-loop": `
fun main() {
    var i = 0;
    var total = 0;
    for i < 100000 {
        total += i;
        i += 1;
    }
    return total;
}
main()
`,
	// String building: concat
	"strings": `
fun main() {
    var s = "";
    for i in 0..5000 {
        s = s + "x";
    }
    return len(s);
}
main()
`,
	// List: build via lshift append, index reads, iterate
	"list-ops": `
fun main() {
    var l = [];
    for i in 0..10000 {
        l << i;
    }
    var total = 0;
    for v in l {
        total += v;
    }
    return total;
}
main()
`,
	// Map: insert + lookup + iterate keys
	"map-ops": `
fun main() {
    var m = {};
    for i in 0..5000 {
        m["key#{i}"] = i;
    }
    var total = 0;
    for k in m.keys() {
        total += m[k];
    }
    return total;
}
main()
`,
	// Higher-order builtins: map + filter call a lambda per element via
	// applyFunctionFast, the dominant pattern in idiomatic blue scripts.
	"hof-map-filter": `
fun main() {
    val l = [x for (x in 0..2000)];
    val doubled = l.map(|x| => x * 2);
    val evens = doubled.filter(|x| => x % 4 == 0);
    var total = 0;
    for v in evens {
        total += v;
    }
    return total;
}
main()
`,
	// List comprehension with a filter condition
	"list-comp": `
fun main() {
    var l = [x * x for (x in 0..10000) if (x % 3 == 0)];
    var total = 0;
    for v in l {
        total += v;
    }
    return total;
}
main()
`,
	// String interpolation churn: two substitutions per iteration plus join
	"string-interp": `
fun main() {
    var parts = [];
    for i in 0..5000 {
        parts << "item #{i} at #{i * 7}";
    }
    var joined = parts.join(", ");
    return len(joined);
}
main()
`,
	// Closure state: free-variable reads + method-call sugar index set,
	// called through an OpCall with zero args every iteration.
	"closure-counter": `
fun makeCounter() {
    var state = {count: 0};
    return fun() {
        state.count += 1;
        return state.count;
    };
}
fun main() {
    var counter = makeCounter();
    var total = 0;
    for i in 0..20000 {
        total += counter();
    }
    return total;
}
main()
`,
	// Struct literals: build + field set + field get each iteration
	"struct-ops": `
fun main() {
    var total = 0;
    for i in 0..20000 {
        var p = @{x: i, y: i * 2};
        p.x = p.x + 1;
        p.y = p.y + p.x;
        total += p.y;
    }
    return total;
}
main()
`,
	// Match dispatch over integer literal arms
	"match-dispatch": `
fun classify(n) {
    return match n {
        0 => { 100 },
        1 => { 101 },
        2 => { 102 },
        3 => { 103 },
        _ => { 999 },
    };
}
fun main() {
    var total = 0;
    for i in 0..19999 {
        total += classify(i % 5);
    }
    return total;
}
main()
`,
	// Native sort of LCG-generated ints: exercises int boxing beyond the
	// small-int cache and the native sort path.
	// NOTE: uses _sort directly because `sort` is a core.b wrapper and this
	// harness compiles without the core library (see compiler.New()).
	"sort-ints": `
fun main() {
    var l = [];
    var seed = 123456789;
    for i in 0..20000 {
        seed = (seed * 1103515245 + 12345) % 2147483648;
        l << seed;
    }
    val s = _sort(l, false, null);
    return s[10000];
}
main()
`,
	// Sort of map records with a key lambda: applyFunctionFast runs inside
	// the sort comparator (O(n log n) callback invocations).
	"sort-records": `
fun main() {
    var users = [];
    var seed = 42;
    for i in 0..1500 {
        seed = (seed * 1103515245 + 12345) % 2147483648;
        users << {name: "u#{i}", age: seed % 100};
    }
    val sorted = _sort(users, false, |user| => user.age);
    var total = 0;
    for u in sorted {
        total += u.age;
    }
    return total;
}
main()
`,
	// Big integer arithmetic: int64 overflow promotes to BigInteger which
	// then flows through mixed-type ops and modulus.
	"bigint-ops": `
fun main() {
    var x = 9223372036854775807;
    var total = 0;
    for i in 0..5000 {
        x = x + 9223372036854775807;
        total += x % 1000003;
    }
    return total;
}
main()
`,
}

func compileAndRunBench(b *testing.B, src string) object.Object {
	b.Helper()
	blueutil.ENABLE_VM_CACHING = true
	l := lexer.New(src, "<bench>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		b.Fatalf("parser errors: %v", p.ErrorMessages())
	}
	c := compiler.New()
	if err := c.Compile(prog); err != nil {
		b.Fatalf("compile error: %s", err)
	}
	globals := make([]object.Object, GlobalsSize)
	v := NewWithGlobalsStore(c.Bytecode(), globals)
	if err := v.Run(); err != nil {
		b.Fatalf("vm error: %s", err)
	}
	return v.LastPoppedStackElem()
}

func BenchmarkBlueEndToEnd(b *testing.B) {
	for name, src := range benchPrograms {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = compileAndRunBench(b, src)
			}
		})
	}
}

func BenchmarkBlueFrontendOnly(b *testing.B) {
	for name, src := range benchPrograms {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				l := lexer.New(src, "<bench>")
				p := parser.New(l)
				prog := p.ParseProgram()
				c := compiler.New()
				if err := c.Compile(prog); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBlueVmRunOnly(b *testing.B) {
	for name, src := range benchPrograms {
		b.Run(name, func(b *testing.B) {
			// Compile once outside the timer
			l := lexer.New(src, "<bench>")
			p := parser.New(l)
			prog := p.ParseProgram()
			c := compiler.New()
			if err := c.Compile(prog); err != nil {
				b.Fatal(err)
			}
			bc := c.Bytecode()
			b.ResetTimer()
			b.ReportAllocs()
			for range b.N {
				globals := make([]object.Object, GlobalsSize)
				v := NewWithGlobalsStore(bc, globals)
				if err := v.Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkLexerOnly(b *testing.B) {
	for name, src := range benchPrograms {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				l := lexer.New(src, "<bench>")
				for {
					tok := l.NextToken()
					if tok.Type == token.EOF {
						break
					}
				}
			}
		})
	}
}

// BenchmarkOpcodeDispatch isolates raw interpreter dispatch cost on a long
// straight-line instruction stream of pushes and pops.
func BenchmarkOpcodePushPop(b *testing.B) {
	ins := code.Instructions{}
	for range 500000 {
		ins = append(ins, byte(code.OpTrue), byte(code.OpPop))
	}
	mainFn := &object.CompiledFunction{Instructions: ins}
	cl := &object.Closure{Fun: mainFn}
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		v := &VM{
			stack:       make([]object.Object, StackSize),
			frames:      make([]Frame, MaxFrames),
			framesIndex: 1,
			lastNodePos: -1,
		}
		v.frames[0] = *NewFrame(cl, 0)
		_ = v.Run()
	}
}

// BenchmarkFib25 measures the classic call/arith heavy workload with a
// correctness check on the result.
func BenchmarkFib25(b *testing.B) {
	src := `
fun fib(n) {
    if n < 2 {
        return n;
    }
    return fib(n-1) + fib(n-2);
}
fib(25)
`
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		r := compileAndRunBench(b, src)
		i, ok := r.(*object.Integer)
		if !ok || i.Value != 75025 {
			b.Fatalf("bad result %v", r)
		}
	}
}

// BenchmarkProcessSpawnE2E measures the concurrency machinery: spawning 16
// child processes, each sending one message back to the parent via channels.
// Kept out of benchPrograms because it is concurrent (goroutine churn) and
// would pollute the compile-only benchmarks.
func BenchmarkProcessSpawnE2E(b *testing.B) {
	src := `
fun child(parent, n) {
    parent.send(n);
}

fun main() {
    val me = self();
    var pids = [];
    for i in 0..15 {
        pids << spawn(child, [me, i]);
    }
    var total = 0;
    for i in 0..15 {
        total += me.recv();
    }
    return total;
}
main()
`
	l := lexer.New(src, "<bench>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		b.Fatalf("parser errors: %v", p.ErrorMessages())
	}
	c := compiler.New()
	if err := c.Compile(prog); err != nil {
		b.Fatal(err)
	}
	bc := c.Bytecode()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		v := New(bc)
		if err := v.Run(); err != nil {
			b.Fatal(err)
		}
		r := v.LastPoppedStackElem()
		i, ok := r.(*object.Integer)
		if !ok || i.Value != 120 {
			b.Fatalf("bad result %v", r)
		}
	}
}
