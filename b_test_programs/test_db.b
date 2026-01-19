import db

var x = db.open();
println("AFTER OPEN");
x.ping();
println("AFTER PING");
# Also Note that NULL isnt a real datatype so its kind of like 'any'
val bytes_example = "Hello World!".to_bytes();
val result = x.execute(
    """drop table if exists t;
    create table t(i INTEGER, a TEXT, b REAL, c NULL, d BLOB);
    insert into t values(42, 'Hello', 3.14159, null, ?), (314, 'World!', 0.09991234, 'asf', ?);
    """, [bytes_example, bytes_example]);
println("execute result = #{result}");
assert(result == {last_insert_id: 2, rows_affected: 2})

val result_query_no_named_cols = x.query("select * from t;");
val expected_result_query_no_named_cols = [[42, "Hello", 3.14159, null, bytes_example], [314, "World!", 0.09991234, "asf", bytes_example]];
assert(result_query_no_named_cols == expected_result_query_no_named_cols);

val result_query_with_query_params = x.query("select * from t where i = ?;", [42]);
println("result_query_with_query_params = #{result_query_with_query_params}");
val expected_result_query_with_query_params = [[42, "Hello", 3.14159, null, bytes_example]];
assert(result_query_with_query_params == expected_result_query_with_query_params)

val result_query_with_query_params_named_cols = x.query("select * from t where i = ?;", [42], named_cols=true);
println("result_query_with_query_params = #{result_query_with_query_params}");
val expected_result_query_with_query_params_named_cols = [{i: 42, a: "Hello", b: 3.14159, c: null, d: bytes_example}];
assert(result_query_with_query_params_named_cols == expected_result_query_with_query_params_named_cols)

val result_query_with_query_params_default_param = x.query("select * from t where i = ?;", query_args=[42]);
println("result_query_with_query_params = #{result_query_with_query_params}");
val expected_result_query_with_query_params_default_param = [[42, "Hello", 3.14159, null, bytes_example]];
assert(result_query_with_query_params_default_param == expected_result_query_with_query_params_default_param)

val result_query_named_cols = x.query("select * from t;", named_cols=true);
println("result_query_named_cols = #{result_query_named_cols}");
val expected_result_query_named_cols = [{i: 42, a: "Hello", b: 3.14159, c: null, d: bytes_example}, {i: 314, a: "World!", b: 0.09991234, c: "asf", d: bytes_example}];
if (result_query_named_cols != expected_result_query_named_cols) {
    assert(false)
}

println("AFTER QUERY");
println("END");
assert(true);