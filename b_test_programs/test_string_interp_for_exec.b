
val output_file_name = "bal.com"
var cmd_str2 = "objcopy -S -O binary #{output_file_name}.dbg #{output_file_name}";
val expected = "objcopy -S -O binary bal.com.dbg bal.com";
println("cmd_str2 = `#{cmd_str2}`, output_file_name = #{output_file_name}")

assert(cmd_str2 == expected);