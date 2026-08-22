# Whole number floats keep a decimal place when formatted as strings so
# they stay visually distinct from integers

assert(str(5.0) == '5.0')
assert(str(-7.0) == '-7.0')
assert(str(0.5) == '0.5')
assert('#{2.0}' == '2.0')
assert("#{1.5 + 0.5}" == '2.0')

assert(str([1.5, 2.0]) == '[1.5, 2.0]')
assert(str({a: 3.0}) == '{a: 3.0}')

# exponent forms and non finite values are left untouched
assert(str(1e8) == '1e+08')
assert(str(to_num('+Inf')) == '+Inf')
assert(str(to_num('-Inf')) == '-Inf')

# integers are unaffected
assert(str(5) == '5')
