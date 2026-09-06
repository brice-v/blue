import time

# helper to normalize Windows backslashes to forward slashes for cross-platform asserts
fun _norm(p) {
    replace(p, "\\", "/")
}

var tmp_root = temp_dir("", "blue_test_everyday_*")
assert(exists(tmp_root))
assert(is_dir(tmp_root))

# path core map `path` (from lib/core/core.b) via private _path_* builtins
# Use _norm() so tests pass on both Unix (/) and Windows (\) separators
assert(_norm(path.join("/tmp", "a", "b")) == "/tmp/a/b")
assert(_norm(path.clean("a//b/../c")) == "a/c")
assert(_norm(path.dir("/tmp/a/b.txt")) == "/tmp/a")
assert(path.base("/tmp/a/b.txt") == "b.txt")
assert(path.ext("archive.tar.gz") == ".gz")
assert(path.is_abs("/tmp") == true)
assert(path.is_abs("a/b") == false)
assert(_norm(path.rel("/tmp/a", "/tmp/a/b/c")) == "b/c")
# also test private helpers still work
assert(_norm(_path_join("/x", "y")) == "/x/y")
assert(_norm(_path_clean("a/./b")) == "a/b")

var nested = path.join(tmp_root, "a", "b", "c")
mkdir_all(nested)
assert(is_dir(nested))
var file_a = path.join(nested, "hello.txt")
_write(file_a, "hello blue")
assert(exists(file_a))
assert(is_file(file_a))
var st = stat(file_a)
assert(st.size == 10)
assert(st.is_dir == false)
assert(st.name == "hello.txt")
assert(has(st, "mtime"))

var file_b = path.join(nested, "copy.txt")
cp(file_a, file_b)
assert(exists(file_b))
var file_c = path.join(tmp_root, "moved.txt")
mv(file_b, file_c)
assert(exists(file_c))
assert(exists(file_b) == false)
rename(file_c, file_b)
assert(exists(file_b))

var glob_res = glob(path.join(nested, "*.txt"))
assert(glob_res.len() == 2)
var walk_res = walk(tmp_root)
assert(walk_res.len() >= 4)

var tmp_file = temp_file(tmp_root, "tmpfile_*")
assert(exists(tmp_file))
assert(is_file(tmp_file))

# ENV
setenv("BLUE_TMP_TEST", "xyz123")
assert(getenv("BLUE_TMP_TEST") == "xyz123")
assert(getenv("BLUE_MISSING_XYZ", "fallback") == "fallback")
assert(getenv("BLUE_MISSING_XYZ") == null)
var env_all = environ()
assert(has(env_all, "BLUE_TMP_TEST"))
unsetenv("BLUE_TMP_TEST")
assert(getenv("BLUE_TMP_TEST", "gone") == "gone")

# JSON pretty
var obj = {"a": 1, "b": [2,3], "c": {"nested": true}}
var pretty = to_json_pretty(obj, 2)
assert(has(pretty, "\n"))
assert(has(pretty, "\"a\""))
var pretty2 = to_json_pretty(obj, "    ")
assert(pretty2.len() > pretty.len())

# Collections: get/has/flat/uniq/chunk
var m = {"a": 1, "b": 2}
assert(get(m, "a", 0) == 1)
assert(get(m, "z", 99) == 99)
assert(get(m, "z") == null)
assert(has(m, "a") == true)
assert(has(m, "z") == false)
assert(flat([1,[2,3],4]) == [1,2,3,4])
assert(uniq([1,2,2,3,1]) == [1,2,3])
assert(chunk([1,2,3,4,5], 2) == [[1,2],[3,4],[5]])
assert(get([10,20,30], -1) == 30)
assert(get([10,20,30], 99, "miss") == "miss")
assert(has([1,2,3], 2) == true)
assert(has("hello", "ell") == true)
assert(get("hello", 1) == "e")

# String
assert(repeat("ab", 3) == "ababab")
assert(trim_prefix("foobar", "foo") == "bar")
assert(trim_suffix("foobar", "bar") == "foo")
assert(trim_prefix("hello", "xyz") == "hello")

# Time via core _time_* (private) and via import time module
assert(_time_add(1000, 500) == 1500)
assert(_time_format(1700000000000, "date") == "2023-11-14")
assert(time.add(1000, 500) == 1500)
assert(time.format(1700000000000, "date") == "2023-11-14")
assert(time.format(1700000000000, "2006/01/02") == "2023/11/14")
# time.Unit and time.now still work
assert(time.Unit.DAY == 86400000)
assert(time.now() > 0)

# Path via import path std module (core map `path` also provides same)
assert(_norm(path.join("/x", "y")) == "/x/y")
assert(_norm(path.clean("a//b")) == "a/b")
assert(_norm(path.dir("/a/b/c.txt")) == "/a/b")
assert(path.base("/a/b/c.txt") == "c.txt")
assert(path.ext("a.tar.gz") == ".gz")
assert(path.is_abs("/a") == true)
assert(_norm(path.rel("/a/b", "/a/b/c/d")) == "c/d")
assert(_norm(path.abs("a/b")) == _norm(_abs_path("a/b")))
assert(_norm(path.abs("/tmp/a")) == _norm(_abs_path("/tmp/a")))

# cleanup
rm(tmp_root)
assert(exists(tmp_root) == false)
