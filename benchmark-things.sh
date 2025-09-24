#!/bin/bash
hyperfine -n blue-baseline "./old-blue manual_tests/fib-ex.b" -n blue-new "./new-blue manual_tests/fib-ex.b" --runs 10 --export-markdown  hf.md && cat hf.md