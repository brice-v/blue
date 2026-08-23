package vm

import (
	"testing"
)

// Programs isolating the allocation patterns a small-string cache could
// improve (mirrors the small-int cache idea for strings).
var stringCacheBenchPrograms = map[string]string{
	// Repeated single-rune extraction: every s[i] allocates a fresh
	// 1-char Stringo today.
	"char-index": `
fun main() {
    val s = "the quick brown fox jumps over the lazy dog and runs away fast";
    var hits = 0;
    for i in 0..50000 {
        if s[i % len(s)] == "e" {
            hits += 1;
        }
    }
    return hits;
}
main()
`,
	// String ranges materialize a list whose elements are 1-char
	// Stringo objects ("a".."z" -> 26 per iteration here).
	"char-range": `
fun main() {
    var total = 0;
    for i in 0..2000 {
        val letters = "a".."z";
        total += len(letters);
    }
    return total;
}
main()
`,
	// str() of small integers: CustomInspect formats digits (still
	// allocated) but the wrapping Stringo is cacheable.
	"str-small-ints": `
fun main() {
    var acc = "";
    for i in 0..20000 {
        acc = str(i % 10);
    }
    return len(acc);
}
main()
`,
	// Slicing a string: every s[a..<b] materialized a full []rune of the
	// source plus a rune slice for the result.
	"string-slice": `
fun main() {
    val s = "the quick brown fox jumps over the lazy dog";
    var total = 0;
    for i in 0..20000 {
        total += len(s[0..<4]);
    }
    return total;
}
main()
`,
	// str() of multi-digit ints: FormatInt allocated digits on every call
	// and the wrapper was fresh each time; interning makes repeats free.
	"str-counters": `
fun main() {
    var total = 0;
    for i in 0..20000 {
        total += len(str(i % 1000));
    }
    return total;
}
main()
`,
	// Repeated short concat keys ("user-0".."user-49"): the classic
	// loop-built map key pattern from the profiling notes.
	"map-key-reuse": `
fun main() {
    var m = {};
    for i in 0..20000 {
        val k = "user-" + str(i % 50);
        m[k] = i;
    }
    return len(m);
}
main()
`,
}

func BenchmarkBlueStringScenarios(b *testing.B) {
	for name, src := range stringCacheBenchPrograms {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = compileAndRunBench(b, src)
			}
		})
	}
}
