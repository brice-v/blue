try {
    var output = `ls`;

    println(output);

    var out = `ls  -l`;

    print(out);
} catch (ignored) {
    var output = `dir`;

    println(output);
}

return true;