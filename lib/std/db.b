val __db_open = _db_open;
fun open(db_name=":memory:") {
    __db_open(db_name)
}
val db_ping = _db_ping;
val db_exec = _db_exec;
val db_query = _db_query;
val db_close = _db_close;
