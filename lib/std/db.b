## `db` is the module that contains functions needed to
## operate on a database with. ping, exec, query, and close
##
## It is currently a sqlite backend that is completely
## compiled in go so it is slower than those compiled
## in c

val __db_open = _db_open;
val __db_ping = _db_ping;
val __db_exec = _db_exec;
val __db_query = _db_query;
val __db_close = _db_close;

fun db_ping(db_id) {
    ## `db_ping` pings the DB connection and returns null if successful
    ##
    ## db_ping(db_id: uint) -> null
    __db_ping(db_id)
}

fun db_exec(db_id, exec_query, query_args) {
    ## `db_exec` will execute the statment on the DB connection and return an object
    ## with the last_insert_id as well as the rows_affected
    ##
    ## query_args is a list of any of the following, str|int|float|null|bool|bytes
    ##
    ## db_exec(db_id: uint, exec_query: str, query_args: list[any]) -> {last_insert_id: int, rows_affected: int}
    __db_exec(db_id, exec_query, query_args)
}

fun db_query(db_id, query_s, query_args, named_cols) {
    ## `db_query` will execute the given query with the parameters set on the DB connection
    ## and it will return a list of objects or lists depending if named_cols is true
    ##
    ## query_args is a list of any of the following, str|int|float|null|bool|bytes
    ## named_cols if true will add the column name to each of the rows returned
    ##            and it will return a list of objects
    ##
    ## db_query(db_id: uint, query_s: str, query_args: list[any], named_cols: bool) ->
    ##         list[list[any]] or list[map[str:any]]
    __db_query(db_id, query_s, query_args, named_cols)
}

fun db_close(db_id) {
    ## `db_close` will close the DB connection object and remove it
    ##
    ## db_close(db_id: uint) -> null
    __db_close(db_id)
}

fun open(db_name=":memory:") {
    ## `open` will return a core db object that can be used
    ## to execute sql queries against
    ##
    ## open(db_name: str=":memory") -> DB_OBj
    var this = {};
    this._db = __db_open(db_name);
    this.ping = fun() {
        db_ping(this._db)
    };
    this.execute = fun(exec_query, query_args=[]) {
        db_exec(this._db, exec_query, query_args)
    };
    this.query = fun(query_s, query_args=[], named_cols=false) {
        db_query(this._db, query_s, query_args, named_cols)
    };
    this.close = fun() {
        db_close(this._db);
    };
    return this;
}