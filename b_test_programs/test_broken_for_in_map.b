val stuff = {
    '2022': {
        '12': {
            '01': {'thing': 'thing content', 'thing1': 'thing1 content'},
            '02': {'thing2': 'thing2 content'},
        },
    },
    '2023': {
        '08': {
            '03': {'thing3': 'thing3 content'},
        },
        '05': {
            '04': {'thing4': 'thing4 content'},
        },
    },
};

for ([year, month_posts] in stuff) {
    for ([month, day_posts] in month_posts) {
        for ([day, posts] in day_posts) {
            for ([fname, post] in posts) {
                println('fname = #{fname}, post = #{post}');
            }
        }
    }
}
true;