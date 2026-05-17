import http

val url = "https://go.dev"
val expected = @{
    scheme: 'https',
    opaque: '',
    username: '',
    password: null,
    host: 'go.dev',
    path: '',
    fragment: '',
    raw_query: '',
    raw_path: '',
    raw_fragment: '',
    force_query: false,
    omit_host: false
};
assert(expected == http.url_parse(url));