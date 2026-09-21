# A selective import compiles only the requested declarations plus whatever those
# declarations reference. The private helper and the module level value it uses
# come along, while `public_unused` is left out entirely.
from selective_mod import {public_double}

assert(public_double(5) == 30);

