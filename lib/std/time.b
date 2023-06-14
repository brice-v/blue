## `time` is the module that contains time related functions

val __now = _now;
val __sleep = _sleep;
val __parse = _parse;
val __to_str = _to_str;

fun now() {
    ## `now` returns the current unix timestamp as an int
    ##
    ## now() -> int
    __now()
}

fun sleep(ms) {
    ## `sleep` will pause execution of the current process for the given number of milliseconds
    ##
    ## sleep(ms: int) -> null
    __sleep(ms)
}

fun parse(date_time_str) {
    ## `parse` will parse the string as a date/time and return the unix timestamp as an int
    ##
    ## parse(date_time_str: str) -> int
    __parse(date_time_str)
}

fun to_str(time_as_unix_timestamp, timezone=null) {
    ## `to_str` will take the int unix timestamp and convert it to a human readable date/time string
    ##
    ## to_str(time_as_unix_timestamp: int, timezone: null|str=null) -> str
    __to_str(time_as_unix_timestamp, timezone)
}

val timezone = {
    'Local': 'Local',
    'UTC': 'UTC',
    'GMT': 'GMT',
    'CST': 'CST',
    'EET': 'EET',
    'WET': 'WET',
    'CET': 'CET',
    'EST': 'EST',
    'MST': 'MST',
    'Cuba': 'Cuba',
    'Egypt': 'Egypt',
    'Eire': 'Eire',
    'Greenwich': 'Greenwich',
    'Iceland': 'Iceland',
    'Iran': 'Iran',
    'Israel': 'Israel',
    'Jamaica': 'Jamaica',
    'Japan': 'Japan',
    'Libya': 'Libya',
    'Poland': 'Poland',
    'Portugal': 'Portugal',
    'PRC': 'PRC',
    'Singapore': 'Singapore',
    'Turkey': 'Turkey',
    'Shanghai': 'Asia/Shanghai',
    'Chongqing': 'Asia/Chongqing',
    'Harbin': 'Asia/Harbin',
    'Urumqi': 'Asia/Urumqi',
    'HongKong': 'Asia/Hong_Kong',
    'Macao': 'Asia/Macao',
    'Taipei': 'Asia/Taipei',
    'Tokyo': 'Asia/Tokyo',
    'Saigon': 'Asia/Saigon',
    'Seoul': 'Asia/Seoul',
    'Bangkok': 'Asia/Bangkok',
    'Dubai': 'Asia/Dubai',
    'NewYork': 'America/New_York',
    'LosAngeles': 'America/Los_Angeles',
    'Chicago': 'America/Chicago',
    'Moscow': 'Europe/Moscow',
    'London': 'Europe/London',
    'Berlin': 'Europe/Berlin',
    'Paris': 'Europe/Paris',
    'Rome': 'Europe/Rome',
    'Sydney': 'Australia/Sydney',
    'Melbourne': 'Australia/Melbourne',
    'Darwin': 'Australia/Darwin',
}

# TODO: Eventually include methods to do operations for things like + 1 day, - 1 month, etc.