val topic1 = 'a';
val topic2 = 'b';


for (x in 1..10) {
    KV.put(topic1, x, x ** 2);
}


for (x in 10..20) {
    KV.put(topic2, x, x + 2);
}

for (x in 1..10) {
    assert(KV.get(topic1, x) == x ** 2);
}
for (x in 10..20) {
    assert(KV.get(topic2, x) == x + 2)
}

KV.delete(topic1);
for (x in 1..10) {
    assert(KV.get(topic1, x) == null);
}

KV.delete(topic2, 10);
assert(KV.get(topic2, 10) == null);
for (x in 11..20) {
    assert(KV.get(topic2, x) == x + 2)
}
