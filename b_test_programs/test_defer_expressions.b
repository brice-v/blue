val fname = "test_defer_expressions_test.txt";
fun() {
    fun appendToTestFile(content) {
        assert(type(content) == Type.STRING);
        
        try {
            var fcontent = fname.read();
            fname.write(fcontent+content);
        } catch (e) {
            fname.write(content);
        }
    }
    
    fun main() {
        val p1 = fun() { appendToTestFile("Hello 1"); }
        val p2 = fun() { appendToTestFile("Hello 2"); }
        val p3 = fun() { appendToTestFile("Hello 3"); }
        defer(p3);
        defer(p2);
        defer(p1);
        fun() {
            val p0 = fun() { appendToTestFile("Hello 0"); };
            defer(p0);
        }()
    }
    
    defer(fun(){ appendToTestFile("Hello 4"); });
    main();
}()

var current_fcontent = fname.read();
println("current_fcontent = #{current_fcontent}")
assert(current_fcontent == 'Hello 0Hello 1Hello 2Hello 3Hello 4');
fname.rm()
assert(true);