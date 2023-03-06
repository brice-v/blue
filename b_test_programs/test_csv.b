import csv

val fdata = read("example.csv");

var rows = csv.parse(fdata);
for (row in rows) {
    println("row = #{row}")
}
var expected = [['id', 'name', 'salary', 'department'],
['1', 'john', '2000', 'sales'],
['2', 'Andrew', '5000', 'finance'],
['3', 'Mark', '8000', 'hr'],
['4', 'Rey', '5000', 'marketing'],
['5', 'Tan', '4000', 'IT']];

assert(rows == expected);

var dumped = csv.dump(rows);
println("csv.dump(rows) = #{dumped}");
expected = """id,name,salary,department
1,john,2000,sales
2,Andrew,5000,finance
3,Mark,8000,hr
4,Rey,5000,marketing
5,Tan,4000,IT
""".replace("\r", "");
assert(dumped == expected);

rows = csv.parse(fdata, named_fields=true);
for (row in rows) {
    println("named| row = #{row}")
}

expected = [{id: '1', name: 'john', salary: '2000', department: 'sales'},
{id: '2', name: 'Andrew', salary: '5000', department: 'finance'},
{id: '3', name: 'Mark', salary: '8000', department: 'hr'},
{id: '4', name: 'Rey', salary: '5000', department: 'marketing'},
{id: '5', name: 'Tan', salary: '4000', department: 'IT'}];
assert(rows == expected);

dumped = csv.dump(rows);
println("csv.dump(rows) = #{csv.dump(rows)}")
expected = """id,name,salary,department
1,john,2000,sales
2,Andrew,5000,finance
3,Mark,8000,hr
4,Rey,5000,marketing
5,Tan,4000,IT
""".replace("\r", "");
assert(dumped == expected);