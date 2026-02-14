#VM IGNORE
# TODO: vm will support destructuring
var test_input ="""467..114..
...*......
..35..633.
......#...
617*......
.....+.58.
..592.....
......755.
...$.*....
.664.598..""".replace("\r", "");

fun is_digit(c) {
    if c == null {
        return false;
    }
    return c in "0123456789";
}

fun is_valid_digit_and_gear_near(split_lines, i, j) {
    val line_len = split_lines[0].len() - 1;
    val tot_lines_len = split_lines.len() - 1;
    var gear_poss = [];
    val top_left = if i == 0 { null } else {
        if j - 1 < 0 {
            null
        } else {
            split_lines[i-1][j-1]
        }
    };
    if top_left == '*' {
        gear_poss << [i-1,j-1];
    }
    val top_mid = if i == 0 { null } else {
        split_lines[i-1][j] 
    };
    if top_mid == '*' {
        gear_poss << [i-1,j];
    }
    val top_right = if i == 0 { null } else {
        if j + 1 > line_len {
            null
        } else {
            split_lines[i-1][j+1]
        }
    };
    if top_right == '*' {
        gear_poss << [i-1,j+1];
    }
    val bot_left = if i == tot_lines_len { null } else {
        if j - 1 < 0 {
            null
        } else {
            split_lines[i+1][j-1]
        }
    };
    if bot_left == '*' {
        gear_poss << [i+1,j-1];
    }
    val bot_mid = if i == tot_lines_len { null } else {
        split_lines[i+1][j] 
    };
    if bot_mid == '*' {
        gear_poss << [i+1,j];
    }
    val bot_right = if i == tot_lines_len { null } else {
        if j + 1 > line_len {
            null
        } else {
            split_lines[i+1][j+1]
        }
    };
    if bot_right == '*' {
        gear_poss << [i+1,j+1];
    }
    val right = if j + 1 > line_len {
        null
    } else {
        split_lines[i][j+1]
    };
    if right == '*' {
        gear_poss << [i,j+1];
    }
    val left = if j - 1 < 0 {
        null
    } else {
        split_lines[i][j-1]
    }
    if left == '*' {
        gear_poss << [i,j-1];
    }
    #println("[i = #{i}, j = #{j}] top_left = #{top_left}, top_mid = #{top_mid}, top_right = #{top_right}");
    #println("[i = #{i}, j = #{j}] bot_left = #{bot_left}, bot_mid = #{bot_mid}, bot_right = #{bot_right}");
    #println("[i = #{i}, j = #{j}] left = #{left}, right = #{right}")
    val all_of_them = [top_left,top_mid,top_right,bot_left,bot_mid,bot_right,left,right];
    var is_valid = false;
    for (one in all_of_them) {
        is_valid = is_valid || (one != null && one != '.' && !is_digit(one));
    }
    println("after here")
    #println("is_valid = #{is_valid}")
    return [is_valid,gear_poss];
}

fun is_next_char_digit(split_lines, i, j) {
    val line_len = split_lines[0].len() - 1;

    val next_char = if j + 1 > line_len {
        null
    } else {
        split_lines[i][j+1]
    };
    return is_digit(next_char);
}

fun get_valid_gear_combos(valid_nums) {
    var nums = set([]);
    for var i = 0; i < len(valid_nums); i += 1 {
        var valid_num = valid_nums[i];
        for var j = 0; j < len(valid_nums); j += 1 {
            if j == i {
                continue;
            }
            if (valid_num.gears & valid_nums[j].gears) != set([]) {
                var this = [valid_num.num,valid_nums[j].num];
                var that = [valid_nums[j].num,valid_num.num];
                if this notin nums && that notin nums {
                    nums |= {this};
                }
            }
        }
    }
    println("nums = #{nums}");
    return nums;
}

fun part2(_input) {
    val split_lines = _input.split("\n");
    var valid_nums = [];
    println("split_lines = #{split_lines}")
    for var i = 0; i < len(split_lines); i += 1 {
        var current_num = "";
        var current_num_is_valid = false;
        var all_gear_pos_near = set([]);
        for [j,c] in split_lines[i] {
            # if the c is a . just continue
            if c == '.' {
                continue;
            }
            if is_digit(c) {
                current_num += c;
                # Check Above diagnols and vertical
                val [vd,gear_poss] = is_valid_digit_and_gear_near(split_lines, i, j);
                current_num_is_valid = current_num_is_valid || vd;
                # Another bug here with scopes when I did for gear in gear_poss
                for (gear in gear_poss) {
                    all_gear_pos_near |= {gear};
                }
            }
            #assert(false);
            if is_digit(c) && !is_next_char_digit(split_lines, i, j) {
                println("current_num = #{current_num}, current_num_is_valid = #{current_num_is_valid}, all_gear_pos_near = #{all_gear_pos_near}")
                if current_num_is_valid {
                    valid_nums << {'num': int(current_num), 'gears': all_gear_pos_near};
                }
                current_num = "";
                current_num_is_valid = false;
                all_gear_pos_near = set([]);
            }
        }
    }
    val nums = get_valid_gear_combos(valid_nums);
    var sum = 0;
    for (e in nums) {
        sum += (e[0]*e[1]);
    }
    assert(sum == 467835)
}

part2(test_input);

assert(true);