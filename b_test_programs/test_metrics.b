###
val metrics = runtime.go_metrics();
println("metrics = #{metrics}");

val metrics_flat = runtime.go_metrics(flat=true);
println("metrics_flat = #{metrics_flat}");
###

assert(true);