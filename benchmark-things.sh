#!/bin/bash
hyperfine -n blue-baseline "./blue manual_tests/fib-ex.b" \
 -n blue-vm "./blue --vm manual_tests/fib-ex.b" \
 --runs 10 --export-markdown  hf.md && cat hf.md