## `crypto` has crypto related functions that primarily operate
## on strings

val __sha = _sha;
val __md5 = _md5;
val __generate_from_password = _generate_from_password;
val __compare_hash_and_password = _compare_hash_and_password;
val __encrypt = _encrypt;
val __decrypt = _decrypt;

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

fun encrypt(pw, data) {
    ## `encrypt` wil take a pw and the data to encrypt and encrypt it
    ## with a key made from the pw
    ##
    ## it will always return bytes as theres no guarantees its a valid
    ## string after being encrypted.
    ##
    ## encrypt(pw: str|bytes, data: str|bytes) -> bytes
    __encrypt(pw, data)
}

fun decrypt(pw, data, as_bytes=false) {
    ## `decrypt` wil take a pw and the data to decrypt and decrypt it
    ## with a key derived from the pw
    ##
    ## if the data was initially a string then it will return the string
    ## otherwise, as_bytes should be set to true to return it as bytes
    ## instead
    ##
    ## decrypt(pw: str|bytes, data: bytes, as_bytes: bool=false) -> str|bytes
    __decrypt(pw, data, as_bytes)
}