# Regression test for the "'timezone' is immutable" bug
#
# A function parameter with a default value that shares its name with a module
# level val would fail to compile. The compiler resolved the assignment of the
# default parameter to the qualified module level symbol instead of the local
# parameter shadowing it.

import shadowmod

assert(shadowmod.fmt(5) == 'ts=5 tz=UTC')
assert(shadowmod.fmt(5, 'EST') == 'ts=5 tz=EST')
assert(shadowmod.fmt(5, null) == 'ts=5 tz=UTC')

assert(shadowmod.idx() == [99, 2, 3])
assert(shadowmod.idx(['a', 'b']) == [99, 'b'])

assert(shadowmod.bump() == 1)
assert(shadowmod.bump(5) == 6)
assert(shadowmod.counter == 6)
