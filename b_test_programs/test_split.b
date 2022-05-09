val hw = "Hello World";
val h = "Hello";
val w = "World";

if ("Hello World".split(" ") != ["Hello", "World"]) {
    return false;
}

if ("Hello World".split(" ") != [h, w]) {
    return false;
}

if (hw.split(" ") != ["Hello", "World"]) {
    return false;
}

if (hw.split(" ") != [h, w]) {
    return false;
}

if (hw.split() != [h, w]) {
    return false;
}

if (split("Hello World") != ["Hello", "World"]) {
    return false;
}

if (split("Hello World", " ") != ["Hello", "World"]) {
    return false;
}

if (split(hw, " ") != ["Hello", "World"]) {
    return false;
}

if (split(hw) != [h, "World"]) {
    return false;
}

if (split("Hello World") != [h, w]) {
    return false;
}

if (split("Hello World", " ") != ["Hello", w]) {
    return false;
}

if (split(hw, " ") != [h, "World"]) {
    return false;
}

if (split(hw) != [h, w]) {
    return false;
}
true;