import crypto

val to_encode = "49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d";
val hex_to_encode = crypto.decode(to_encode, as_bytes=true);
val encoded = crypto.encode(hex_to_encode, method='base64');
# Set 1 problem 1
println("encoded = #{encoded}");


val hex_str = "1c0111001f010100061a024b53535009181c";
val hex_str_decoded = crypto.decode(hex_str, as_bytes=true);
val other_hex_str = "686974207468652062756c6c277320657965";
val other_hex_str_decoded = crypto.decode(other_hex_str, as_bytes=true);

val xord = hex_str_decoded^other_hex_str_decoded;

# Set 1 problem 2
println("xord = #{crypto.encode(xord)}");

val hex_enc_str = "1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736";
val dec_hex_enc_str = crypto.decode(hex_enc_str, as_bytes=true);
var alphabet = "abcdefghijklmnopqrstuvwxyz"+("abcdefghijklmnopqrstuvwxyz".to_upper());
println("alphabet = `#{alphabet}`");
fun print_if_multiple_word(s, char_used) {
    val words = s.find_all("([A-z])\\w+ ");
    if (len(words) > 1) {
        # set 1 problem 3 (also 4)
        println("char used = '#{char_used}', string = `#{s}`");
    }
}
for (c in alphabet) {
    var toXorAgainst = (c * len(dec_hex_enc_str)).to_bytes();    
    var afterXord = dec_hex_enc_str^toXorAgainst;
    print_if_multiple_word(str(afterXord), c);
}

###
val fdata = read('4.txt');
val rows = fdata.split("\n");
alphabet += "1234567890";

for (row in rows) {
    var enc_str_from_file = crypto.decode(row, as_bytes=true);
    for (c in alphabet) {
        var toXorAgainst = (c * len(enc_str_from_file)).to_bytes();
        var afterXord = enc_str_from_file^toXorAgainst;
        # println(str(afterXord));
        print_if_multiple_word(str(afterXord), c);
    }
}
###
# set 1 problem 4
#char used = '5', string = `Now that the party is jumping

