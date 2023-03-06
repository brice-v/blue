import crypto

var x = "SOME SECRET STRING";
var pw = "12345";

val data = crypto.encrypt(pw, x);
println("(after encrypt) data = #{data}");
assert(data.type() == 'BYTES');

val resp = crypto.decrypt(pw, data);
println("(after decrypt) resp = #{resp}");
assert(resp == x);