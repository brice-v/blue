# Test bitwise operations: AND, OR, XOR, NOT, left shift, right shift

# Bitwise AND
assert(5 & 3 == 1)    # 101 & 011 = 001
assert(7 & 7 == 7)    # 111 & 111 = 111
assert(0 & 123 == 0)  # 000 & xxx = 000
assert(255 & 15 == 15) # 11111111 & 00001111 = 00001111

# Bitwise OR
assert(5 | 3 == 7)    # 101 | 011 = 111
assert(0 | 7 == 7)    # 000 | 111 = 111
assert(4 | 2 == 6)    # 100 | 010 = 110
assert(255 | 0 == 255)

# Bitwise XOR
assert(5 ^ 3 == 6)    # 101 ^ 011 = 110
assert(7 ^ 7 == 0)    # 111 ^ 111 = 000
assert(0 ^ 5 == 5)    # 000 ^ 101 = 101
assert(15 ^ 15 == 0)

# Bitwise NOT (unary prefix)
assert(~0 == -1)      # ~000...000 = 111...111 = -1
assert(~1 == -2)      # ~000...001 = 111...110 = -2
assert(~-1 == 0)      # ~111...111 = 000...000 = 0

# Left shift
assert(1 << 0 == 1)   # 1 << 0 = 1
assert(1 << 1 == 2)   # 1 << 1 = 2
assert(1 << 2 == 4)   # 1 << 2 = 4
assert(1 << 3 == 8)   # 1 << 3 = 8
assert(1 << 4 == 16)  # 1 << 4 = 16
assert(1 << 8 == 256) # 1 << 8 = 256
assert(3 << 2 == 12)  # 011 << 2 = 1100 = 12

# Right shift
assert(4 >> 1 == 2)   # 100 >> 1 = 10
assert(8 >> 1 == 4)   # 1000 >> 1 = 100
assert(8 >> 2 == 2)   # 1000 >> 2 = 10
assert(16 >> 3 == 2)  # 10000 >> 3 = 10
assert(1 >> 1 == 0)   # 1 >> 1 = 0

# Combined operations
assert((1 << 4) | (1 << 2) | 1 == 21)  # 10000 | 100 | 1 = 10101 = 21
assert((1 << 8) & 255 == 1)             # 256 & 255 = 0
assert((1 << 5) ^ (1 << 3) == 48)       # 32 ^ 8 = 40... wait

# Let me recalculate: 32 = 100000, 8 = 001000, XOR = 101000 = 40
assert((1 << 5) ^ (1 << 3) == 40)

# Shift with larger numbers
assert(1 << 10 == 1024)
assert(1 << 16 == 65536)
assert(1024 >> 4 == 64)

# Compound bitwise assignment
var x = 15  # 1111
x &= 5      # 1111 & 0101 = 0101 = 5
assert(x == 5)

x = 15
x |= 1      # 1111 | 0001 = 1111 = 15
assert(x == 15)

x = 15
x ^= 10     # 1111 ^ 1010 = 0101 = 5
assert(x == 5)

x = 1
x <<= 3     # 1 << 3 = 8
assert(x == 8)

x = 16
x >>= 2     # 16 >> 2 = 4
assert(x == 4)

# Bitwise operations with 0
assert(0 & 0 == 0)
assert(0 | 0 == 0)
assert(0 ^ 0 == 0)
assert(~0 == -1)

# Bitwise operations with 1
assert(1 & 1 == 1)
assert(1 | 1 == 1)
assert(1 ^ 1 == 0)

# Complex expression
val result = ((1 << 7) | (1 << 5) | (1 << 3) | (1 << 1)) & 255
# 128 | 32 | 8 | 2 = 170
assert(result == 170)

# Shift precedence check
assert(1 << 2 + 1 == 8)   # 1 << (2+1) = 1 << 3 = 8