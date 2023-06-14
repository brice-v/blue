import time

val current = time.now();
println("current = #{current}");
assert(current > 1686707897417);
val parsed = time.parse("now");
println("parsed = #{parsed}");
assert(parsed > 1686707897417);
val as_a_str = time.to_str(parsed, "America/New_York");
val as_a_str1 = time.to_str(parsed, time.timezone.NewYork);
println("as_a_str = #{as_a_str}, len(as_a_str) = #{len(as_a_str)}");
println("as_a_str1 = #{as_a_str1}, len(as_a_str1) = #{len(as_a_str1)}");
assert(len(as_a_str) == 23)
assert(len(as_a_str1) == 23)
