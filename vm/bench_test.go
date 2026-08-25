package vm

import (
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
	// Broad mixed workload covering most language features in one program:
	// recursion, match dispatch with default args, closures over struct
	// state, struct field get/set, map insert/lookup/key-iteration with
	// interpolated keys, set literals/add/membership, list append/index/
	// slice, list comprehension with filter, HOF map/filter lambdas,
	// char-range iteration, string interpolation/concat/slice/index, float
	// math, bigint promotion, try/catch/finally unwinding, defer callbacks,
	// record sort with key lambda, and all three loop forms.
	"mixed-workload": `
fun warm(n) {
    if n < 2 {
        return n;
    }
    return warm(n-1) + warm(n-2);
}

fun classify(ch, base=1000) {
    return match ch {
        0 => { base + 1 },
        1 => { base + 2 },
        2 => { base + 4 },
        _ => { base + 8 },
    };
}

fun makeCounter(start) {
    var state = @{count: start, steps: 0};
    return fun(step) {
        state.count += step;
        state.steps += 1;
        return state.count;
    };
}

fun withDefer(acc, n) {
    val dfun = fun() {
        acc(n);
    };
    defer(dfun);
    return n * 2;
}

fun processBatch(items) {
    val doubled = items.map(|x| => x * 2);
    val filtered = doubled.filter(|x| => x % 3 != 0);
    var total = 0;
    for v in filtered {
        total += v;
    }
    return total;
}

fun main() {
    var checksum = 0;

    checksum += warm(13);

    val counter = makeCounter(0);
    for i in 0..2999 {
        checksum += classify(i % 5);
        checksum += counter(1) % 7;
    }

    val records = [];
    var seed = 12345;
    for i in 0..1499 {
        seed = (seed * 1103515245 + 12345) % 2147483648;
        var r = @{id: i, score: seed % 1000, name: "u#{i}"};
        r.score = r.score + i % 10;
        records << r;
        checksum += r.score % 13;
    }

    var byName = {};
    for r in records {
        byName[r.name] = r.score;
    }
    for k in byName.keys() {
        checksum += byName[k] % 11;
    }

    val unique = {1, 1, 2, 3};
    for r in records[100..200] {
        unique << r.id % 50;
        if (r.id % 50) in unique {
            checksum += 1;
        }
    }
    for v in unique {
        checksum += v;
    }

    checksum += processBatch([x for (x in 0..999) if (x % 2 == 0)]);

    var s = "";
    for ch in 'a'..'f' {
        s = s + ch;
    }
    checksum += len(s);
    checksum += len(s[1..<4]);
    if s[2] == 'c' {
        checksum += 1;
    }

    var fsum = 0.0;
    var i = 0;
    for i < 2000 {
        fsum += i * 0.5;
        i += 1;
    }
    if fsum > 0.0 {
        checksum += 1;
    }

    var caught = 0;
    for i in 0..49 {
        try {
            val z = 100 / (i - 25);
            checksum += z % 3;
        } catch (e) {
            caught += 1;
        } finally {
            checksum += 1;
        }
    }
    checksum += caught;

    var big = 9223372036854775807;
    for i in 0..99 {
        big = big + big;
        if big % 1000000007 >= 0 {
            checksum += 1;
        }
    }

    val rows = [];
    for r in records {
        rows << {sid: r.id, sc: r.score};
    }
    val sorted = _sort(rows, false, |r| => r.sc);
    checksum += sorted[750].sc;

    var ctotal = 0;
    for (var j = 0; j < 3000; j += 1) {
        ctotal += j % 17;
    }
    checksum += ctotal;

    for i in 0..299 {
        checksum += withDefer(counter, i % 7);
    }

    return checksum;
}
main()
`,
}

func compileAndRunBench(b *testing.B, src string) object.Object {
	b.Helper()
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
			for b.Loop() {
				_ = compileAndRunBench(b, src)
			}
		})
	}
}

func BenchmarkBlueFrontendOnly(b *testing.B) {
	for name, src := range benchPrograms {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
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
			for b.Loop() {
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
			for b.Loop() {
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

	b.ReportAllocs()
	for b.Loop() {
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

// BenchmarkMixedWorkload is the primary end-to-end benchmark: one program
// exercising recursion, calls, closures, structs, maps, sets, lists, slices,
// comprehensions, HOFs, strings, floats, bigints, try/catch/finally unwinding,
// defer and all loop forms, with a pinned checksum to catch miscompiles.
func BenchmarkMixedWorkload(b *testing.B) {
	src := benchPrograms["mixed-workload"]

	b.ReportAllocs()
	for b.Loop() {
		r := compileAndRunBench(b, src)
		i, ok := r.(*object.Integer)
		if !ok || i.Value != 4432407 {
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

	b.ReportAllocs()
	for b.Loop() {
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
