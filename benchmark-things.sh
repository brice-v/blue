#!/bin/bash
hyperfine -n blue-v0.2.3 "./blue-v0.2.3 manual_tests/fib-ex.b" \
 -n blue-v0.3.0 "./blue manual_tests/fib-ex.b --vm" \
 --runs 10 --export-markdown  hf.md && cat hf.md