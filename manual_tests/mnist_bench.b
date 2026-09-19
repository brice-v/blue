## CPU vs GPU speed by batch size for the MNIST MLP in blue.
##
##   go run ./manual_tests/mnist_prepare -train 60000 -test 10000
##   ./blue manual_tests/mnist_bench.b
##
## Each measurement trains the same model for a few epochs on a batch of the
## given size (a prefix of the training set). Data loading, slicing, the device
## transfer, and one warmup step are excluded from the timing, so this compares
## compute per epoch. The GPU column is skipped when no WebGPU adapter exists.
##
## The GPU backend requests the adapter's limits, so large single batches work
## when the device supports them; a 60000 x 784 float32 input is 188 MiB, which
## needs a maxStorageBufferBindingSize above the WebGPU default of 128 MiB.

import ml
import time

val dir = "manual_tests/mnist_data/";
val epochs = 3;

fun make_model(dev) {
    return ml.nn.Sequential([
        ml.nn.Linear(784, 128),
        ml.nn.ReLU(),
        ml.nn.Linear(128, 10),
    ]).to(dev);
}

fun seconds_per_epoch(dev, batch, full_x, full_y) {
    # slicing and the device transfer are setup, so keep them off the tape
    val x = ml.no_grad(fun() { return ml.slice(full_x, 0, 0, batch).to(dev); });
    val y = ml.no_grad(fun() { return ml.slice(full_y, 0, 0, batch).to(dev); });

    ml.manual_seed(0);
    val model = make_model(dev);
    val opt = ml.optim.Adam(ml.nn.parameters(model), lr=0.001);

    # warmup, not timed
    opt.zero_grad();
    val warm = ml.cross_entropy(model.forward(x), y);
    warm.backward();
    opt.step();

    val start_ms = time.now();
    for (e in 1..epochs) {
        opt.zero_grad();
        val loss = ml.cross_entropy(model.forward(x), y);
        loss.backward();
        opt.step();
    }
    return (time.now() - start_ms) / (1000.0 * epochs);
}

val full_x = ml.load(dir + "train_images.bin");
val full_y = ml.load(dir + "train_labels.bin");
val gpu_ok = ml.gpu.is_available();

print("epochs per measurement: #{epochs}, gpu available: #{gpu_ok}");
print("batch | cpu s/epoch | gpu s/epoch | speedup");

for (b in [1000, 5000, 10000, 20000, 40000, 60000]) {
    val cpu = seconds_per_epoch("cpu", b, full_x, full_y);
    var gpu = 0.0;
    var speed = 0.0;
    if (gpu_ok) {
        gpu = seconds_per_epoch("gpu", b, full_x, full_y);
        if (gpu > 0.0) {
            speed = cpu / gpu;
        }
    }
    print("#{b} | #{cpu} | #{gpu} | #{speed}");
}
