val __open = _open;
fun open(db_name=":memory:") {
    return __open(db_name);
}

val ping_ = _ping;
val close_ = _close;
