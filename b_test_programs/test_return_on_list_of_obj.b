var students = [
    { id: 14, name: "Kyle" },
    { id: 73, name: "Suzy" },
    { id: 112, name: "Frank" },
    { id: 6, name: "Sarah" }
];

fun getStudentName(studentID) {
    for (student in students) {
        println("student=#{student}");
        println("studentID=#{student.id == studentID}");
        if (student.id == studentID) {
            return student.name;
        }
    }
}

var nextStudent = getStudentName(73);

if (nextStudent != "Suzy") {
    return false;
}

println(nextStudent);
# This is failing - but also the tests are still passing, need to figure out what exactly is happening
# Had to do a lot of other fixes but this is now working, we check if theres a return value in the for expression
println(nextStudent == "Suzy");
println(nextStudent != "Suzy");

return true;