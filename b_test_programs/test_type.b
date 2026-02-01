val t1 = type(1);
val t2 = type("");
val t3 = type({});
val t4 = type([]);
val t5 = type(set([]));
val t6 = type(fun() {});
val t7 = type(null);
val t8 = type((2 ** 100));
val t9 = type(0x1234);
val t10 = type(true);
val t11 = type(0.1234);
# Only MAP_COMP_OBJ is a separate type according to this
val t12 = type([x for (x in 1..10)]);
val t13 = type({x for (x in 1..10)});
val t14 = type({x: "" for (x in 1..10)});

println("t1=#{t1}, t2=#{t2}, t3=#{t3}, t4=#{t4}, t5=#{t5}, t6=#{t6}, t7=#{t7}, t8=#{t8}, t9=#{t9}, t10=#{t10}, t11=#{t11}, t12=#{t12}, t13=#{t13}, t14=#{t14}");

assert(t1 == "INTEGER");
assert(t2 == "STRING");
assert(t3 == "MAP");
assert(t4 == "LIST");
assert(t5 == "SET");
# This doesnt work for some reason
assert(t6 == "FUNCTION" || t6 == "CLOSURE_OBJ");
assert(t7 == "NULL");
assert(t8 == "BIG_INTEGER");
assert(t9 == "UINTEGER");
assert(t10 == "BOOLEAN");
assert(t11 == "FLOAT");