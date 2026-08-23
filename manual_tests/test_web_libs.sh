#!/bin/bash
# Manual test for the vendored browser libraries in lib/web/.
#
# twind.js, preact.js and htmx.js are minified CDN builds that get embedded
# into the blue binary (see lib/lib.go) and served by
# http.serve(use_embedded_lib_web=true). They should never be hand-edited,
# only re-vendored from their upstream CDNs.
#
# This test checks that the files are present, syntactically valid,
# self-contained, expose the APIs documented in lib/std/http.b, and still
# match the expected upstream versions. Bump the EXPECTED_* variables at
# the top whenever you re-vendor a library.
#
# Usage:
#     ./manual_tests/test_web_libs.sh
#
# Exits 0 when every check passes, 1 otherwise.

EXPECTED_HTMX_VERSION="2.0.10"
EXPECTED_TWIND_VERSION="1.0.8"
EXPECTED_HTM_VERSION="3.1.1"
EXPECTED_PREACT_VERSION="10.29.8"

set -u
cd "$(dirname "$0")/.." || exit 1

WEB="lib/web"
PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); echo "PASS: $1"; }
fail() {
    FAIL=$((FAIL + 1))
    echo "FAIL: $1"
    [ $# -gt 1 ] && printf '      detail: %s\n' "$2"
}

# check_contains NAME SUBSTRING FILE
check_contains() {
    case "$(cat "$3" 2>/dev/null)" in
        *"$2"*) pass "$1" ;;
        *) fail "$1" "[$3] missing [$2]" ;;
    esac
}

# check_lacks NAME SUBSTRING FILE
check_lacks() {
    case "$(cat "$3" 2>/dev/null)" in
        *"$2"*) fail "$1" "[$3] should not contain [$2]" ;;
        *) pass "$1" ;;
    esac
}

echo "== vendored files exist and are non-empty"
for f in twind.js preact.js htmx.js water.css water-dark.css water-light.css; do
    if [ -s "$WEB/$f" ]; then
        pass "$WEB/$f"
    else
        fail "$WEB/$f is missing or empty"
    fi
done

echo "== twind.js (@twind/cdn, tailwind-in-the-browser)"
check_contains "twind provenance header pins @twind/cdn@$EXPECTED_TWIND_VERSION" \
    "/npm/@twind/cdn@$EXPECTED_TWIND_VERSION/cdn.global.js" "$WEB/twind.js"
check_contains "twind exports the tailwind preset" "e.presetTailwind_colors=" "$WEB/twind.js"
check_contains "twind exports install as global 'twind'" "install=eO" "$WEB/twind.js"

echo "== preact.js (htm/preact/standalone ES module)"
check_contains "preact header documents htm@$EXPECTED_HTM_VERSION" \
    "htm@$EXPECTED_HTM_VERSION/preact/standalone" "$WEB/preact.js"
check_contains "preact header documents preact@$EXPECTED_PREACT_VERSION" \
    "preact@$EXPECTED_PREACT_VERSION" "$WEB/preact.js"
check_lacks "preact bundle is self-contained (no external esm.sh imports)" 'from"/' "$WEB/preact.js"
for export_name in Component createContext h html render useCallback useContext \
    useDebugValue useEffect useErrorBoundary useImperativeHandle useLayoutEffect \
    useMemo useReducer useRef useState; do
    # minified export list looks like: "A as Component,Sn as createContext,..."
    if grep -Eq " as $export_name(,|})" "$WEB/preact.js"; then
        pass "preact exports $export_name"
    else
        fail "preact exports $export_name" "[$WEB/preact.js] missing export [$export_name]"
    fi
done

echo "== htmx.js (htmx.org UMD build)"
check_contains "htmx version string is $EXPECTED_HTMX_VERSION" \
    "version:\"$EXPECTED_HTMX_VERSION\"" "$WEB/htmx.js"
check_contains "htmx exposes defineExtension" "defineExtension" "$WEB/htmx.js"
check_contains "htmx exposes onLoad" "onLoad:" "$WEB/htmx.js"
check_contains "htmx keeps default swap style innerHTML" \
    'defaultSwapStyle:"innerHTML"' "$WEB/htmx.js"
check_contains "htmx fires htmx:beforeSwap" "htmx:beforeSwap" "$WEB/htmx.js"

if ! command -v node >/dev/null 2>&1; then
    echo "== node not found: skipping syntax/runtime checks"
else
    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT

    echo "== syntax checks (node --check)"
    cp "$WEB/twind.js" "$TMP_DIR/twind.js"
    cp "$WEB/htmx.js" "$TMP_DIR/htmx.js"
    cp "$WEB/preact.js" "$TMP_DIR/preact.mjs"
    if node --check "$TMP_DIR/twind.js"; then
        pass "twind.js parses as a classic script"
    else
        fail "twind.js has a syntax error"
    fi
    if node --check "$TMP_DIR/htmx.js"; then
        pass "htmx.js parses as a classic script"
    else
        fail "htmx.js has a syntax error"
    fi
    if node --check "$TMP_DIR/preact.mjs"; then
        pass "preact.js parses as an ES module"
    else
        fail "preact.js has a syntax error"
    fi

    echo "== preact runtime smoke test (node)"
    cat > "$TMP_DIR/smoke.mjs" <<'EOF'
import * as p from "./preact.mjs";

const api = [
    "Component", "createContext", "h", "html", "render",
    "useCallback", "useContext", "useDebugValue", "useEffect",
    "useErrorBoundary", "useImperativeHandle", "useLayoutEffect",
    "useMemo", "useReducer", "useRef", "useState",
];
const bad = api.filter((n) => typeof p[n] !== "function");
if (bad.length > 0) {
    console.error("non-function exports:", bad.join(","));
    process.exit(1);
}

// html`` builds a vnode tree without needing a DOM (render needs one).
const tree = p.html`<div class="greeting">hello from blue</div>`;
if (!tree || (typeof tree !== "object" && typeof tree !== "function")) {
    console.error("html`` did not produce a vnode tree");
    process.exit(1);
}
console.log("all 16 preact standalone exports are functions, html`` works");
EOF
    if (cd "$TMP_DIR" && node smoke.mjs); then
        pass "preact standalone loads and runs in node"
    else
        fail "preact standalone smoke test failed"
    fi
fi

echo
echo "== $PASS passed, $FAIL failed"
if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
exit 0
