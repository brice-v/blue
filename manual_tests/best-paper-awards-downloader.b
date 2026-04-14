import http

fun main() {
    val fname = "manual_tests/best-paper-awards-html.txt"
    if !is_file(fname) {
        val url = 'https://jeffhuang.com/best_paper_awards/';
        val resp = http.get(url);
        println("resp = #{resp}")
        fname.write(resp);
    }
    println("is_file(#{fname}) = #{is_file(fname)}");
    val file_data = fname.read();
    val trs = file_data.find_all("//tr", method='xpath');
    for tr in trs {
        #'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
        println("tr = #{tr}\n\n");
        val result1 = tr.find_one("//a[@href]", method='xpath');
        val result2 = tr.find_one("//a[@href]", method='regex')
        println("result1 = #{result1}")
        println("result2 = #{result2}")
        #'BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB'
    }
}

main();