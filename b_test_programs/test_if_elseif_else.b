val input = """$ cd /
$ ls
dir a
14848514 b.txt
8504156 c.dat
dir d
$ cd a
$ ls
dir e
29116 f
2557 g
62596 h.lst
$ cd e
$ ls
584 i
$ cd ..
$ cd ..
$ cd d
$ ls
4060174 j
8033020 d.log
5626152 d.ext
7214296 k""".replace("\r", "");

val expected = """cmd = $ cd /cmd = $ lsdir = dir afile = 14848514 b.txtfile = 8504156 c.datdir = dir dcmd = $ cd acmd = $ lsdir = dir efile = 29116 ffile = 2557 gfile = 62596 h.lstcmd = $ cd ecmd = $ lsfile = 584 icmd = $ cd ..cmd = $ cd ..cmd = $ cd dcmd = $ lsfile = 4060174 jfile = 8033020 d.logfile = 5626152 d.extfile = 7214296 kcmd = $ cd /cmd = $ lsdir = dir adir = dir dcmd = $ cd acmd = $ lsdir = dir ecmd = $ cd ecmd = $ lscmd = $ cd ..cmd = $ cd ..cmd = $ cd dcmd = $ ls""";

val lines = input.split("\n");

var result = "";
for ([i, line] in lines) {
    #println("line = #{line}");
    var current_dir = '';
    if (line.startswith("$")) {
        # Handle cmd
        result += "cmd = #{line}";
    } else if (line.startswith("dir")) {
        # handle dir
        result += "dir = #{line}";
    } else {
        # handle file?
        result += "file = #{line}";
    }
}


for ([i, line] in lines) {
    #println("line = #{line}");
    var current_dir = '';
    if (line.startswith("$")) {
        # Handle cmd
        result += "cmd = #{line}";
    } else if (line.startswith("dir")) {
        # handle dir
        result += "dir = #{line}";
    }
}

println(result);

assert(result == expected);