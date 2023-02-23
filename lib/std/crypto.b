## `crypto` has crypto related functions that primarily operate
## on strings

val __sha = _sha;
val __md5 = _md5;
val __generate_from_password = _generate_from_password;
val __compare_hash_and_password = _compare_hash_and_password;

fun md5(content) {
    ## `md5` is the stringified version of the md5 sum for the string content passed in
    ##
    ## md5(content: str) -> str
    __md5(content)
}

fun generate_from_password(password) {
    ## `generate_from_password` will return a bcrypt hash string from the given password string
    ##
    ## generate_from_password(password: str) -> str
    __generate_from_password(password)
}

fun compare_hash_and_password(password, hashed_pw) {
    ## `compare_hash_and_password` will return true if the given password matches the hashed password
    ##
    ## compare_hash_and_password(password: str, hashed_pw: str) -> bool
    __compare_hash_and_password(password, hashed_pw)
}

fun sha(str_to_hash, algo=256) {
    ## `sha` wil take a string and compute the sha1/256/512 value and return a string
    ##
    ## sha(str_to_hash: str, algo: 1|256|512=256) -> str
    __sha(str_to_hash, algo)
}