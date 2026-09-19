## Manual MNIST MLP test.
##
## The MNIST loader lives in manual_tests/mnist_prepare (Go), not in the blue
## source. Prepare the tensors once, then run this script:
##
##   go run ./manual_tests/mnist_prepare
##   ./blue manual_tests/mnist_mlp.b
##
## The PyTorch equivalent is manual_tests/mnist_mlp.py. Device handling mirrors
## PyTorch: build on CPU, then .to(device), and pick the device with
## ml.gpu.is_available().

import ml

val dir = "manual_tests/mnist_data/";

var dev = "cpu";
if (ml.gpu.is_available()) {
    dev = "gpu";
}

val train_x = ml.load(dir + "train_images.bin").to(dev);
val train_y = ml.load(dir + "train_labels.bin").to(dev);
val test_x = ml.load(dir + "test_images.bin").to(dev);
val test_y = ml.load(dir + "test_labels.bin").to(dev);

print("device", dev);
print("train", train_x.shape, "test", test_x.shape);

ml.manual_seed(0);

val model = ml.nn.Sequential([
    ml.nn.Linear(784, 128),
    ml.nn.ReLU(),
    ml.nn.Linear(128, 10),
]).to(dev);

val params = ml.nn.parameters(model);
val opt = ml.optim.Adam(params, lr=0.001);

fun accuracy(xs, ys) {
    val logits = model.forward(xs);
    val preds = ml.argmax(logits, 1);
    val correct = ml.sum(ml.eq(preds, ys)).item();
    return correct / ys.shape[0];
}

for (epoch in 1..11) {
    opt.zero_grad();
    val logits = model.forward(train_x);
    val loss = ml.cross_entropy(logits, train_y);
    loss.backward();
    opt.step();
    print("epoch", epoch, "loss", loss.item(), "test_acc", accuracy(test_x, test_y));
}

print("final test accuracy:", accuracy(test_x, test_y));
