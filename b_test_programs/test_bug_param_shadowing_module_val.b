# Regression test for the "'timezone' is immutable" bug
#
# A function parameter with a default value that shares its name with a module
# level val would fail to compile. The compiler resolved the assignment of the
# default parameter to the qualified module level symbol instead of the local
# parameter shadowing it.

import shadowmod
from shadowmod import {dtype}

assert(shadowmod.fmt(5) == 'ts=5 tz=UTC')
assert(shadowmod.fmt(5, 'EST') == 'ts=5 tz=EST')
assert(shadowmod.fmt(5, null) == 'ts=5 tz=UTC')

assert(shadowmod.idx() == [99, 2, 3])
assert(shadowmod.idx(['a', 'b']) == [99, 'b'])

assert(shadowmod.bump() == 1)
assert(shadowmod.bump(5) == 6)
assert(shadowmod.counter == 6)

# The parameter name may also shadow the module level value it uses in its own
# default expression (the default is resolved in the enclosing scope).
assert(shadowmod.tensor([1, 2]) == [[1, 2], 'float32'])
assert(shadowmod.tensor([1, 2], 'int32') == [[1, 2], 'int32'])
assert(shadowmod.tensor([1, 2], dtype='int64') == [[1, 2], 'int64'])
assert(shadowmod.tensor([1, 2], dtype=dtype.float32) == [[1, 2], 'float32'])

assert(shadowmod.scale(4) == 12)
assert(shadowmod.scale(4, 5) == 20)
