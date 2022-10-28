val __open = _open;
fun open(db_name=":memory:") {
    __open(db_name)
}
val ping_ = _ping;
val exec_ = _exec;
val query_ = _query;
val close_ = _close;
