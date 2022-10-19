val sha_ = _sha;
val md5 = _md5;
val generate_from_password = _generate_from_password;
val compare_hash_and_password = _compare_hash_and_password;

fun sha(str_to_hash, algo=256) {
    sha_(str_to_hash, algo)
}