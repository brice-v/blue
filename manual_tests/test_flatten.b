fun flatten(l) {
    return [item for sublist in l for item in sublist];
}


# TODO: 
println(flatten([[1,2,3],[4,5,6]]));