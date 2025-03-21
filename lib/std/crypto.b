## `crypto` has crypto related functions that primarily operate
## on strings

val __sha = _sha;
val __md5 = _md5;
val __generate_from_password = _generate_from_password;
val __compare_hash_and_password = _compare_hash_and_password;
val __encrypt = _encrypt;
val __decrypt = _decrypt;
val __encode_base_64_32 = _encode_base_64_32;
val __decode_base_64_32 = _decode_base_64_32;
val __encode_hex = _encode_hex;
val __decode_hex = _decode_hex;

fun md5(content) {
    ##std:this,__md5
    ## `md5` is the stringified version of the md5 sum for the string or bytes content passed in
    ##
    ## md5(content: str|bytes) -> str
    __md5(content)
}

fun generate_from_password(password) {
    ##std:this,__generate_from_password
    ## `generate_from_password` will return a bcrypt hash string from the given password string
    ##
    ## generate_from_password(password: str) -> str
    __generate_from_password(password)
}

fun compare_hash_and_password(password, hashed_pw) {
    ##std:this,__compare_hash_and_password
    ## `compare_hash_and_password` will return true if the given password matches the hashed password
    ##
    ## compare_hash_and_password(password: str, hashed_pw: str) -> bool
    __compare_hash_and_password(password, hashed_pw)
}

fun sha(content, algo=256) {
    ##std:this,__sha
    ## `sha` will take a string or bytes and compute the sha1/256/512 value and return a string
    ##
    ## sha(content: str|bytes, algo: 1|256|512=256) -> str
    __sha(content, algo)
}

fun encrypt(pw, data) {
    ##std:this,__encrypt
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
    ##std:this,__decrypt
    ## `decrypt` will take a pw and the data to decrypt and decrypt it
    ## with a key derived from the pw
    ##
    ## if the data was initially a string then it will return the string
    ## otherwise, as_bytes should be set to true to return it as bytes
    ## instead
    ##
    ## decrypt(pw: str|bytes, data: bytes, as_bytes: bool=false) -> str|bytes
    __decrypt(pw, data, as_bytes)
}

fun encode(data, as_bytes=false, method='hex') {
    ##std:this,__encode_hex,__encode_base_64_32
    ## `encode` will take data as STRING or BYTES and encode it with the specified method
    ## supported methods are 'hex', 'base32', or 'base64'
    ##
    ## as_bytes determines if the value returned should be BYTES or STRING
    ##
    ## encode(data: str|bytes, as_bytes: bool=false, method: str='hex') -> str|bytes
    return match method {
        'hex' => {
            __encode_hex(data, as_bytes)
        },
        'base64' => {
            __encode_base_64_32(data, as_bytes, true)
        },
        'base32' => {
            __encode_base_64_32(data, as_bytes, false)
        },
        _ => {
            error("method #{method} not supported for encoding. expected hex, base64, or base32")
        },
    };
}

fun decode(data, as_bytes=false, method='hex') {
    ##std:this,__decode_hex,__decode_base_64_32
    ## `decode` will take data as STRING or BYTES and decode it with the specified method
    ## supported methods are 'hex', 'base32', or 'base64'
    ##
    ## as_bytes determines if the value returned should be BYTES or STRING
    ##
    ## decode(data: str|bytes, as_bytes: bool=false, method: str='hex') -> str|bytes
    return match method {
        'hex' => {
            __decode_hex(data, as_bytes)
        },
        'base64' => {
            __decode_base_64_32(data, as_bytes, true)
        },
        'base32' => {
            __decode_base_64_32(data, as_bytes, false)
        },
        _ => {
            error("method #{method} not supported for decoding. expected hex, base64, or base32")
        },
    };
}