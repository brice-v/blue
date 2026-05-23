import http

try {
    http.post("http://localhost:3001", body={'data': 'test'});
} catch (e) {
    assert(e == 'function called with default argument that is not in default function parameters' || e == "wrong number of arguments: want=4, got=2");
}
assert(true);