var ver = `java -version`.split("\n")[0];
if ("openjdk" notin ver) {
    return false;
}

true;