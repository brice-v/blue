# Comprehensions inside an imported module used to fail to compile because the
# internal accumulator variable was defined with the module prefix but resolved
# without it.
import comprehension_mod

assert(comprehension_mod.scaled() == [3, 6, 9]);
assert(comprehension_mod.filtered() == [2, 4]);
assert(comprehension_mod.flattened() == [1, 2, 3, 4]);
assert(comprehension_mod.mapped() == {1: 1, 2: 4, 3: 9});
assert(comprehension_mod.setified() == {1, 2, 3});
assert(comprehension_mod.via_private() == [100, 200, 300]);
assert(comprehension_mod.nested() == [[1], [1, 2], [1, 2, 3]]);
assert(comprehension_mod.nested_from_iterable() == [1, 2, 3]);

# A selective import has to walk the comprehension's deferred program to keep the
# private helper it calls.
from comprehension_mod import {scaled, via_private, nested}

assert(scaled() == [3, 6, 9]);
assert(via_private() == [100, 200, 300]);
assert(nested() == [[1], [1, 2], [1, 2, 3]]);
