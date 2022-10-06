import db

var x = db.open();
println("AFTER OPEN");
x.ping();
println("AFTER PING");
# TODO: Handle BLOB (may need a new datatype in blue)
# Also Note that NULL isnt a real datatype so its kind of like 'any'
val result = x.execute(
    """drop table if exists t;
    create table t(i INTEGER, a TEXT, b REAL, c NULL);
    insert into t values(42, 'Hello', 3.14159, null), (314, 'World!', 0.09991234, 'asf');
    """);
println("execute result = #{result}");
if (result != {last_insert_id: 2, rows_affected: 2}) {
    false
}

val result_query_no_named_cols = x.query("select * from t;");
println("result_query_no_named_cols = #{result_query_no_named_cols}");
val expected_result_query_no_named_cols = [[42, "Hello", 3.141590, null], [314, "World!", 0.099912, "asf"]];
if (result_query_no_named_cols != expected_result_query_no_named_cols) {
    false
}

val result_query_named_cols = x.query("select * from t;", named_cols=true);
println("result_query_named_cols = #{result_query_named_cols}");
val expected_result_query_named_cols = [{i: 42, a: "Hello", b: 3.141590, c: null}, {i: 314, a: "World!", b: 0.099912, c: "asf"}];
if (result_query_named_cols != expected_result_query_named_cols) {
    false
}

println("AFTER QUERY");
println("END");
true;