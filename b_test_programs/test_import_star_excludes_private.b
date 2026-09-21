# A wildcard import pulls in public names only. The module's private (underscore
# prefixed) names stay private, while public functions can still call them.
from star_private_mod import *

assert(public_helper() == "helper");
assert(public_secret() == "secret");
