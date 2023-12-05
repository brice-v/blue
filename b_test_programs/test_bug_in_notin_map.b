var a = {1: 2};

var card_cache = {};

var card = {'card': 1};


if card.card in card_cache {
    println("in cache")
} else {
    card_cache[card.card] = a;
}

println("card_cache = #{card_cache}")

if card.card in card_cache {
    println("Now in cache?")
}
assert(card.card in card_cache);