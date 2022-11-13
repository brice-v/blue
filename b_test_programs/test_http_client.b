import http
import time

import test_http_server

fun main() {
    val server_pid = spawn(test_http_server.run, []);
    println("server_pid = #{server_pid}");
    # This makes sure the server is up
    for (true) {
        time.sleep(1000);
        try {
            http.get("http://localhost:3001/");
            break;
        } catch (ignored) {
            continue;
        }
    }
    val resp = http.get("http://localhost:3001/abc/AAAA/BBBB?name=Brice&email=brice%40example.com")
    println("resp = `#{resp}`");
    val expected_resp = "<p>url params: a=AAAA, b=BBBB | query_params: name=Brice, email=brice@example.com, headers="
    assert(resp.startswith(expected_resp), "Response for get_handler_for_multiple_things did not return expected result");
    assert("Host: localhost:3001" in resp, "Response for get_handler_for_multiple_things did not return expected result");
    assert("User-Agent: blue/v" in resp, "Response for get_handler_for_multiple_things did not return expected result");
    assert("Accept-Encoding: gzip" in resp, "Response for get_handler_for_multiple_things did not return expected result");

    val resp1 = http.post("http://localhost:3001/post/abc/213/?name=john&pass=doe", null, mime_type="application/x-www-form-urlencoded");
    println("resp1 = `#{resp1}`");
    val expected_resp1 = '{"name":"john","pass":"doe","a":"abc","b":"213"}';
    assert(resp1 == expected_resp1, "Response for post_handler with url form data did not return expected");

    val json_data1 = '{"name":"john","pass":"doe"}';
    val resp2 = http.post("http://localhost:3001/post/abc/213", json_data1);
    println("resp2 = `#{resp2}`");
    val expected_resp2 = '{"name":"john","pass":"doe","a":"abc","b":"213"}';
    assert(resp2 == expected_resp2, "Response for post_handler with json data did not return expected");

    val json_data2 = '{"name":"john","pass":["doe",2]}';
    val resp3 = http.post("http://localhost:3001/post/json/213", json_data2);
    println("resp3 = `#{resp3}`");
    val expected_resp3 = '{"name":"john","pass":["doe",2.000000],"a":"json","b":"213"}';
    assert(resp3 == expected_resp3, "Response for post_handler with json data (with list) did not return expected");

    val xml_data1 = "<login><name>john</name><pass>doe</pass></login>";
    val resp4 = http.post("http://localhost:3001/post/xml/data1", xml_data1, mime_type="application/xml");
    println("resp4 = `#{resp4}`");
    val expected_resp4 = '{"name":"john","pass":"doe","a":"xml","b":"data1"}';
    assert(resp4 == expected_resp4, "Response for post_handler, with xml data did not return expected");

    val form_data1 = "name=john&pass=doe";
    val resp5 = http.post("http://localhost:3001/post/FORM/data", form_data1, mime_type="application/x-www-form-urlencoded");
    println("resp5 = `#{resp5}`");
    val expected_resp5 = '{"name":"john","pass":"doe","a":"FORM","b":"data"}';
    assert(resp5 == expected_resp5, "Response for post_handler, with form encoded as body url encoded, did not return expected");

    val resp6 = http.get("http://localhost:3001/json");
    println("resp6 = `#{resp6}`");
    val expected_resp6 = '{"a":1,"b":2,"c":["d","e","f"]}';
    assert(resp6 == expected_resp6, "Response for get handler with json resp did not return expected");

    val resp7 = http.get("http://localhost:3001/hello/Someone");
    println("resp7 = `#{resp7}`");
    val expected_resp7 = '<p><b>Hello Someone!</b></p>';
    assert(resp7 == expected_resp7, "Response for get handler with param did not return expected");

    val ws = http.new_ws("ws://localhost:3001/ws");
    #for (true) {
        var x = "Sending from Client!";
        ws.send(x);
        val y = ws.recv();
        println("Received on Client, y = #{y}");
        assert(x == y, "Websocket Handler did not echo the sent value from the client");
    #}
    http.shutdown_server();
}

main();
true;