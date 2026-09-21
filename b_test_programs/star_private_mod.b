# Fixture for test_import_star_excludes_private.b. The underscore prefixed names
# are private to this module and must not be pulled in by `from ... import *`.
val _secret = "secret";

fun _helper() {
    "helper";
}

fun public_helper() {
    _helper();
}

fun public_secret() {
    _secret;
}

assert(true);
