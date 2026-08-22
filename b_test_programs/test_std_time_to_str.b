import time

# to_str with an explicit UTC timezone formats as expected
assert(time.to_str(0, 'UTC') == '1970-01-01 00:00:00')
assert(time.to_str(1703479130205, 'UTC') == '2023-12-25 04:38:50.205')
assert(time.to_str(1703479130000, 'UTC') == '2023-12-25 04:38:50')

# named timezone keys from the timezone map resolve to valid zone strings
assert(time.timezone.UTC == 'UTC')
assert(time.timezone.Tokyo == 'Asia/Tokyo')
assert(time.to_str(0, time.timezone.UTC) == '1970-01-01 00:00:00')

# null timezone (and the omitted default param) format in the machine local
# timezone so only structural checks are made here to stay machine independent
val local = time.to_str(0)
assert(local.len() == 19)
assert(local.split(' ').len() == 2)
assert(local.split(' ')[0].split('-').len() == 3)
assert(local.split(' ')[1].split(':').len() == 3)
assert(local == time.to_str(0, null))

# parse and to_str round trip through UTC preserving milliseconds
val ts = 1699999999999
assert(time.parse(time.to_str(ts, 'UTC')) == ts)

# invalid arguments surface catchable errors
var caught_type = false
try {
    time.to_str('nope', 'UTC')
} catch (e) {
    caught_type = true
}
assert(caught_type)

var caught_tz_type = false
try {
    time.to_str(0, ['UTC'])
} catch (e) {
    caught_tz_type = true
}
assert(caught_tz_type)
