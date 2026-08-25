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
