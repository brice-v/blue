fun main() {
    for (i in 0..10) {
        if (i == 0) {
            return true;
        }
    }
    return false;
}

assert(main(), "This should return true")