val line = '';

match {
    line.startswith('$') => {
        println("HERE1");
        assert(false);
    },
    line.startswith('dir') => {
        println("HERE2");
        assert(false);
    },
    _ => {
        println("HERE1");
        assert(true);
    },
}