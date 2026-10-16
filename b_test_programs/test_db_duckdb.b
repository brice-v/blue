# VM IGNORE
# TODO: Replace above with STATIC IGNORE once its working (it still cant be run for statically built blue)
import db
import csv
import time


# Scaffolding for the duckdb backend of the db module. It is written against the
# api the db module is meant to have rather than the one it has today, so a good
# number of these checks are expected to fail until the duckdb type mapping, the
# file format support, and first class time values land.
#
# Nothing stops at the first failure, every miss is collected and printed as one
# report at the end so the whole gap list is visible in a single run.
#
# Sections:
#   1  connection lifecycle and open options
#   2  execute/query api (params, named cols, result shapes)
#   3  duckdb scalar datatypes
#   4  duckdb nested datatypes
#   5  duckdb temporal datatypes
#   6  binding every blue type as a parameter
#   7  csv (read_csv, read_csv_auto, COPY TO, globs, csv module interop)
#   8  parquet (COPY TO, read_parquet, metadata, partitioning)
#   9  json (extension, read_json_auto, COPY TO, json sql functions)
#  10  extensions and remote sources
#  11  sql feature smoke tests
#  12  transactions
#  13  schemas, attach/detach, export/import
#  14  catalog and settings introspection
#  15  error handling
#  16  interop with the rest of blue
#  17  cleanup

# every path lives on this map, bare top level vals get clobbered once a file
# has this many of them
var f = {};
f.workdir = temp_dir("", "blue_test_db_duckdb_*");

fun fp(name) {
    path.join(f.workdir, name)
}

fun open_duck(db_name=":memory:") {
    db.open(db_name=db_name, db_type="duckdb")
}

# Nothing in here aborts on the first problem. A failing assert unwinds to the
# nearest enclosing try, which would silently skip every section after it, so
# every check records into this report and the full inventory of what is still
# missing gets printed at the end.
var report = {failures: []};

fun fail(msg) {
    report.failures << msg
}

fun check(cond, msg="assertion failed") {
    if (cond != true) {
        fail(msg)
    }
    return cond
}

# The helpers swallow the error and record it, a raise from deep inside a
# section would otherwise skip the rest of the file
fun q(conn, sql, args=[], named_cols=false) {
    try {
        return conn.query(sql, args, named_cols=named_cols)
    } catch (e) {
        fail("query error: #{e}")
        return []
    }
}

fun qmap(conn, sql, args=[]) {
    try {
        return conn.query(sql, args, named_cols=true)
    } catch (e) {
        fail("query error: #{e}")
        return []
    }
}

# first column of the first row, null when the query returned nothing
fun q_one(conn, sql, args=[]) {
    val rows = q(conn, sql, args);
    if (len(rows) == 0) {
        return null
    }
    return rows[0][0]
}

# first row as a map, an empty map when the query returned nothing
fun q1(conn, sql, args=[]) {
    val rows = qmap(conn, sql, args);
    if (len(rows) == 0) {
        return {}
    }
    return rows[0]
}

# count(*) that never hands back null, a NULL from the driver becomes -1
fun cnt(conn, sql, args=[]) {
    val v = q_one(conn, sql, args);
    if (v == null) {
        return -1
    }
    return v
}

# index that hands back an empty map instead of raising on a short list
fun at(list, i) {
    if (len(list) <= i) {
        return {}
    }
    return list[i]
}

# message of the error a query raises, null when it succeeds
fun query_error(conn, sql, args=[]) {
    try {
        conn.query(sql, args);
    } catch (e) {
        return "#{e}"
    }
    return null
}

fun exec_error(conn, sql, args=[]) {
    try {
        conn.execute(sql, args);
    } catch (e) {
        return "#{e}"
    }
    return null
}

# runs a statement that returns no rows and hands back {last_insert_id, rows_affected}
fun exec(conn, sql, args=[]) {
    try {
        return conn.execute(sql, args)
    } catch (e) {
        fail("exec error: #{e}")
        return {last_insert_id: -1, rows_affected: -1}
    }
}

# prints what duckdb handed back for each datatype and records the misses
fun check_types(conn, cases) {
    for (tc in cases) {
        val sql = "select cast(#{tc.literal} as #{tc.sql_type}) as v";
        var got = null;
        try {
            val rows = conn.query(sql, [], named_cols=true);
            if (len(rows) > 0) {
                got = rows[0].v;
            }
        } catch (e) {
            got = "ERROR #{e}";
        }
        println("#{tc.label} (#{tc.sql_type}) => #{got} [#{type(got)}]");
        if (type(got) != tc.want_type) {
            fail("#{tc.label}: want blue type #{tc.want_type}, got #{type(got)} (#{got})");
        } else if (tc.want != null && got != tc.want) {
            fail("#{tc.label}: want #{tc.want}, got #{got}");
        }
    }
}

# records a statement that does not run, prints what came back
fun smoke(conn, label, sql) {
    try {
        val rows = conn.query(sql, [], named_cols=true);
        println("smoke #{label} => #{rows}");
        return true
    } catch (e) {
        println("smoke #{label} FAILED => #{e}");
        fail("smoke #{label}: #{e}")
        return false
    }
}

fun smoke_all(conn, cases) {
    for (c in cases) {
        smoke(conn, c.label, c.sql);
    }
}

fun print_report() {
    println("");
    println("==== duckdb scaffolding: #{len(report.failures)} failing checks ====");
    for (msg in report.failures) {
        println("  - #{msg}");
    }
}


# ---------------------------------------------------------------------------
# 1. connection lifecycle and open options
# ---------------------------------------------------------------------------

check(exists(f.workdir), "exists(f.workdir)");

var conn = open_duck();
check(conn != null, "conn != null");
check(conn.ping() == null, "conn.ping() == null");

# sqlite stays the default and keeps working alongside duckdb
var sqlite_conn = db.open();
check(sqlite_conn.ping() == null, "sqlite_conn.ping() == null");
check(q_one(sqlite_conn, "select sqlite_version()") != null, "q_one(sqlite_conn, \"select sqlite_version()\") != null");
sqlite_conn.close();

var sqlite_named = db.open(db_name=":memory:", db_type="sqlite");
check(sqlite_named.ping() == null, "sqlite_named.ping() == null");
sqlite_named.close();

# unknown backends are rejected
try {
    db.open(db_name=":memory:", db_type="not_a_db");
    check(false, "db_type not_a_db should not open");
} catch (e) {
    println("bad db_type => #{e}");
}

# duckdb takes its config through the dsn query string
var configured = db.open(db_name=":memory:?access_mode=READ_WRITE", db_type="duckdb");
check(configured.ping() == null, "configured.ping() == null");
configured.close();

var threaded = db.open(db_name=":memory:?threads=2", db_type="duckdb");
check(q_one(threaded, "select current_setting('threads')") == 2, "q_one(threaded, \"select current_setting('threads')\") == 2");
threaded.close();

# file backed database, written to, closed, and reopened to prove it persisted
f.file_db = fp("persist.db");
var writer = open_duck(f.file_db);
exec(writer, "create table persisted(id integer, label varchar)");
check(exec(writer, "insert into persisted values (1, 'one'), (2, 'two')").rows_affected == 2, "exec(writer, \"insert into persisted values (1, 'one'), (2, 'two')\").rows_affected == 2");
check(exec(writer, "insert into persisted values (3, 'three')").rows_affected == 1, "exec(writer, \"insert into persisted values (3, 'three')\").rows_affected == 1");
writer.close();
check(exists(f.file_db), "exists(f.file_db)");

var reader = open_duck(f.file_db);
check(q_one(reader, "select count(*) from persisted") == 3, "q_one(reader, \"select count(*) from persisted\") == 3");
reader.close();

# read only access mode rejects writes but still serves reads
f.ro_db = fp("read_only.db");
var ro_writer = open_duck(f.ro_db);
exec(ro_writer, "create table ro_t(a integer)");
exec(ro_writer, "insert into ro_t values (1)");
ro_writer.close();

var ro_conn = db.open(db_name="#{f.ro_db}?access_mode=READ_ONLY", db_type="duckdb");
check(q_one(ro_conn, "select a from ro_t") == 1, "q_one(ro_conn, \"select a from ro_t\") == 1");
val ro_err = exec_error(ro_conn, "insert into ro_t values (2)");
check(ro_err != null, "writes must fail against a READ_ONLY database");
println("read only write => #{ro_err}");
ro_conn.close();

# using a closed connection has to surface an error, not a panic
var closed = open_duck();
closed.close();
check(query_error(closed, "select 1") != null, "query on a closed db must error");
check(exec_error(closed, "create table nope(a int)") != null, "exec on a closed db must error");


# ---------------------------------------------------------------------------
# 2. execute/query api
# ---------------------------------------------------------------------------

check(conn.ping() == null, "conn.ping() == null");

exec(conn, """
    drop table if exists users;
    create table users(
        id integer primary key,
        name varchar not null,
        age integer,
        height float,
        awesome boolean,
        balance decimal(12,2),
        bday date,
        joined timestamp,
        payload blob,
        tags varchar[],
        meta json
    );
""");

val insert_result = exec(conn, """
    insert into users values
        (1, 'marc', 99, 1.91, true, 10.50, '1970-01-01', '1970-01-01 00:00:00', 'abc', ['a','b'], '{"k":1}'),
        (2, 'macgyver', 70, 1.85, true, 20.25, '1951-01-23', '1951-01-23 00:00:00', 'def', ['c'], '{"k":2}');
""");
check(insert_result == {last_insert_id: 0, rows_affected: 2}, "insert_result == {last_insert_id: 0, rows_affected: 2}");

# duckdb has no rowid so last_insert_id stays 0, rows_affected tracks the
# statement and covers update and delete as well
check(exec(conn, "update users set age = age + 1 where name = 'marc'").rows_affected == 1, "exec(conn, \"update users set age = age + 1 where name = 'marc'\").rows_affected == 1");
check(exec(conn, "delete from users where name = 'macgyver'").rows_affected == 1, "exec(conn, \"delete from users where name = 'macgyver'\").rows_affected == 1");
exec(conn, "insert into users values (2, 'macgyver', 70, 1.85, true, 20.25, '1951-01-23', '1951-01-23 00:00:00', 'def', ['c'], '{\"k\":2}')");

# positional ? parameters
check(q_one(conn, "select * from users where name = ?", ['marc']) == 1, "q_one(conn, \"select * from users where name = ?\", ['marc']) == 1");
check(len(q(conn, "select * from users where age > ?", [80])) == 1, "len(q(conn, \"select * from users where age > ?\", [80])) == 1");

# duckdb also understands $1 style placeholders
check(q_one(conn, "select * from users where name = $1", ['macgyver']) == 2, "q_one(conn, \"select * from users where name = $1\", ['macgyver']) == 2");

# empty result sets come back as an empty list, not null
check(q(conn, "select * from users where name = 'nobody'") == [], "q(conn, \"select * from users where name = 'nobody'\") == []");
check(qmap(conn, "select * from users where name = 'nobody'") == [], "qmap(conn, \"select * from users where name = 'nobody'\") == []");

# unnamed columns come back as lists, named_cols=true comes back as maps that
# keep the order of the select list
check(q(conn, "select id, name from users order by id") == [[1, 'marc'], [2, 'macgyver']], "q(conn, \"select id, name from users order by id\") == [[1, 'marc'], [2, 'macgyver']]");
val rows_named = q(conn, "select id, name from users order by id", named_cols=true);
check(rows_named == [{id: 1, name: 'marc'}, {id: 2, name: 'macgyver'}], "rows_named == [{id: 1, name: 'marc'}, {id: 2, name: 'macgyver'}]");
check(keys(at(rows_named, 0)) == ['id', 'name'], "keys(at(rows_named, 0)) == ['id', 'name']");

# query_args can be passed positionally or by name
check(q(conn, "select * from users where id = ?", [1]) == q(conn, "select * from users where id = 1"), "q(conn, \"select * from users where id = ?\", [1]) == q(conn, \"select * from users where id = 1\")");

# column aliases become the map keys
check(q1(conn, "select 42 as the_answer").the_answer == 42, "q1(conn, \"select 42 as the_answer\").the_answer == 42");
check(len(qmap(conn, "select 1 where false")) == 0, "len(qmap(conn, \"select 1 where false\")) == 0");

# exec and query both accept the same parameter types
check(q_one(conn, "select ?", [null]) == null, "q_one(conn, \"select ?\", [null]) == null");
check(q_one(conn, "select ?", [true]) == true, "q_one(conn, \"select ?\", [true]) == true");
check(q_one(conn, "select ?", ['str']) == 'str', "q_one(conn, \"select ?\", ['str']) == 'str'");
check(q_one(conn, "select ?", [3]) == 3, "q_one(conn, \"select ?\", [3]) == 3");
check(q_one(conn, "select ?", [3.5]) == 3.5, "q_one(conn, \"select ?\", [3.5]) == 3.5");

# a multi statement exec runs every statement in order
exec(conn, "create table multi_a(a int); create table multi_b(b int); insert into multi_a values (1); insert into multi_b values (2); insert into multi_b values (3);");
check(cnt(conn, "select count(*) from multi_a") == 1, "cnt(conn, \"select count(*) from multi_a\") == 1");
check(cnt(conn, "select count(*) from multi_b") == 2, "cnt(conn, \"select count(*) from multi_b\") == 2");


# ---------------------------------------------------------------------------
# 3. duckdb scalar datatypes
# ---------------------------------------------------------------------------

val SIGNED_INT_CASES = [
    {label: 'TINYINT',       sql_type: 'TINYINT',  literal: '-128',                 want_type: Type.INT,     want: -128},
    {label: 'SMALLINT',      sql_type: 'SMALLINT', literal: '-32768',               want_type: Type.INT,     want: -32768},
    {label: 'INT2',          sql_type: 'INT2',     literal: '-32768',               want_type: Type.INT,     want: -32768},
    {label: 'INTEGER',       sql_type: 'INTEGER',  literal: '-2147483648',          want_type: Type.INT,     want: -2147483648},
    {label: 'INT',           sql_type: 'INT',      literal: '42',                   want_type: Type.INT,     want: 42},
    {label: 'INT4',          sql_type: 'INT4',     literal: '42',                   want_type: Type.INT,     want: 42},
    {label: 'SIGNED',        sql_type: 'SIGNED',   literal: '42',                   want_type: Type.INT,     want: 42},
    {label: 'BIGINT',        sql_type: 'BIGINT',   literal: '9223372036854775807',  want_type: Type.INT,     want: 9223372036854775807},
    {label: 'INT8',          sql_type: 'INT8',     literal: '-9223372036854775807', want_type: Type.INT,     want: -9223372036854775807},
    {label: 'LONG',          sql_type: 'LONG',     literal: '42',                   want_type: Type.INT,     want: 42},
    # hugeint does not fit an int64 so these have to land as BIG_INTEGER
    {label: 'HUGEINT',       sql_type: 'HUGEINT',  literal: '170141183460469231731687303715884105727',  want_type: Type.BIGINT, want: 170141183460469231731687303715884105727},
    {label: 'HUGEINT neg',   sql_type: 'HUGEINT',  literal: '-170141183460469231731687303715884105727', want_type: Type.BIGINT, want: -170141183460469231731687303715884105727},
];

val UNSIGNED_INT_CASES = [
    {label: 'UTINYINT',  sql_type: 'UTINYINT',  literal: '255',                  want_type: Type.UINT,     want: 255},
    {label: 'USMALLINT', sql_type: 'USMALLINT', literal: '65535',                want_type: Type.UINT,     want: 65535},
    {label: 'UINTEGER',  sql_type: 'UINTEGER',  literal: '4294967295',           want_type: Type.UINT,     want: 4294967295},
    # ubigint overflows int64 so these need BIG_INTEGER or a wide UINTEGER
    {label: 'UBIGINT',   sql_type: 'UBIGINT',   literal: '18446744073709551615', want_type: Type.BIGINT,   want: 18446744073709551615},
    {label: 'UHUGEINT',  sql_type: 'UHUGEINT',  literal: '340282366920938463463374607431768211455', want_type: Type.BIGINT, want: 340282366920938463463374607431768211455},
];

val FLOAT_CASES = [
    {label: 'FLOAT',      sql_type: 'FLOAT',  literal: '1.5',               want_type: Type.FLOAT, want: 1.5},
    {label: 'REAL',       sql_type: 'REAL',   literal: '1.5',               want_type: Type.FLOAT, want: 1.5},
    {label: 'FLOAT4',     sql_type: 'FLOAT4', literal: '1.5',               want_type: Type.FLOAT, want: 1.5},
    {label: 'DOUBLE',     sql_type: 'DOUBLE', literal: '3.141592653589793', want_type: Type.FLOAT, want: 3.141592653589793},
    {label: 'FLOAT8',     sql_type: 'FLOAT8', literal: '3.141592653589793', want_type: Type.FLOAT, want: 3.141592653589793},
    # a float4 widened into a blue float drags in binary noise, 1.91 comes back
    # as 1.909999966621399 today, worth pinning down once the mapping is picked
    {label: 'FLOAT lossy', sql_type: 'FLOAT', literal: '1.91',              want_type: Type.FLOAT, want: 1.91},
    {label: 'FLOAT inf',  sql_type: 'FLOAT',  literal: "'inf'",             want_type: Type.FLOAT, want: null},
    {label: 'FLOAT nan',  sql_type: 'FLOAT',  literal: "'nan'",             want_type: Type.FLOAT, want: null},
];

val DECIMAL_CASES = [
    {label: 'DECIMAL(3,0)',   sql_type: 'DECIMAL(3,0)',   literal: '5',      want_type: Type.FLOAT, want: 5},
    {label: 'DECIMAL(9,3)',   sql_type: 'DECIMAL(9,3)',   literal: '1.234',  want_type: Type.FLOAT, want: 1.234},
    {label: 'DECIMAL(18,3)',  sql_type: 'DECIMAL(18,3)',  literal: '1.234',  want_type: Type.FLOAT, want: 1.234},
    {label: 'DECIMAL(38,10)', sql_type: 'DECIMAL(38,10)', literal: '1.234',  want_type: Type.FLOAT, want: 1.234},
    {label: 'NUMERIC',        sql_type: 'NUMERIC',        literal: '1.25',   want_type: Type.FLOAT, want: 1.25},
    {label: 'DEC',            sql_type: 'DEC',            literal: '1.25',   want_type: Type.FLOAT, want: 1.25},
    {label: 'NUMERIC(38,0)',  sql_type: 'NUMERIC(38,0)',  literal: '12345678901234567890123456789012345678', want_type: Type.BIGINT, want: 12345678901234567890123456789012345678},
];

val STRING_CASES = [
    {label: 'VARCHAR',         sql_type: 'VARCHAR',  literal: "'hello'",    want_type: Type.STRING, want: 'hello'},
    {label: 'TEXT',            sql_type: 'TEXT',     literal: "'hello'",    want_type: Type.STRING, want: 'hello'},
    {label: 'STRING',          sql_type: 'STRING',   literal: "'hello'",    want_type: Type.STRING, want: 'hello'},
    {label: 'CHAR(10)',        sql_type: 'CHAR(10)', literal: "'hi'",       want_type: Type.STRING, want: 'hi'},
    {label: 'VARCHAR unicode', sql_type: 'VARCHAR',  literal: "'héllo ☃'",  want_type: Type.STRING, want: 'héllo ☃'},
    {label: 'VARCHAR empty',   sql_type: 'VARCHAR',  literal: "''",         want_type: Type.STRING, want: ''},
];

val BOOL_CASES = [
    {label: 'BOOLEAN true',     sql_type: 'BOOLEAN', literal: 'true',  want_type: Type.BOOL, want: true},
    {label: 'BOOLEAN false',    sql_type: 'BOOLEAN', literal: 'false', want_type: Type.BOOL, want: false},
    {label: 'BOOL',             sql_type: 'BOOL',    literal: 'true',  want_type: Type.BOOL, want: true},
    {label: 'BOOLEAN from int', sql_type: 'BOOLEAN', literal: '1',     want_type: Type.BOOL, want: true},
];

val BYTES_CASES = [
    {label: 'BLOB',   sql_type: 'BLOB',   literal: "'abc'", want_type: Type.BYTES, want: to_bytes('abc')},
    {label: 'BYTEA',  sql_type: 'BYTEA',  literal: "'abc'", want_type: Type.BYTES, want: to_bytes('abc')},
    {label: 'BINARY', sql_type: 'BINARY', literal: "'abc'", want_type: Type.BYTES, want: to_bytes('abc')},
];

val OTHER_SCALAR_CASES = [
    {label: 'UUID',   sql_type: 'UUID',   literal: "'123e4567-e89b-12d3-a456-426614174000'", want_type: Type.STRING, want: '123e4567-e89b-12d3-a456-426614174000'},
    {label: 'BIT',    sql_type: 'BIT',    literal: "'1010'", want_type: Type.STRING, want: '1010'},
    {label: 'JSON',   sql_type: 'JSON',   literal: '{"a": 1}', want_type: Type.MAP, want: {a: 1.0}},
    {label: 'ENUM',   sql_type: "ENUM('ok','bad')", literal: "'ok'", want_type: Type.STRING, want: 'ok'},
    {label: 'BIGNUM', sql_type: 'BIGNUM', literal: "'123456789012345678901234567890'", want_type: Type.BIGINT, want: 123456789012345678901234567890},
    # NULL is not really a datatype, it behaves like 'any'
    {label: 'NULL as INTEGER', sql_type: 'INTEGER', literal: 'null', want_type: Type.NULL, want: null},
    {label: 'NULL as VARCHAR', sql_type: 'VARCHAR', literal: 'null', want_type: Type.NULL, want: null},
];

check_types(conn, SIGNED_INT_CASES);
check_types(conn, UNSIGNED_INT_CASES);
check_types(conn, FLOAT_CASES);
check_types(conn, DECIMAL_CASES);
check_types(conn, STRING_CASES);
check_types(conn, BOOL_CASES);
check_types(conn, BYTES_CASES);
check_types(conn, OTHER_SCALAR_CASES);

# the same scalar types stored in a table rather than cast inline, the storage
# path and the scan path are separate code in the driver
val ALL_SCALAR_TYPES = [
    {name: 'c_boolean',   sql_type: 'BOOLEAN'},
    {name: 'c_tinyint',   sql_type: 'TINYINT'},
    {name: 'c_smallint',  sql_type: 'SMALLINT'},
    {name: 'c_integer',   sql_type: 'INTEGER'},
    {name: 'c_bigint',    sql_type: 'BIGINT'},
    {name: 'c_hugeint',   sql_type: 'HUGEINT'},
    {name: 'c_utinyint',  sql_type: 'UTINYINT'},
    {name: 'c_usmallint', sql_type: 'USMALLINT'},
    {name: 'c_uinteger',  sql_type: 'UINTEGER'},
    {name: 'c_ubigint',   sql_type: 'UBIGINT'},
    {name: 'c_float',     sql_type: 'FLOAT'},
    {name: 'c_double',    sql_type: 'DOUBLE'},
    {name: 'c_decimal',   sql_type: 'DECIMAL(18,3)'},
    {name: 'c_varchar',   sql_type: 'VARCHAR'},
    {name: 'c_blob',      sql_type: 'BLOB'},
    {name: 'c_uuid',      sql_type: 'UUID'},
    {name: 'c_bit',       sql_type: 'BIT'},
    {name: 'c_json',      sql_type: 'JSON'},
    {name: 'c_date',      sql_type: 'DATE'},
    {name: 'c_time',      sql_type: 'TIME'},
    {name: 'c_ts',        sql_type: 'TIMESTAMP'},
];

val ALL_SCALAR_VALUES = [
    'true', '-128', '-32768', '-2147483648', '9223372036854775807',
    '170141183460469231731687303715884105727',
    '255', '65535', '4294967295', '18446744073709551615',
    '1.5', '3.141592653589793', '1.234',
    "'hello'", "'abc'", "'123e4567-e89b-12d3-a456-426614174000'", "'1010'", '{"a": 1}',
    "'2024-02-29'", "'12:34:56'", "'2024-02-29 12:34:56.123456'",
];

fun col_defs() {
    var parts = [];
    for (col in ALL_SCALAR_TYPES) {
        parts << "#{col.name} #{col.sql_type}";
    }
    return join(parts, ", ")
}

fun null_defs() {
    var parts = [];
    for (col in ALL_SCALAR_TYPES) {
        parts << 'null';
    }
    return join(parts, ", ")
}

exec(conn, "drop table if exists all_scalars");
exec(conn, "create table all_scalars(#{col_defs()})");
check(exec(conn, "insert into all_scalars values (#{join(ALL_SCALAR_VALUES, ', ')})").rows_affected == 1, "exec(conn, \"insert into all_scalars values (${join(ALL_SCALAR_VALUES, ', ')})\").rows_affected == 1");
check(exec(conn, "insert into all_scalars values (#{null_defs()})").rows_affected == 1, "exec(conn, \"insert into all_scalars values (${null_defs()})\").rows_affected == 1");

val wide_rows = qmap(conn, "select * from all_scalars");
check(len(wide_rows) == 2, "len(wide_rows) == 2");
println("all_scalars => #{at(wide_rows, 0)}");
val wide_null_row = at(wide_rows, 1);
for (col in ALL_SCALAR_TYPES) {
    check(has(at(wide_rows, 0), col.name), "all_scalars row is missing #{col.name}");
    check(has(wide_null_row, col.name) && wide_null_row[col.name] == null, "#{col.name} should round trip NULL");
}
val EXPECTED_COLS = [col.name for (col in ALL_SCALAR_TYPES)];
check(keys(at(wide_rows, 0)) == EXPECTED_COLS, "keys(at(wide_rows, 0)) == EXPECTED_COLS");


# ---------------------------------------------------------------------------
# 4. duckdb nested datatypes
# ---------------------------------------------------------------------------

val NESTED_CASES = [
    {label: 'INTEGER[]',      sql_type: 'INTEGER[]',                         literal: '[1, 2, 3]',              want_type: Type.LIST, want: [1, 2, 3]},
    {label: 'VARCHAR[]',      sql_type: 'VARCHAR[]',                         literal: "['a', 'b']",             want_type: Type.LIST, want: ['a', 'b']},
    {label: 'LIST of LIST',   sql_type: 'INTEGER[][]',                       literal: '[[1], [2,3]]',           want_type: Type.LIST, want: [[1], [2, 3]]},
    {label: 'LIST of STRUCT', sql_type: 'STRUCT(a INTEGER, b VARCHAR)[]',    literal: "[{'a': 1, 'b': 'x'}]",    want_type: Type.LIST, want: [{a: 1, b: 'x'}]},
    {label: 'INTEGER[3]',     sql_type: 'INTEGER[3]',                        literal: '[1, 2, 3]',              want_type: Type.LIST, want: [1, 2, 3]},
    {label: 'STRUCT',         sql_type: 'STRUCT(a INTEGER, b VARCHAR)',      literal: "{'a': 1, 'b': 'x'}",     want_type: Type.MAP,  want: {a: 1, b: 'x'}},
    {label: 'nested STRUCT',  sql_type: 'STRUCT(a INTEGER, b STRUCT(c VARCHAR))', literal: "{'a': 1, 'b': {'c': 'x'}}", want_type: Type.MAP, want: {a: 1, b: {c: 'x'}}},
    {label: 'MAP',            sql_type: 'MAP(VARCHAR, INTEGER)',             literal: "map(['a','b'], [1,2])",  want_type: Type.MAP,  want: {a: 1, b: 2}},
    {label: 'MAP of LIST',    sql_type: 'MAP(VARCHAR, INTEGER[])',           literal: "map(['a','b'], [[1],[2]])", want_type: Type.MAP, want: {a: [1], b: [2]}},
    {label: 'UNION',          sql_type: 'UNION(i INTEGER, s VARCHAR)',       literal: "union_value(s := 'x')",  want_type: Type.STRING, want: 'x'},
    {label: 'VARIANT',        sql_type: 'VARIANT',                           literal: '1',                      want_type: Type.INT,  want: 1},
];

check_types(conn, NESTED_CASES);

# nested values have to survive a real table as well
exec(conn, """
    drop table if exists nested_t;
    create table nested_t(
        id integer,
        int_list integer[],
        str_list varchar[],
        fixed_arr integer[3],
        person struct(first varchar, age integer),
        lookup map(varchar, integer),
        deep struct(tags varchar[], score struct(a integer, b integer))
    );
    insert into nested_t
    values (1, [1,2,3], ['a','b'], [1,2,3], {'first': 'marc', 'age': 99}, map(['x','y'], [1,2]),
            {'tags': ['q','r'], 'score': {'a': 1, 'b': 2}});
""");

val nested_row = q1(conn, "select * from nested_t");
println("nested_row => #{nested_row}");
check(nested_row.int_list == [1, 2, 3], "nested_row.int_list == [1, 2, 3]");
check(nested_row.str_list == ['a', 'b'], "nested_row.str_list == ['a', 'b']");
check(nested_row.fixed_arr == [1, 2, 3], "nested_row.fixed_arr == [1, 2, 3]");
check(nested_row.person == {first: 'marc', age: 99}, "nested_row.person == {first: 'marc', age: 99}");
check(nested_row.lookup == {x: 1, y: 2}, "nested_row.lookup == {x: 1, y: 2}");
check(nested_row.deep == {tags: ['q', 'r'], score: {a: 1, b: 2}}, "nested_row.deep == {tags: ['q', 'r'], score: {a: 1, b: 2}}");

# the list, map, and struct sql functions have to agree with the blue side
check(q(conn, "select unnest([1,2,3])") == [[1], [2], [3]], "q(conn, \"select unnest([1,2,3])\") == [[1], [2], [3]]");
check(q_one(conn, "select list_value(1,2,3)") == [1, 2, 3], "q_one(conn, \"select list_value(1,2,3)\") == [1, 2, 3]");
check(q_one(conn, "select list_append([1], 2)") == [1, 2], "q_one(conn, \"select list_append([1], 2)\") == [1, 2]");
check(q_one(conn, "select list_has([1,2], 2)") == true, "q_one(conn, \"select list_has([1,2], 2)\") == true");
check(q_one(conn, "select array_length([1,2,3])") == 3, "q_one(conn, \"select array_length([1,2,3])\") == 3");
check(q_one(conn, "select list_concat([1],[2])") == [1, 2], "q_one(conn, \"select list_concat([1],[2])\") == [1, 2]");
check(q_one(conn, "select list_slice([1,2,3], 1, 2)") == [1, 2], "q_one(conn, \"select list_slice([1,2,3], 1, 2)\") == [1, 2]");
check(q_one(conn, "select struct_pack(a := 1, b := 'x')") == {a: 1, b: 'x'}, "q_one(conn, \"select struct_pack(a := 1, b := 'x')\") == {a: 1, b: 'x'}");
check(q_one(conn, "select struct_insert({'a': 1}, b := 2)") != null, "q_one(conn, \"select struct_insert({'a': 1}, b := 2)\") != null");
check(q_one(conn, "select map_keys(map(['a','b'],[1,2]))") == ['a', 'b'], "q_one(conn, \"select map_keys(map(['a','b'],[1,2]))\") == ['a', 'b']");
check(q_one(conn, "select map_values(map(['a','b'],[1,2]))") == [1, 2], "q_one(conn, \"select map_values(map(['a','b'],[1,2]))\") == [1, 2]");
check(q_one(conn, "select map_entries(map(['a','b'],[1,2]))") != null, "q_one(conn, \"select map_entries(map(['a','b'],[1,2]))\") != null");
check(q_one(conn, "select map_from_entries([{'k':'a','v':1}])") != null, "q_one(conn, \"select map_from_entries([{'k':'a','v':1}])\") != null");
check(q_one(conn, "select map_extract(map(['a'],[1]), 'a')") == 1, "q_one(conn, \"select map_extract(map(['a'],[1]), 'a')\") == 1");
check(q_one(conn, "select list_transform([1,2,3], x -> x * 2)") == [2, 4, 6], "q_one(conn, \"select list_transform([1,2,3], x -> x * 2)\") == [2, 4, 6]");
check(q_one(conn, "select list_filter([1,2,3,4], x -> x % 2 = 0)") == [2, 4], "q_one(conn, \"select list_filter([1,2,3,4], x -> x % 2 = 0)\") == [2, 4]");
check(q_one(conn, "select list_reduce([1,2,3], (a, b) -> a + b)") == 6, "q_one(conn, \"select list_reduce([1,2,3], (a, b) -> a + b)\") == 6");
check(q_one(conn, "select list_zip([1,2], ['a','b'])") != null, "q_one(conn, \"select list_zip([1,2], ['a','b'])\") != null");
check(q_one(conn, "select list_aggregate([1,2,3], 'sum')") == 6, "q_one(conn, \"select list_aggregate([1,2,3], 'sum')\") == 6");
check(q_one(conn, "select unnest(map_keys(map(['a'],['x'])))") == ['a'], "q_one(conn, \"select unnest(map_keys(map(['a'],['x'])))\") == ['a']");


# ---------------------------------------------------------------------------
# 5. duckdb temporal datatypes
#
# Everything here comes back as a carbon formatted string today, see blue-TODO
# about making time.Time a first class blue value. Until then the exact string
# form is what gets pinned down.
# ---------------------------------------------------------------------------

val TEMPORAL_CASES = [
    {label: 'DATE',         sql_type: 'DATE',        literal: "'2024-02-29'",                              want_type: Type.STRING, want: '2024-02-29'},
    {label: 'TIME',         sql_type: 'TIME',        literal: "'12:34:56'",                                want_type: Type.STRING, want: '12:34:56'},
    {label: 'TIME_NS',      sql_type: 'TIME_NS',     literal: "'12:34:56.123456789'",                      want_type: Type.STRING, want: '12:34:56.123456789'},
    {label: 'TIMETZ',       sql_type: 'TIMETZ',      literal: "'12:34:56-08'",                             want_type: Type.STRING, want: '12:34:56-08'},
    {label: 'TIMESTAMP',    sql_type: 'TIMESTAMP',   literal: "'2024-02-29 12:34:56.123456'",              want_type: Type.STRING, want: '2024-02-29 12:34:56.123456'},
    {label: 'DATETIME',     sql_type: 'DATETIME',    literal: "'2024-02-29 12:34:56'",                     want_type: Type.STRING, want: '2024-02-29 12:34:56'},
    {label: 'TIMESTAMPTZ',  sql_type: 'TIMESTAMPTZ', literal: "'2024-02-29 12:34:56+02'",                  want_type: Type.STRING, want: '2024-02-29 12:34:56+02'},
    {label: 'TIMESTAMP_S',  sql_type: 'TIMESTAMP_S', literal: "'2024-02-29 12:34:56'",                     want_type: Type.STRING, want: '2024-02-29 12:34:56'},
    {label: 'TIMESTAMP_MS', sql_type: 'TIMESTAMP_MS', literal: "'2024-02-29 12:34:56.123'",                want_type: Type.STRING, want: '2024-02-29 12:34:56.123'},
    {label: 'TIMESTAMP_NS', sql_type: 'TIMESTAMP_NS', literal: "'2024-02-29 12:34:56.123456789'",          want_type: Type.STRING, want: '2024-02-29 12:34:56.123456789'},
    {label: 'INTERVAL',     sql_type: 'INTERVAL',    literal: "'1 year 2 months 3 days 4 hours 5 minutes 6 seconds'", want_type: Type.STRING, want: '1 year 2 mons 3 days 04:05:06'},
];

check_types(conn, TEMPORAL_CASES);

# duckdb date and timestamp functions should agree with the blue time module
check(q_one(conn, "select typeof('2024-02-29'::date)") == 'DATE', "q_one(conn, \"select typeof('2024-02-29'::date)\") == 'DATE'");
check(q_one(conn, "select typeof('2024-02-29'::timestamp)") == 'TIMESTAMP', "q_one(conn, \"select typeof('2024-02-29'::timestamp)\") == 'TIMESTAMP'");
check(q_one(conn, "select extract(year from date '2024-02-29')") == 2024, "q_one(conn, \"select extract(year from date '2024-02-29')\") == 2024");
check(q_one(conn, "select date_add(date '2024-01-01', interval 1 day)") == '2024-01-02', "q_one(conn, \"select date_add(date '2024-01-01', interval 1 day)\") == '2024-01-02'");
check(q_one(conn, "select datediff('day', date '2024-01-01', date '2024-01-05')") == 4, "q_one(conn, \"select datediff('day', date '2024-01-01', date '2024-01-05')\") == 4");
check(q_one(conn, "select strftime(timestamp '2024-02-29 12:34:56', '%Y-%m-%dT%H:%M:%S')") == '2024-02-29T12:34:56', "q_one(conn, \"select strftime(timestamp '2024-02-29 12:34:56', '%Y-%m-%dT%H:%M:%S')\") == '2024-02-29T12:34:56'");
check(q_one(conn, "select epoch(date '1970-01-02')") == 86400, "q_one(conn, \"select epoch(date '1970-01-02')\") == 86400");
check(q_one(conn, "select make_date(2024, 2, 29)") == '2024-02-29', "q_one(conn, \"select make_date(2024, 2, 29)\") == '2024-02-29'");
check(q_one(conn, "select make_timestamp(2024, 2, 29, 12, 34, 56)") == '2024-02-29 12:34:56', "q_one(conn, \"select make_timestamp(2024, 2, 29, 12, 34, 56)\") == '2024-02-29 12:34:56'");
check(q_one(conn, "select current_date") != null, "q_one(conn, \"select current_date\") != null");
check(q_one(conn, "select current_timestamp") != null, "q_one(conn, \"select current_timestamp\") != null");
check(q_one(conn, "select now()") != null, "q_one(conn, \"select now()\") != null");
check(q_one(conn, "select today()") != null, "q_one(conn, \"select today()\") != null");
check(q_one(conn, "select date_part('month', date '2024-02-29')") == 2, "q_one(conn, \"select date_part('month', date '2024-02-29')\") == 2");


# ---------------------------------------------------------------------------
# 6. binding every blue type as a parameter
# ---------------------------------------------------------------------------

exec(conn, """
    drop table if exists bind_t;
    create table bind_t(
        i INTEGER, ui UINTEGER, bi BIGINT, f FLOAT, d DOUBLE, dec_v DECIMAL(18,3),
        s VARCHAR, b BOOLEAN, bl BLOB, dt DATE, ts TIMESTAMP, n INTEGER
    );
""");

val bind_ok = exec(conn, """
    insert into bind_t values (?, ?, ?, ?, ?, ?, ?, ?, ?, cast(? as date), cast(? as timestamp), ?)
""", [42, 7, 99, 1.5, 3.5, 1.234, 'str', true, to_bytes('abc'), '2024-02-29', '2024-02-29 12:34:56', null]);
check(bind_ok.rows_affected == 1, "bind_ok.rows_affected == 1");

val bound = q1(conn, "select * from bind_t");
check(bound.i == 42, "bound.i == 42");
check(bound.s == 'str', "bound.s == 'str'");
check(bound.b == true, "bound.b == true");
check(bound.bl == to_bytes('abc'), "bound.bl == to_bytes('abc')");
check(bound.n == null, "bound.n == null");

# blue types the driver still has to learn to translate on the way in
check(exec_error(conn, "insert into bind_t(i) values (?)", [170141183460469231731687303715884105727]) == null, "exec_error(conn, \"insert into bind_t(i) values (?)\", [170141183460469231731687303715884105727]) == null");
check(exec_error(conn, "insert into bind_t(ui) values (?)", [uint(5)]) == null, "exec_error(conn, \"insert into bind_t(ui) values (?)\", [uint(5)]) == null");
check(exec_error(conn, "insert into bind_t(s) values (?)", ['ünïcodé']) == null, "exec_error(conn, \"insert into bind_t(s) values (?)\", ['ünïcodé']) == null");

# types with no sql equivalent should error clearly rather than panic
check(exec_error(conn, "insert into bind_t(i) values (?)", [[1, 2]]) != null, "exec_error(conn, \"insert into bind_t(i) values (?)\", [[1, 2]]) != null");
check(exec_error(conn, "insert into bind_t(i) values (?)", [{a: 1}]) != null, "exec_error(conn, \"insert into bind_t(i) values (?)\", [{a: 1}]) != null");
check(exec_error(conn, "insert into bind_t(i) values (?)", [fun() { 1 }]) != null, "exec_error(conn, \"insert into bind_t(i) values (?)\", [fun() { 1 }]) != null");
check(query_error(conn, "select ?", [[1, 2]]) != null, "query_error(conn, \"select ?\", [[1, 2]]) != null");

# placeholder count and arg count have to line up
check(query_error(conn, "select ?, ?", [1]) != null, "query_error(conn, \"select ?, ?\", [1]) != null");


# ---------------------------------------------------------------------------
# 7. csv
# ---------------------------------------------------------------------------

f.csv_in = fp("in.csv");
_write(f.csv_in, """id,name,salary,department
1,john,2000,sales
2,Andrew,5000,finance
3,Mark,8000,hr
""");

# duckdb reads the file straight off disk with full type inference
check(cnt(conn, "select count(*) from read_csv_auto('#{f.csv_in}')") == 3, "cnt(conn, \"select count(*) from read_csv_auto('${f.csv_in}')\") == 3");
check(q1(conn, "select * from read_csv_auto('#{f.csv_in}') limit 1") == {id: 1, name: 'john', salary: 2000, department: 'sales'}, "q1(conn, \"select * from read_csv_auto('${f.csv_in}') limit 1\") == {id: 1, name: 'john', salary: 2000, department: 'sales'}");
check(q_one(conn, "select sum(salary) from read_csv_auto('#{f.csv_in}')") == 15000, "q_one(conn, \"select sum(salary) from read_csv_auto('${f.csv_in}')\") == 15000");

# the same file through the shorthand table name
check(q_one(conn, "select sum(salary) from '#{f.csv_in}'") == 15000, "q_one(conn, \"select sum(salary) from '${f.csv_in}'\") == 15000");

# explicit schema, no header, custom delimiter, custom quote, null string
f.csv_tsv = fp("in.tsv");
_write(f.csv_tsv, "1\tx\tnull\n2\ty\t\n");
check(qmap(conn, "select * from read_csv('#{f.csv_tsv}', delim='\t', header=false, columns={'a':'INTEGER','b':'VARCHAR','c':'VARCHAR'}, nullstr='null')") == [{a: 1, b: 'x', c: null}, {a: 2, b: 'y', c: ''}], "qmap(conn, \"select * from read_csv('${f.csv_tsv}', delim='\\t', header=false, columns={'a':'INTEGER','b':'VARCHAR','c':'VARCHAR'}, nullstr='null')\") == [{a: 1, b: 'x', c: null}, {a: 2, b: 'y', c: ''}]");

# all_varchar stops duckdb from guessing, sample_size and auto_detect are knobs too
check(q1(conn, "select * from read_csv_auto('#{f.csv_in}', all_varchar=true) limit 1").id == '1', "q1(conn, \"select * from read_csv_auto('${f.csv_in}', all_varchar=true) limit 1\").id == '1'");
check(cnt(conn, "select count(*) from read_csv('#{f.csv_in}', header=true, auto_detect=false, sample_size=-1, columns={'id':'INTEGER','name':'VARCHAR','salary':'INTEGER','department':'VARCHAR'})") == 3, "cnt(conn, \"select count(*) from read_csv('${f.csv_in}', header=true, auto_detect=false, sample_size=-1, columns={'id':'INTEGER','name':'VARCHAR','salary':'INTEGER','department':'VARCHAR'})\") == 3");
check(len(qmap(conn, "select * from read_csv('#{f.csv_in}', header=true, columns={'id':'INTEGER','name':'VARCHAR','salary':'INTEGER','department':'VARCHAR'}, skip=1)")) == 2, "len(qmap(conn, \"select * from read_csv('${f.csv_in}', header=true, columns={'id':'INTEGER','name':'VARCHAR','salary':'INTEGER','department':'VARCHAR'}, skip=1)\")) == 2");

# quoted fields containing the delimiter, and embedded newlines
f.csv_quoted = fp("quoted.csv");
_write(f.csv_quoted, "a,b\n\"x,y\",2\n");
check(q1(conn, "select * from read_csv_auto('#{f.csv_quoted}')").a == 'x,y', "q1(conn, \"select * from read_csv_auto('${f.csv_quoted}')\").a == 'x,y'");

# globs over several files, union_by_name lines up mismatched headers
f.csv_dir = fp("csv_dir");
mkdir_all(f.csv_dir);
_write(path.join(f.csv_dir, "a.csv"), "id\n1\n");
_write(path.join(f.csv_dir, "b.csv"), "id\n2\n");
f.csv_glob = path.join(f.csv_dir, '*.csv');
check(cnt(conn, "select count(*) from read_csv_auto('#{f.csv_glob}')") == 2, "cnt(conn, \"select count(*) from read_csv_auto('${f.csv_glob}')\") == 2");
check(cnt(conn, "select count(*) from read_csv_auto('#{f.csv_glob}', union_by_name=true)") == 2, "cnt(conn, \"select count(*) from read_csv_auto('${f.csv_glob}', union_by_name=true)\") == 2");

# COPY TO csv, then read it back
f.csv_out = fp("out.csv");
exec(conn, "copy (select * from users order by id) to '#{f.csv_out}' (format csv, header true)");
check(exists(f.csv_out), "exists(f.csv_out)");
check(cnt(conn, "select count(*) from read_csv_auto('#{f.csv_out}')") == 2, "cnt(conn, \"select count(*) from read_csv_auto('${f.csv_out}')\") == 2");
check(qmap(conn, "select name from read_csv_auto('#{f.csv_out}') order by id") == [{name: 'marc'}, {name: 'macgyver'}], "qmap(conn, \"select name from read_csv_auto('${f.csv_out}') order by id\") == [{name: 'marc'}, {name: 'macgyver'}]");

# custom delimiter, custom quote, no header
f.csv_out2 = fp("out2.csv");
exec(conn, "copy (select id, name from users) to '#{f.csv_out2}' (format csv, header false, delimiter '|', quote '\"')");
check(cnt(conn, "select count(*) from read_csv('#{f.csv_out2}', header=false, delim='|', columns={'id':'INTEGER','name':'VARCHAR'})") == 2, "cnt(conn, \"select count(*) from read_csv('${f.csv_out2}', header=false, delim='|', columns={'id':'INTEGER','name':'VARCHAR'})\") == 2");

# gzip compressed csv in both directions
f.csv_gz = fp("out.csv.gz");
exec(conn, "copy (select id, name from users) to '#{f.csv_gz}' (format csv, header true, compression gzip)");
check(cnt(conn, "select count(*) from read_csv_auto('#{f.csv_gz}')") == 2, "cnt(conn, \"select count(*) from read_csv_auto('${f.csv_gz}')\") == 2");

# one file per group with partition_by
f.csv_per_group = fp("per_group");
exec(conn, "copy (select department, name from read_csv_auto('#{f.csv_in}')) to '#{f.csv_per_group}' (format csv, header true, partition_by (department))");
check(len(walk(f.csv_per_group)) > 1, "len(walk(f.csv_per_group)) > 1");

# interop with the blue csv module, parse here then hand the rows to duckdb
val csv_module_rows = csv.parse(f.csv_in.read());
check(len(csv_module_rows) == 4, "len(csv_module_rows) == 4");
exec(conn, "drop table if exists from_csv_module");
exec(conn, "create table from_csv_module(id integer, name varchar, salary integer, department varchar)");
for (row in csv_module_rows[1..<len(csv_module_rows)]) {
    exec(conn, "insert into from_csv_module values (?, ?, ?, ?)", [int(row[0]), row[1], int(row[2]), row[3]]);
}
check(cnt(conn, "select count(*) from from_csv_module") == 3, "cnt(conn, \"select count(*) from from_csv_module\") == 3");

# named_fields gives maps, which is closer to what named_cols returns
val csv_named_rows = csv.parse(f.csv_in.read(), named_fields=true);
check(at(csv_named_rows, 0) == {id: '1', name: 'john', salary: '2000', department: 'sales'}, "at(csv_named_rows, 0) == {id: '1', name: 'john', salary: '2000', department: 'sales'}");

# and the other direction, duckdb query results dumped with the csv module
val dumped = csv.dump(qmap(conn, "select id, name from users order by id"));
check(dumped.contains("marc"), "dumped.contains(\"marc\")");
f.dumped_csv_path = fp('from_dump.csv');
_write(f.dumped_csv_path, dumped);
check(cnt(conn, "select count(*) from read_csv_auto('#{f.dumped_csv_path}')") == 2, "cnt(conn, \"select count(*) from read_csv_auto('${f.dumped_csv_path}')\") == 2");


# ---------------------------------------------------------------------------
# 8. parquet
# ---------------------------------------------------------------------------

f.pq_out = fp("out.parquet");
exec(conn, "copy (select * from users order by id) to '#{f.pq_out}' (format parquet)");
check(exists(f.pq_out), "exists(f.pq_out)");
check(cnt(conn, "select count(*) from read_parquet('#{f.pq_out}')") == 2, "cnt(conn, \"select count(*) from read_parquet('${f.pq_out}')\") == 2");
check(qmap(conn, "select name from read_parquet('#{f.pq_out}') order by id") == [{name: 'marc'}, {name: 'macgyver'}], "qmap(conn, \"select name from read_parquet('${f.pq_out}') order by id\") == [{name: 'marc'}, {name: 'macgyver'}]");

# the shorthand table name works for parquet too
check(cnt(conn, "select count(*) from '#{f.pq_out}'") == 2, "cnt(conn, \"select count(*) from '${f.pq_out}'\") == 2");

# compression and row group sizing
f.pq_zstd = fp("zstd.parquet");
exec(conn, "copy (select range i from range(1000)) to '#{f.pq_zstd}' (format parquet, compression zstd, row_group_size 128)");
check(cnt(conn, "select count(*) from read_parquet('#{f.pq_zstd}')") == 1000, "cnt(conn, \"select count(*) from read_parquet('${f.pq_zstd}')\") == 1000");
for (codec in ['snappy', 'gzip', 'zstd']) {
    val coded = fp("#{codec}.parquet");
    exec(conn, "copy (select 1 a) to '#{coded}' (format parquet, compression #{codec})");
    check(q_one(conn, "select a from read_parquet('#{coded}')") == 1, "q_one(conn, \"select a from read_parquet('${coded}')\") == 1");
}

# parquet v1 vs v2 data page formats
f.pq_v2 = fp("v2.parquet");
exec(conn, "copy (select 1 a, 'x' b) to '#{f.pq_v2}' (format parquet, parquet_version v2)");
check(q_one(conn, "select a from read_parquet('#{f.pq_v2}')") == 1, "q_one(conn, \"select a from read_parquet('${f.pq_v2}')\") == 1");

# metadata table functions
check(q_one(conn, "select num_rows from parquet_file_metadata('#{f.pq_out}')") == 2, "q_one(conn, \"select num_rows from parquet_file_metadata('${f.pq_out}')\") == 2");
check(cnt(conn, "select count(*) from parquet_metadata('#{f.pq_out}')") > 0, "cnt(conn, \"select count(*) from parquet_metadata('${f.pq_out}')\") > 0");
check(len(qmap(conn, "select name, type from parquet_schema('#{f.pq_out}')")) > 0, "len(qmap(conn, \"select name, type from parquet_schema('${f.pq_out}')\")) > 0");

# every scalar type has to survive a parquet round trip
f.pq_types = fp("types.parquet");
exec(conn, "copy (select * from all_scalars limit 1) to '#{f.pq_types}' (format parquet)");
val pq_round_trip = q1(conn, "select * from read_parquet('#{f.pq_types}')");
val original = q1(conn, "select * from all_scalars limit 1");
for (col in ALL_SCALAR_TYPES) {
    check(pq_round_trip[col.name] == original[col.name], "parquet round trip changed #{col.name}: #{original[col.name]} => #{pq_round_trip[col.name]}");
}

# nested types through parquet as well
f.pq_nested = fp("nested.parquet");
exec(conn, "copy (select * from nested_t) to '#{f.pq_nested}' (format parquet)");
check(q1(conn, "select int_list, person, lookup from read_parquet('#{f.pq_nested}')").person == {first: 'marc', age: 99}, "q1(conn, \"select int_list, person, lookup from read_parquet('${f.pq_nested}')\").person == {first: 'marc', age: 99}");

# partitioned writes with hive partitioning, then read the tree back
f.pq_partitioned = fp("pq_part");
exec(conn, "copy (select id, name, awesome from users) to '#{f.pq_partitioned}' (format parquet, partition_by (awesome), write_partition_columns false)");
check(len(walk(f.pq_partitioned)) >= 1, "len(walk(f.pq_partitioned)) >= 1");
f.pq_glob = path.join(f.pq_partitioned, "**/*.parquet");
check(cnt(conn, "select count(*) from read_parquet('#{f.pq_glob}', hive_partitioning=true)") == 2, "cnt(conn, \"select count(*) from read_parquet('${f.pq_glob}', hive_partitioning=true)\") == 2");
check(q1(conn, "select awesome, count(*) c from read_parquet('#{f.pq_glob}', hive_partitioning=true) group by awesome").c == 2, "q1(conn, \"select awesome, count(*) c from read_parquet('${f.pq_glob}', hive_partitioning=true) group by awesome\").c == 2");

# several files at once read back through a glob
f.pq_multi = fp("pq_multi");
mkdir_all(f.pq_multi);
f.pq_multi_file = path.join(f.pq_multi, 'part.parquet');
f.pq_multi_glob = path.join(f.pq_multi, '*.parquet');
exec(conn, "copy (select id, name from users) to '#{f.pq_multi_file}' (format parquet)");
check(cnt(conn, "select count(*) from read_parquet('#{f.pq_multi_glob}')") == 2, "cnt(conn, \"select count(*) from read_parquet('${f.pq_multi_glob}')\") == 2");

# build a table straight out of a parquet file
exec(conn, "drop table if exists from_parquet");
exec(conn, "create table from_parquet as select * from read_parquet('#{f.pq_out}')");
check(cnt(conn, "select count(*) from from_parquet") == 2, "cnt(conn, \"select count(*) from from_parquet\") == 2");

# filter pushdown and column pruning must not change results
check(cnt(conn, "select count(*) from read_parquet('#{f.pq_out}') where id > 1") == 1, "cnt(conn, \"select count(*) from read_parquet('${f.pq_out}') where id > 1\") == 1");
check(qmap(conn, "select id from read_parquet('#{f.pq_out}') order by id") == [{id: 1}, {id: 2}], "qmap(conn, \"select id from read_parquet('${f.pq_out}') order by id\") == [{id: 1}, {id: 2}]");


# ---------------------------------------------------------------------------
# 9. json
# ---------------------------------------------------------------------------

# json ships with the duckdb build but it is worth exercising explicitly
check(cnt(conn, "select count(*) from duckdb_extensions() where extension_name = 'json' and loaded") == 1, "cnt(conn, \"select count(*) from duckdb_extensions() where extension_name = 'json' and loaded\") == 1");
exec(conn, "load json");

f.json_out = fp("out.json");
exec(conn, "copy (select id, name from users order by id) to '#{f.json_out}' (format json)");
check(exists(f.json_out), "exists(f.json_out)");
check(cnt(conn, "select count(*) from read_json_auto('#{f.json_out}')") == 2, "cnt(conn, \"select count(*) from read_json_auto('${f.json_out}')\") == 2");
check(qmap(conn, "select name from read_json_auto('#{f.json_out}') order by id") == [{name: 'marc'}, {name: 'macgyver'}], "qmap(conn, \"select name from read_json_auto('${f.json_out}') order by id\") == [{name: 'marc'}, {name: 'macgyver'}]");

# newline delimited json, one object per line
f.ndjson_out = fp("out.ndjson");
exec(conn, "copy (select id, name from users order by id) to '#{f.ndjson_out}' (format json, array false)");
check(cnt(conn, "select count(*) from read_ndjson_auto('#{f.ndjson_out}')") == 2, "cnt(conn, \"select count(*) from read_ndjson_auto('${f.ndjson_out}')\") == 2");

# a single json array instead of ndjson
f.json_arr = fp("out_array.json");
exec(conn, "copy (select id, name from users order by id) to '#{f.json_arr}' (format json, array true)");
check(cnt(conn, "select count(*) from read_json_auto('#{f.json_arr}')") == 2, "cnt(conn, \"select count(*) from read_json_auto('${f.json_arr}')\") == 2");

# explicit schema instead of auto detection
check(qmap(conn, "select * from read_json('#{f.json_out}', columns={'id':'INTEGER','name':'VARCHAR'})") == [{id: 1, name: 'marc'}, {id: 2, name: 'macgyver'}], "qmap(conn, \"select * from read_json('${f.json_out}', columns={'id':'INTEGER','name':'VARCHAR'})\") == [{id: 1, name: 'marc'}, {id: 2, name: 'macgyver'}]");

# json written by blue itself has to be readable
f.blue_json = fp("blue.json");
_write(f.blue_json, [{a: 1, b: 'x'}, {a: 2, b: 'y'}].to_json());
check(qmap(conn, "select * from read_json_auto('#{f.blue_json}')") == [{a: 1, b: 'x'}, {a: 2, b: 'y'}], "qmap(conn, \"select * from read_json_auto('${f.blue_json}')\") == [{a: 1, b: 'x'}, {a: 2, b: 'y'}]");

# json sql functions
check(q_one(conn, "select json_valid('{\"a\":1}')") == true, "q_one(conn, \"select json_valid('{\\\"a\\\":1}')\") == true");
check(q_one(conn, "select json_valid('nope')") == false, "q_one(conn, \"select json_valid('nope')\") == false");
check(q_one(conn, "select json_extract('{\"a\":1}', '$.a')") == 1, "q_one(conn, \"select json_extract('{\\\"a\\\":1}', '$.a')\") == 1");
check(q_one(conn, "select json_extract_string('{\"a\":\"x\"}', '$.a')") == 'x', "q_one(conn, \"select json_extract_string('{\\\"a\\\":\\\"x\\\"}', '$.a')\") == 'x'");
check(q_one(conn, "select json_extract('{\"a\":{\"b\":2}}', '$.a.b')") == 2, "q_one(conn, \"select json_extract('{\\\"a\\\":{\\\"b\\\":2}}', '$.a.b')\") == 2");
check(q_one(conn, "select json_keys('{\"a\":1,\"b\":2}')") == ['a', 'b'], "q_one(conn, \"select json_keys('{\\\"a\\\":1,\\\"b\\\":2}')\") == ['a', 'b']");
check(q_one(conn, "select json_array_length('[1,2,3]')") == 3, "q_one(conn, \"select json_array_length('[1,2,3]')\") == 3");
check(q_one(conn, "select json_array(1,2,3)") == '[1,2,3]', "q_one(conn, \"select json_array(1,2,3)\") == '[1,2,3]'");
check(q_one(conn, "select json_object('a', 1)") == '{"a":1}', "q_one(conn, \"select json_object('a', 1)\") == '{\"a\":1}'");
check(q_one(conn, "select to_json({'a': 1})") == '{"a":1}', "q_one(conn, \"select to_json({'a': 1})\") == '{\"a\":1}'");
check(q_one(conn, "select json_group_array(id) from users") != null, "q_one(conn, \"select json_group_array(id) from users\") != null");
check(q_one(conn, "select json_group_object(name, id) from users") != null, "q_one(conn, \"select json_group_object(name, id) from users\") != null");
check(q_one(conn, "select array_to_json([1,2,3])") != null, "q_one(conn, \"select array_to_json([1,2,3])\") != null");
check(q_one(conn, "select from_json('{\"a\":1}', '{\"a\":\"INTEGER\"}')") != null, "q_one(conn, \"select from_json('{\\\"a\\\":1}', '{\\\"a\\\":\\\"INTEGER\\\"}')\") != null");

# the json column on users has to behave like every other type on the way out
check(q1(conn, "select meta from users where id = 1").meta == '{"k":1}', "q1(conn, \"select meta from users where id = 1\").meta == '{\"k\":1}'");
check(q_one(conn, "select json_extract(meta, '$.k') from users where id = 1") == 1, "q_one(conn, \"select json_extract(meta, '$.k') from users where id = 1\") == 1");
check(cnt(conn, "select count(*) from read_json_auto('#{f.json_out}') j join users u on u.id = j.id") == 2, "cnt(conn, \"select count(*) from read_json_auto('${f.json_out}') j join users u on u.id = j.id\") == 2");


# ---------------------------------------------------------------------------
# 10. extensions and remote sources
#
# Installing an extension needs network access, so each one is wrapped and
# reported instead of failing the whole file on an offline machine.
# ---------------------------------------------------------------------------

fun try_extension(name) {
    try {
        conn.execute("install #{name}");
        conn.execute("load #{name}");
        println("extension #{name} loaded => #{q_one(conn, 'select 1')}");
        return true
    } catch (e) {
        println("extension #{name} skipped: #{e}");
        return false
    }
}

val HTTPFS_OK = try_extension('httpfs');
try_extension('sqlite_scanner');
try_extension('postgres_scanner');
try_extension('mysql_scanner');
try_extension('excel');
try_extension('fts');
try_extension('icu');
val TPCH_OK = try_extension('tpch');

# bundled extensions never need an install
check(cnt(conn, "select count(*) from duckdb_extensions() where extension_name = 'core_functions' and loaded") == 1, "cnt(conn, \"select count(*) from duckdb_extensions() where extension_name = 'core_functions' and loaded\") == 1");
check(cnt(conn, "select count(*) from duckdb_extensions() where extension_name = 'parquet' and loaded") == 1, "cnt(conn, \"select count(*) from duckdb_extensions() where extension_name = 'parquet' and loaded\") == 1");
check(len(qmap(conn, "select extension_name from duckdb_extensions() where loaded")) > 0, "len(qmap(conn, \"select extension_name from duckdb_extensions() where loaded\")) > 0");
check(query_error(conn, "load definitely_not_an_extension") != null, "query_error(conn, \"load definitely_not_an_extension\") != null");

# tpch is bundled, generate a tiny dataset and run a real analytic query
if (TPCH_OK) {
    exec(conn, "call dbgen(sf := 0.01)");
    check(cnt(conn, "select count(*) from part") > 0, "cnt(conn, \"select count(*) from part\") > 0");
    check(cnt(conn, "select count(*) from lineitem") > 0, "cnt(conn, \"select count(*) from lineitem\") > 0");
    val top_ship = qmap(conn, """
        select l_returnflag, sum(l_extendedprice) as revenue
        from lineitem
        where l_shipdate < date '1995-01-01'
        group by l_returnflag
        order by revenue desc
        limit 1
    """);
    println("tpch top_ship => #{top_ship}");
    check(len(top_ship) == 1, "len(top_ship) == 1");
    check(has(at(top_ship, 0), 'l_returnflag'), "has(at(top_ship, 0), 'l_returnflag')");
    check(has(at(top_ship, 0), 'revenue'), "has(at(top_ship, 0), 'revenue')");
    exec(conn, "call dbgen(sf := 0)");
}

# a remote parquet over https, only exercised when httpfs is around
if (HTTPFS_OK) {
    try {
        val remote = "https://raw.githubusercontent.com/duckdb/duckdb/main/data/parquet-testing/userdata1.parquet";
        val remote_rows = cnt(conn, "select count(*) from read_parquet('#{remote}')");
        println("remote parquet rows => #{remote_rows}");
    } catch (e) {
        println("remote parquet skipped: #{e}");
    }
}


# ---------------------------------------------------------------------------
# 11. sql feature smoke tests
#
# Objects the smoke statements lean on, created up front.
# ---------------------------------------------------------------------------

exec(conn, """
    drop table if exists smoke_pk;
    drop table if exists smoke_ret;
    create table smoke_pk(id integer primary key, v varchar);
    insert into smoke_pk values (1, 'a');
    create table smoke_ret(id integer, v varchar);
    create sequence smoke_seq;
    create macro smoke_add(a, b) as a + b;
    create type smoke_mood as enum ('ok', 'bad');
    create type smoke_pair as struct(k varchar, v integer);
""");

val SQL_SMOKE_STATEMENTS = [
    {label: 'cte',                 sql: "with a as (select 1 x) select * from a"},
    {label: 'recursive cte',       sql: "with recursive r as (select 1 n union all select n+1 from r where n < 5) select * from r"},
    {label: 'materialized cte',    sql: "with a as materialized (select 1 x) select * from a"},
    {label: 'values clause',       sql: "select * from (values (1,'a'), (2,'b')) t(id, name)"},
    {label: 'window row_number',   sql: "select id, row_number() over (order by id) rn from users"},
    {label: 'window rank',         sql: "select id, rank() over (order by id) r, dense_rank() over (order by id) dr from users"},
    {label: 'window frame',        sql: "select id, sum(id) over (order by id rows between 1 preceding and current row) s from users"},
    {label: 'window lag lead',     sql: "select id, lag(id) over (order by id) l, lead(id) over (order by id) g from users"},
    {label: 'qualify',             sql: "select id from users qualify row_number() over (order by id) = 1"},
    {label: 'group by all',        sql: "select department, count(*) c from read_csv_auto('#{f.csv_in}') group by all order by department"},
    {label: 'group by rollup',     sql: "select department, count(*) from read_csv_auto('#{f.csv_in}') group by rollup(department)"},
    {label: 'having',              sql: "select department from read_csv_auto('#{f.csv_in}') group by department having count(*) >= 1"},
    {label: 'join',                sql: "select u.id, c.department from users u join read_csv_auto('#{f.csv_in}') c on c.id = u.id"},
    {label: 'left join',           sql: "select u.id, c.department from users u left join read_csv_auto('#{f.csv_in}') c on c.id = u.id"},
    {label: 'cross join',          sql: "select count(*) from users a cross join users b"},
    {label: 'natural join',        sql: "select count(*) from (select id from users) a natural join (select id from users) b"},
    {label: 'asof join',           sql: "select count(*) from (values (1, 10, 1)) l(k, v, t) asof join (values (1, 20, 1)) r(k, w, t) on (l.k = r.k and l.t <= r.t)"},
    {label: 'union',               sql: "select 1 union all select 2"},
    {label: 'except',              sql: "select 1 except select 2"},
    {label: 'intersect',           sql: "select 1 intersect select 1"},
    {label: 'pivot',               sql: "from (values (1,'a',2)) t(x,y,z) pivot (sum(z) for y in ('a'))"},
    {label: 'unpivot',             sql: "from (values (1,2,3)) t(x,y,z) unpivot (v for name in (y, z))"},
    {label: 'lateral',             sql: "select u.id, x.n from users u, lateral (select 1 n) x"},
    {label: 'tablesample',         sql: "select * from users tablesample reservoir (2)"},
    {label: 'order by all',        sql: "select * from users order by all"},
    {label: 'nulls last',          sql: "select * from users order by age nulls last"},
    {label: 'limit offset',        sql: "select * from range(10) limit 2 offset 3"},
    {label: 'fetch first',         sql: "select 1 fetch first row only"},
    {label: 'case when',           sql: "select case when 1 = 1 then 'y' else 'n' end"},
    {label: 'coalesce',            sql: "select coalesce(null, null, 3)"},
    {label: 'nullif',              sql: "select nullif(1, 1)"},
    {label: 'cast',                sql: "select cast('1' as integer)"},
    {label: 'try_cast',            sql: "select try_cast('nope' as integer)"},
    {label: 'aggregates',          sql: "select sum(id), avg(id), min(id), max(id), count(id), count(*) from users"},
    {label: 'stddev variance',     sql: "select stddev(id), variance(id), median(id) from users"},
    {label: 'string agg',          sql: "select string_agg(name, ',' order by name) from users"},
    {label: 'list agg',            sql: "select list(id order by id) from users"},
    {label: 'approx count',        sql: "select approx_count_distinct(id) from users"},
    {label: 'regexp',              sql: "select regexp_matches('abc', 'a.c'), regexp_extract('abc', 'a(b)', 1), regexp_replace('abc','b','x')"},
    {label: 'string funcs',        sql: "select upper('a'), lower('A'), length('ab'), trim(' a '), substr('abc', 1, 2), replace('abc','b','x'), lpad('1',3,'0')"},
    {label: 'math funcs',          sql: "select abs(-1), ceil(1.2), floor(1.8), round(1.55, 1), pow(2,3), sqrt(4), mod(5,2), ln(1), exp(1)"},
    {label: 'bit ops',             sql: "select 1 & 3, 1 | 2, xor(1, 3), 1 << 2, 1 >> 1, ~1"},
    {label: 'blob funcs',          sql: "select octet_length('abc'::blob), md5('abc'), sha256('abc')"},
    {label: 'typeof',              sql: "select typeof(1), typeof(1.0), typeof('a'), typeof([1]), typeof({'a':1})"},
    {label: 'if not exists',       sql: "create table if not exists smoke_pk(id integer)"},
    {label: 'select from range',   sql: "select count(*) from range(100)"},
    {label: 'summarize',           sql: "from users summary"},
    {label: 'describe table',      sql: "describe users"},
    {label: 'describe select',     sql: "describe select 1 a, 'x' b"},
    {label: 'show tables',         sql: "show tables"},
    {label: 'show databases',      sql: "show databases"},
    {label: 'show all tables',     sql: "show all tables"},
    {label: 'pragma database_list', sql: "pragma database_list"},
    {label: 'pragma table_info',   sql: "pragma table_info('users')"},
    {label: 'pragma collations',   sql: "pragma collations"},
    {label: 'call duckdb_tables',  sql: "call duckdb_tables()"},
    {label: 'call duckdb_columns', sql: "call duckdb_columns()"},
    {label: 'call duckdb_settings', sql: "call duckdb_settings()"},
    {label: 'pragma version',      sql: "pragma version"},
    {label: 'explain',             sql: "explain select * from users"},
    {label: 'explain analyze',     sql: "explain analyze select * from users"},
];

smoke_all(conn, SQL_SMOKE_STATEMENTS);

# statements that mutate, run outside the read only smoke list
check(exec(conn, "insert into smoke_ret values (100, 'z')").rows_affected == 1, "exec(conn, \"insert into smoke_ret values (100, 'z')\").rows_affected == 1");
check(qmap(conn, "insert into smoke_ret values (101, 'y') returning id") == [{id: 101}], "qmap(conn, \"insert into smoke_ret values (101, 'y') returning id\") == [{id: 101}]");
check(qmap(conn, "update smoke_ret set v = 'q' where id = 100 returning id, v") == [{id: 100, v: 'q'}], "qmap(conn, \"update smoke_ret set v = 'q' where id = 100 returning id, v\") == [{id: 100, v: 'q'}]");
check(qmap(conn, "delete from smoke_ret where id = 100 returning id") == [{id: 100}], "qmap(conn, \"delete from smoke_ret where id = 100 returning id\") == [{id: 100}]");
check(exec(conn, "insert into smoke_pk values (1, 'b') on conflict (id) do update set v = excluded.v").rows_affected == 1, "exec(conn, \"insert into smoke_pk values (1, 'b') on conflict (id) do update set v = excluded.v\").rows_affected == 1");
check(exec(conn, "insert or ignore into smoke_pk values (1, 'c')").rows_affected == 0, "exec(conn, \"insert or ignore into smoke_pk values (1, 'c')\").rows_affected == 0");
check(exec(conn, "insert or replace into smoke_pk values (1, 'd')").rows_affected == 1, "exec(conn, \"insert or replace into smoke_pk values (1, 'd')\").rows_affected == 1");
check(exec(conn, "insert into smoke_pk select id + 10, v from smoke_pk").rows_affected >= 1, "exec(conn, \"insert into smoke_pk select id + 10, v from smoke_pk\").rows_affected >= 1");
check(exec(conn, "update smoke_pk set v = 'u' where id > 10").rows_affected >= 1, "exec(conn, \"update smoke_pk set v = 'u' where id > 10\").rows_affected >= 1");
check(exec(conn, "delete from smoke_pk where id > 10").rows_affected >= 1, "exec(conn, \"delete from smoke_pk where id > 10\").rows_affected >= 1");

exec(conn, "prepare smoke_stmt as select ?, ?");
check(q(conn, "execute smoke_stmt(1, 'two')") == [[1, 'two']], "q(conn, \"execute smoke_stmt(1, 'two')\") == [[1, 'two']]");
check(exec(conn, "deallocate smoke_stmt") != null, "exec(conn, \"deallocate smoke_stmt\") != null");

check(q_one(conn, "select smoke_add(1, 2)") == 3, "q_one(conn, \"select smoke_add(1, 2)\") == 3");
check(q_one(conn, "select nextval('smoke_seq')") != null, "q_one(conn, \"select nextval('smoke_seq')\") != null");
check(q_one(conn, "select 'ok'::smoke_mood") == 'ok', "q_one(conn, \"select 'ok'::smoke_mood\") == 'ok'");
check(q_one(conn, "select typeof({'k':'a','v':1}::smoke_pair)") != null, "q_one(conn, \"select typeof({'k':'a','v':1}::smoke_pair)\") != null");

exec(conn, "create or replace view smoke_view as select id, name from users");
check(cnt(conn, "select count(*) from smoke_view") == 2, "cnt(conn, \"select count(*) from smoke_view\") == 2");
exec(conn, "create or replace table smoke_ctas as select * from users");
check(cnt(conn, "select count(*) from smoke_ctas") == 2, "cnt(conn, \"select count(*) from smoke_ctas\") == 2");
exec(conn, "create temp table if not exists smoke_tmp(a integer)");
exec(conn, "insert into smoke_tmp values (1)");
check(cnt(conn, "select count(*) from smoke_tmp") == 1, "cnt(conn, \"select count(*) from smoke_tmp\") == 1");
exec(conn, "create index if not exists smoke_idx on smoke_pk(v)");
check(cnt(conn, "select count(*) from duckdb_indexes()") >= 1, "cnt(conn, \"select count(*) from duckdb_indexes()\") >= 1");

# constraints all have to be understood and enforced
exec(conn, """
    drop table if exists constraint_t;
    create table parents(id integer primary key);
    create table constraint_t(
        id integer primary key,
        must_be varchar not null,
        only_pos integer check (only_pos > 0),
        defaulted integer default 7,
        parent integer references parents(id)
    );
""");
check(exec(conn, "insert into constraint_t(id, must_be, only_pos, parent) values (1, 'x', 1, null)").rows_affected == 1, "exec(conn, \"insert into constraint_t(id, must_be, only_pos, parent) values (1, 'x', 1, null)\").rows_affected == 1");
check(q1(conn, "select defaulted from constraint_t").defaulted == 7, "q1(conn, \"select defaulted from constraint_t\").defaulted == 7");
check(exec_error(conn, "insert into constraint_t(id, must_be, only_pos) values (2, null, 1)") != null, "exec_error(conn, \"insert into constraint_t(id, must_be, only_pos) values (2, null, 1)\") != null");
check(exec_error(conn, "insert into constraint_t(id, must_be, only_pos) values (3, 'y', -1)") != null, "exec_error(conn, \"insert into constraint_t(id, must_be, only_pos) values (3, 'y', -1)\") != null");
check(exec_error(conn, "insert into constraint_t(id, must_be, only_pos) values (1, 'y', 1)") != null, "exec_error(conn, \"insert into constraint_t(id, must_be, only_pos) values (1, 'y', 1)\") != null");
exec(conn, "insert into parents values (5)");
check(exec_error(conn, "insert into constraint_t(id, must_be, only_pos, parent) values (4, 'z', 1, 999)") != null, "exec_error(conn, \"insert into constraint_t(id, must_be, only_pos, parent) values (4, 'z', 1, 999)\") != null");


# ---------------------------------------------------------------------------
# 12. transactions
# ---------------------------------------------------------------------------

exec(conn, "drop table if exists txn_t");
exec(conn, "create table txn_t(a integer)");

exec(conn, "begin transaction");
exec(conn, "insert into txn_t values (1)");
check(cnt(conn, "select count(*) from txn_t") == 1, "cnt(conn, \"select count(*) from txn_t\") == 1");
exec(conn, "commit");
check(cnt(conn, "select count(*) from txn_t") == 1, "cnt(conn, \"select count(*) from txn_t\") == 1");

exec(conn, "begin transaction");
exec(conn, "insert into txn_t values (2)");
exec(conn, "rollback");
check(cnt(conn, "select count(*) from txn_t") == 1, "cnt(conn, \"select count(*) from txn_t\") == 1");

# begin and commit through a single exec of several statements
exec(conn, "begin transaction; insert into txn_t values (3); commit;");
check(cnt(conn, "select count(*) from txn_t") == 2, "cnt(conn, \"select count(*) from txn_t\") == 2");

# a constraint violation inside a transaction rolls the whole thing back
exec(conn, "begin transaction");
exec_error(conn, "insert into txn_t values (4)");
exec_error(conn, "insert into txn_t select * from nope_not_a_table");
exec(conn, "rollback");
check(cnt(conn, "select count(*) from txn_t") == 2, "cnt(conn, \"select count(*) from txn_t\") == 2");

# duckdb also knows start transaction, savepoint style nesting is not supported
# but the statements at least have to parse
check(query_error(conn, "start transaction") == null || exec_error(conn, "start transaction") == null, "query_error(conn, \"start transaction\") == null || exec_error(conn, \"start transaction\") == null");
exec(conn, "commit");


# ---------------------------------------------------------------------------
# 13. schemas, attach/detach, export/import
# ---------------------------------------------------------------------------

exec(conn, "create schema if not exists s1");
exec(conn, "create table if not exists s1.t(a integer)");
exec(conn, "insert into s1.t values (1)");
check(cnt(conn, "select count(*) from s1.t") == 1, "cnt(conn, \"select count(*) from s1.t\") == 1");
exec(conn, "set schema 's1'");
check(cnt(conn, "select count(*) from t") == 1, "cnt(conn, \"select count(*) from t\") == 1");
exec(conn, "reset schema");
check(cnt(conn, "select count(*) from s1.t") == 1, "cnt(conn, \"select count(*) from s1.t\") == 1");
check(len(qmap(conn, "select schema_name from information_schema.schemata")) >= 2, "len(qmap(conn, \"select schema_name from information_schema.schemata\")) >= 2");
exec(conn, "drop schema s1 cascade");
check(query_error(conn, "select * from s1.t") != null, "query_error(conn, \"select * from s1.t\") != null");

# a second attached duckdb file, queried from the same connection
f.attached_db = fp("attached.db");
exec(conn, "attach '#{f.attached_db}' as other");
exec(conn, "create table other.othertab(a integer)");
exec(conn, "insert into other.othertab values (1), (2)");
check(cnt(conn, "select count(*) from other.othertab") == 2, "cnt(conn, \"select count(*) from other.othertab\") == 2");
check(cnt(conn, "select count(*) from main.users") == 2, "cnt(conn, \"select count(*) from main.users\") == 2");
check(cnt(conn, "select count(*) from information_schema.tables where table_catalog = 'other'") >= 1, "cnt(conn, \"select count(*) from information_schema.tables where table_catalog = 'other'\") >= 1");
exec(conn, "detach other");
check(query_error(conn, "select * from other.othertab") != null, "query_error(conn, \"select * from other.othertab\") != null");
check(exists(f.attached_db), "exists(f.attached_db)");

# a fresh connection has to see what the attached database wrote
var reopen = open_duck(f.attached_db);
check(q_one(reopen, "select count(*) from othertab") == 2, "q_one(reopen, \"select count(*) from othertab\") == 2");
reopen.close();

# export the whole database and import it back
f.export_dir = fp("exported");
exec(conn, "export database '#{f.export_dir}'");
check(len(walk(f.export_dir)) > 0, "len(walk(f.export_dir)) > 0");
var imported = open_duck();
exec(imported, "import database '#{f.export_dir}'");
check(q_one(imported, "select count(*) from users") == 2, "q_one(imported, \"select count(*) from users\") == 2");
imported.close();


# ---------------------------------------------------------------------------
# 14. catalog and settings introspection
# ---------------------------------------------------------------------------

check(len(qmap(conn, "select table_name from information_schema.tables where table_name = 'users'")) == 1, "len(qmap(conn, \"select table_name from information_schema.tables where table_name = 'users'\")) == 1");
val user_cols = qmap(conn, "select column_name, data_type from information_schema.columns where table_name = 'users' order by ordinal_position");
check(len(user_cols) == 11, "len(user_cols) == 11");
check(at(user_cols, 0) == {column_name: 'id', data_type: 'INTEGER'}, "at(user_cols, 0) == {column_name: 'id', data_type: 'INTEGER'}");
check(user_cols[1].data_type == 'VARCHAR', "user_cols[1].data_type == 'VARCHAR'");
println("users columns => #{user_cols}");

val duck_tables = qmap(conn, "select table_name, column_count from duckdb_tables() where table_name = 'users'");
check(at(duck_tables, 0).column_count == 11, "at(duck_tables, 0).column_count == 11");
check(len(qmap(conn, "select * from duckdb_columns() where table_name = 'users'")) == 11, "len(qmap(conn, \"select * from duckdb_columns() where table_name = 'users'\")) == 11");
check(len(qmap(conn, "select function_name from duckdb_functions() where function_name = 'list_transform'")) >= 1, "len(qmap(conn, \"select function_name from duckdb_functions() where function_name = 'list_transform'\")) >= 1");
check(len(qmap(conn, "select * from duckdb_types()")) > 0, "len(qmap(conn, \"select * from duckdb_types()\")) > 0");
check(len(qmap(conn, "select * from duckdb_settings()")) > 0, "len(qmap(conn, \"select * from duckdb_settings()\")) > 0");
check(len(qmap(conn, "select * from duckdb_databases()")) >= 1, "len(qmap(conn, \"select * from duckdb_databases()\")) >= 1");
check(len(qmap(conn, "select * from duckdb_schemas()")) >= 1, "len(qmap(conn, \"select * from duckdb_schemas()\")) >= 1");
check(cnt(conn, "select count(*) from duckdb_views() where view_name = 'smoke_view'") == 1, "cnt(conn, \"select count(*) from duckdb_views() where view_name = 'smoke_view'\") == 1");
check(cnt(conn, "select count(*) from duckdb_constraints() where table_name = 'users'") >= 1, "cnt(conn, \"select count(*) from duckdb_constraints() where table_name = 'users'\") >= 1");
check(q_one(conn, "select database_name from duckdb_databases() limit 1") != null, "q_one(conn, \"select database_name from duckdb_databases() limit 1\") != null");

# settings round trip through set, current_setting, and reset
check(q_one(conn, "select current_setting('threads')") != null, "q_one(conn, \"select current_setting('threads')\") != null");
exec(conn, "set threads = 1");
check(q_one(conn, "select current_setting('threads')") == 1, "q_one(conn, \"select current_setting('threads')\") == 1");
exec(conn, "reset threads");
exec(conn, "set memory_limit = '256MB'");
check(q_one(conn, "select current_setting('memory_limit')") != null, "q_one(conn, \"select current_setting('memory_limit')\") != null");
exec(conn, "reset memory_limit");
exec(conn, "set enable_progress_bar = true");
exec(conn, "reset enable_progress_bar");
check(query_error(conn, "select current_setting('definitely_not_a_setting')") != null, "query_error(conn, \"select current_setting('definitely_not_a_setting')\") != null");
check(q_one(conn, "select version()") != null, "q_one(conn, \"select version()\") != null");


# ---------------------------------------------------------------------------
# 15. error handling
#
# Every one of these has to come back as a blue error value that try/catch can
# see, never a panic and never a silent empty result.
# ---------------------------------------------------------------------------

val ERROR_CASES = [
    {label: 'syntax error',      sql: "selct 1 from"},
    {label: 'unknown table',     sql: "select * from definitely_not_a_table"},
    {label: 'unknown column',    sql: "select not_a_column from users"},
    {label: 'unknown function',  sql: "select definitely_not_a_function(1)"},
    {label: 'bad cast',          sql: "select 'abc'::integer"},
    {label: 'int overflow',      sql: "select 9223372036854775807 + 1"},
    {label: 'wrong arg count',   sql: "select abs(1, 2)"},
    {label: 'ambiguous column',  sql: "select id from users, smoke_pk"},
    {label: 'unknown type',      sql: "select cast(1 as definitely_not_a_type)"},
    {label: 'pk violation',      sql: "select 1; insert into smoke_pk values (1, 'dup')"},
];

for (c in ERROR_CASES) {
    val msg = query_error(conn, c.sql);
    val exec_msg = exec_error(conn, c.sql);
    println("error #{c.label} => #{msg}");
    check(msg != null || exec_msg != null, "#{c.label} should have errored");
}

# errors coming out of the module itself, not duckdb
check(query_error(conn, "select 1", [[1, 2]]) != null, "query_error(conn, \"select 1\", [[1, 2]]) != null");
try {
    db.open(db_name="", db_type="duckdb");
    check(false, "an empty db_name should not open");
} catch (e) {
    println("empty db_name => #{e}");
}


# ---------------------------------------------------------------------------
# 16. interop with the rest of blue
# ---------------------------------------------------------------------------

# query results round trip through blue json
val as_json = qmap(conn, "select id, name from users order by id").to_json();
check(as_json == '[{"id":1,"name":"marc"},{"id":2,"name":"macgyver"}]', "as_json == '[{\"id\":1,\"name\":\"marc\"},{\"id\":2,\"name\":\"macgyver\"}]'");
check(as_json.from_json() == [{id: 1, name: 'marc'}, {id: 2, name: 'macgyver'}], "as_json.from_json() == [{id: 1, name: 'marc'}, {id: 2, name: 'macgyver'}]");

# the blue time module and duckdb timestamps agree on the same instant
val now_str = time.format(time.now());
check(q_one(conn, "select cast('#{now_str}' as timestamp) is not null") == true, "q_one(conn, \"select cast('${now_str}' as timestamp) is not null\") == true");
check(q_one(conn, "select strftime(cast('#{now_str}' as timestamp), '%Y-%m-%dT%H:%M:%SZ')") == now_str, "q_one(conn, \"select strftime(cast('${now_str}' as timestamp), '%Y-%m-%dT%H:%M:%SZ')\") == now_str");

# string built rows can be inserted through parameters
val built = "marc;99";
val parts = built.split(";");
exec(conn, "drop table if exists built_t");
exec(conn, "create table built_t(name varchar, age integer)");
exec(conn, "insert into built_t values (?, ?)", [parts[0], int(parts[1])]);
check(q1(conn, "select * from built_t") == {name: 'marc', age: 99}, "q1(conn, \"select * from built_t\") == {name: 'marc', age: 99}");

# bytes round trip through the db and back into blue
val raw = to_bytes('Hello World!');
exec(conn, "drop table if exists bytes_t");
exec(conn, "create table bytes_t(b blob)");
exec(conn, "insert into bytes_t values (?)", [raw]);
check(q_one(conn, "select b from bytes_t") == raw, "q_one(conn, \"select b from bytes_t\") == raw");

# sets and maps on the blue side line up with duckdb list and map types
val from_db_list = q_one(conn, "select list(id) from users");
check(type(from_db_list) == Type.LIST, "type(from_db_list) == Type.LIST");
check(from_db_list.to_json() != null, "from_db_list.to_json() != null");
val from_db_map = q_one(conn, "select map_from_entries([{'k':'a','v':1}])");
check(type(from_db_map) == Type.MAP, "type(from_db_map) == Type.MAP");

# a larger scan, this is the path a real analytics query takes
exec(conn, "drop table if exists big_t");
exec(conn, "create table big_t as select range i, range % 7 bucket, 'row_' || range::varchar lbl from range(10000)");
check(cnt(conn, "select count(*) from big_t") == 10000, "cnt(conn, \"select count(*) from big_t\") == 10000");
check(q_one(conn, "select sum(i) from big_t") == 49995000, "q_one(conn, \"select sum(i) from big_t\") == 49995000");
check(len(q(conn, "select * from big_t limit 100")) == 100, "len(q(conn, \"select * from big_t limit 100\")) == 100");
check(len(qmap(conn, "select * from big_t limit 100")) == 100, "len(qmap(conn, \"select * from big_t limit 100\")) == 100");
check(cnt(conn, "select count(*) from (select distinct bucket from big_t)") == 7, "cnt(conn, \"select count(*) from (select distinct bucket from big_t)\") == 7");
check(len(qmap(conn, "select bucket, count(*) c from big_t group by bucket order by bucket")) == 7, "len(qmap(conn, \"select bucket, count(*) c from big_t group by bucket order by bucket\")) == 7");


# ---------------------------------------------------------------------------
# 17. cleanup
# ---------------------------------------------------------------------------

conn.close();
check(query_error(conn, "select 1") != null, "using a closed connection must error");

try {
    rm(f.workdir);
} catch (e) {
    println("cleanup: #{e}");
}
check(exists(f.workdir) == false, "workdir #{f.workdir} was not cleaned up");

print_report();
assert(len(report.failures) == 0, "#{len(report.failures)} duckdb checks are still failing, see the report above");
