## Phase 2 integration test (see ML_PLAN.md).
##
## Target: object.Tensor plus `_tensor`, `_matmul`, `_add`, property access, and
## the `lib/std/ml.b` exports, so that `a.matmul(b)` works from blue.
##
## This file is intentionally red until Phase 2 lands. If you need the
## b_test_programs suite green while implementing, add a leading `# IGNORE`.
##
## Value comparisons below use tensor `==`, which currently falls back to
## HashObject value-equality. When Phase 10 routes `==` to the elementwise
## `_eq`, switch these to a value reader such as `ml.to_list(...)`.

import ml

# creation from nested lists infers the shape
val a = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);      # 2x3
val b = ml.tensor([[7.0, 8.0], [9.0, 10.0], [11.0, 12.0]]); # 3x2

assert(type(a) == "TENSOR");

# property access
assert(a.shape == [2, 3]);
assert(a.ndim == 2);
assert(a.dtype == "float32");
assert(a.device == "cpu");
assert(a.strides == [3, 1]);
assert(a.requires_grad == false);

# transpose is a metadata-only property
assert(a.T.shape == [3, 2]);
assert(a.T.strides == [1, 3]);

# matmul, method style
val c = a.matmul(b);
assert(c.shape == [2, 2]);
assert(c == ml.tensor([[58.0, 64.0], [139.0, 154.0]]));

# matmul, module style; both call styles must agree
val d = ml.matmul(a, b);
assert(d.shape == [2, 2]);
assert(c == d);

# add, tensor-tensor
val e = ml.add(c, c);
assert(e == ml.tensor([[116.0, 128.0], [278.0, 308.0]]));

# add, scalar broadcast (float and int scalars both coerce)
val f = ml.add(c, 1.0);
assert(f == ml.tensor([[59.0, 65.0], [140.0, 155.0]]));
assert(ml.add(c, 1) == f);

assert(true);

##############################################################################
## Below: what the same test looks like once the operators land (Phase 3/10).
##
## Kept commented so test_ml_phase2.b stays a clean Phase 2 target. Uncomment
## this block (or copy it into its own file) once the operators exist.
##
## Operator surface from ML_PLAN.md Phase 10:
##   + - * / **   elementwise (note: `*` stays elementwise, matching PyTorch)
##   @            matmul (new operator, Phase 10.2)
##   unary -      negation
##   == != > < >= <=  comparisons, which return bool TENSORS
##   += -= *= /=  compound assignment, desugars to the binary op
##
## import ml
##
## val a = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
## val b = ml.tensor([[7.0, 8.0], [9.0, 10.0], [11.0, 12.0]]);
##
## val c = a @ b;              # matmul, same as a.matmul(b)
## assert(c.shape == [2, 2]);
##
## val e = c + c;             # add
## val f = c + 1.0;           # scalar broadcast add
## val g = c - 1.0;           # sub
## val h = c * 2.0;           # mul, elementwise
## val i = c / 2.0;           # div
## val j = c ** 2.0;          # pow
## val k = -c;                # unary neg
##
## var m = c;
## m += c;                    # compound assignment, desugars to the binary op
##
## # in-place forms are functions in this design, not operators:
## ml.add_(c, c);             # _add_inplace
## ml.mul_(c, 2.0);           # _mul_inplace
##
## # comparisons return bool TENSORS, not a Boolean, so read them with to_list
## # (or a reducer) instead of asserting the tensor directly:
## assert(ml.to_list(c == ml.matmul(a, b)) == [true, true, true, true]);
## assert(ml.to_list(c > 0.0) == [true, true, true, true]);
##
## # the method and module forms still work alongside the operators:
## assert(ml.to_list((a @ b) == a.matmul(b)) == [true, true, true, true]);
