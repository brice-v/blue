import crypto

val to_encode = "49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d";
val hex_to_encode = crypto.decode(to_encode, as_bytes=true);
val encoded = crypto.encode(hex_to_encode, method='base64');
println("encoded = #{encoded}");
var expected = "SSdtIGtpbGxpbmcgeW91ciBicmFpbiBsaWtlIGEgcG9pc29ub3VzIG11c2hyb29t";
assert(encoded == expected);


val hex_str = "1c0111001f010100061a024b53535009181c";
val hex_str_decoded = crypto.decode(hex_str, as_bytes=true);
val other_hex_str = "686974207468652062756c6c277320657965";
val other_hex_str_decoded = crypto.decode(other_hex_str, as_bytes=true);

val xord = hex_str_decoded^other_hex_str_decoded;
var actual_xord = crypto.encode(xord);
println("xord = #{actual_xord}");
expected = "746865206b696420646f6e277420706c6179";
assert(actual_xord == expected);

val hex_enc_str = "1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736";
val dec_hex_enc_str = crypto.decode(hex_enc_str, as_bytes=true);
var alphabet = "abcdefghijklmnopqrstuvwxyz"+("abcdefghijklmnopqrstuvwxyz".to_upper());
println("alphabet = `#{alphabet}`");
fun print_if_multiple_word(s, char_used) {
    val words = s.find_all("([A-z])\\w+ ");
    if (len(words) > 1) {
        return {char: char_used, found: s};
    }
    return null;
}
for (c in alphabet) {
    var toXorAgainst = (c * len(dec_hex_enc_str)).to_bytes();    
    var afterXord = dec_hex_enc_str^toXorAgainst;
    var result = print_if_multiple_word(str(afterXord), c);
    if (result == null) {
        continue;
    } else {
        println("result = #{result}");
        assert(result.char == 'X');
        assert(result.found == "Cooking MC's like a pound of bacon");
    }
}
assert(true);