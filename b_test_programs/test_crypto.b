import crypto

var s = "A sample string to SHA256!";
var sha256result = crypto.sha(s, algo=256);
assert(sha256result == '9abf637c7e39cc4ef84d6d92cf7ffe168dc922b8ae666260d907e0353865ce89');

var s1 = "The fog is getting thicker!And Leon's getting laaarger!";
var md5result = crypto.md5(s1);
assert(md5result == 'e2c569be17396eca2a2e3c11578123ed');


var pw = 'My Password!';
var hashedPw = crypto.generate_from_password(pw);
var comparedPw = crypto.compare_hash_and_password(hashedPw, pw);
println(hashedPw);
assert(hashedPw != '$2a$10$JqHhRzcXDffvPrpNkSI1BeAx7DfkTjKMWqEsB802d8mPBbhWO1AuS');
println(comparedPw);
assert(comparedPw);