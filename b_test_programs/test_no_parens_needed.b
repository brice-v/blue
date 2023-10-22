var x = true;

if x {
    println("ITS TRUE")
}

if (x) {
    println("STILL TRUE")
}

for i in 1..10 {
    println("i = #{i}");
}


for (i in 1..10) {
    println("with parens, i = #{i}");
}


var z = [i for (i in 1..4) if (i % 2 == 0)];
println("z = #{z}");

var xy = [i for i in 1..4 if (i % 2 == 0)];
println("xy = #{xy}");
var xyz = [i for (i in 1..4) if i % 2 == 0];
println("xyz = #{xyz}");
var zz = [i for i in 1..4 if i % 2 == 0];
println("zz = #{zz}");

var zzz = [i for i in 1..4 if i % 2 == 0];
println("zzz = #{zzz}");



var z1 = {i for (i in 1..4) if (i % 2 == 0)};
println("z1 = #{z1}");

var xy1 = {i for i in 1..4 if (i % 2 == 0)};
println("xy1 = #{xy1}");
var xyz1 = {i for (i in 1..4) if i % 2 == 0};
println("xyz1 = #{xyz1}");
var zz1 = {i for i in 1..4 if i % 2 == 0};
println("zz1 = #{zz1}");


var z2 = {i: 'abc' for (i in 1..4) if (i % 2 == 0)};
println("z2 = #{z2}");

var xy2 = {i: 'abc' for i in 1..4 if (i % 2 == 0)};
println("xy2 = #{xy2}");
var xyz2 = {i: 'abc' for (i in 1..4) if i % 2 == 0};
println("xyz2 = #{xyz2}");
var zz2 = {i: 'abc' for i in 1..4 if i % 2 == 0};
println("zz2 = #{zz2}");

assert(true);