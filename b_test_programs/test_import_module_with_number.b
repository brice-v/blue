# Test that import paths can contain numbers just like identifiers
#
# Regression test for the lexer restriction that only allowed letters and dots
# in import paths, module names with numbers (such as numbermod1) would lex as
# a truncated path at the first digit

import numbermod1

assert(numbermod1.fmt1(5) == 'ts=5 tz=UTC')
assert(numbermod1.fmt1(5, 'EST') == 'ts=5 tz=EST')
assert(numbermod1.fmt1(5, null) == 'ts=5 tz=UTC')

assert(numbermod1.bump1() == 1)
assert(numbermod1.bump1(5) == 6)
assert(numbermod1.counter1 == 6)

from numbermod1 import {tz1}

assert(tz1.UTC == 'UTC')
