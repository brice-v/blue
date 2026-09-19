# PyTorch equivalent of manual_tests/mnist_mlp.b.
#
# Same subset (6000 train / 1000 test), same model, same hyperparameters, so the
# two files can be compared line for line.
#
#   python manual_tests/mnist_mlp.py
#
# The blue version reads the .bin tensors written by manual_tests/mnist_prepare;
# here we read the raw IDX files directly with a tiny numpy reader.

import numpy as np
import torch
import torch.nn as nn

DATA_DIR = "borncgo/examples/mnist/data"


def read_idx(path):
    raw = open(path, "rb").read()
    ndim = raw[3]
    dims = [int.from_bytes(raw[4 + 4 * i:8 + 4 * i], "big") for i in range(ndim)]
    return np.frombuffer(raw[4 + 4 * ndim:], dtype=np.uint8).reshape(dims)


def load_split(split, limit):
    images = read_idx(f"{DATA_DIR}/{split}-images-idx3-ubyte")
    labels = read_idx(f"{DATA_DIR}/{split}-labels-idx1-ubyte")
    n = min(limit, len(labels))
    x = torch.tensor(images[:n].reshape(n, -1), dtype=torch.float32) / 255.0
    y = torch.tensor(labels[:n].astype(np.int64), dtype=torch.long)
    return x, y


def main():
    train_x, train_y = load_split("train", 6000)
    test_x, test_y = load_split("t10k", 1000)

    dev = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    train_x, train_y = train_x.to(dev), train_y.to(dev)
    test_x, test_y = test_x.to(dev), test_y.to(dev)

    print("device", dev)
    print("train", tuple(train_x.shape), "test", tuple(test_x.shape))

    torch.manual_seed(0)

    model = nn.Sequential(
        nn.Linear(784, 128),
        nn.ReLU(),
        nn.Linear(128, 10),
    ).to(dev)

    opt = torch.optim.Adam(model.parameters(), lr=0.001)
    loss_fn = nn.CrossEntropyLoss()

    def accuracy(xs, ys):
        with torch.no_grad():
            logits = model(xs)
            preds = logits.argmax(dim=1)
            return (preds == ys).sum().item() / ys.shape[0]

    for epoch in range(1, 12):
        opt.zero_grad()
        logits = model(train_x)
        loss = loss_fn(logits, train_y)
        loss.backward()
        opt.step()
        print("epoch", epoch, "loss", loss.item(), "test_acc", accuracy(test_x, test_y))

    print("final test accuracy:", accuracy(test_x, test_y))


if __name__ == "__main__":
    main()
