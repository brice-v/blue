var x = [1,2,3];
var y = x.push!('4');
assert(x == [1,2,3,'4']);
assert(y == 4);
var z = x.pop!();
assert(x == [1,2,3]);
assert(z == '4');

x = [1,2,3];
y = x.unshift!('Hello', 'World');
assert(x == ['Hello','World',1,2,3]);
assert(y == 5);
z = x.shift!();
assert(x == ['World',1,2,3]);
assert(z == 'Hello');

x = [1,2,3];
y = ['a','b','c'];
z = x.concat(y, [0.1,0.2,0.3]);
assert(x == [1,2,3]);
assert(y == ['a','b','c']);
assert(z == [1,2,3,'a','b','c',0.1,0.2,0.3]);

x = [1,2,3]
x = x.append(4,5,6,7);
assert(x == [1,2,3,4,5,6,7]);

x = [1,2,3];
x = x.prepend(4,5,6,7);
assert(x == [4,5,6,7,1,2,3]);