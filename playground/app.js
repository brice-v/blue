/* blue playground
 *
 * Everything runs client side. The blue interpreter is compiled to
 * WebAssembly (see ../wasmmain and make_wasm.sh) and talks to this page via
 * exactly two functions: blueVersion() and blueRun(source). User code is
 * interpreted inside the wasm sandbox and cannot reach the page, storage,
 * the filesystem or processes.
 */
(function () {
	'use strict';

	// ------------------------------------------------------------------
	// Output capture.
	//
	// go_js_wasm_exec routes every write to fds 1/2 through fs.writeSync.
	// We wrap that single choke point to collect program output instead of
	// spamming the devtools console.
	// ------------------------------------------------------------------
	const decoder = new TextDecoder('utf-8');
	let capturedStdout = '';
	let capturedStderr = '';

	if (!globalThis.fs || typeof globalThis.fs.writeSync !== 'function') {
		throw new Error('wasm_exec.js did not install its fs stub; update it or app.js');
	}
	const stubFs = globalThis.fs;
	const origWriteSync = stubFs.writeSync;
	stubFs.writeSync = function (fd, buf) {
		if (fd === 1 || fd === 2) {
			const text = decoder.decode(buf);
			if (fd === 1) {
				capturedStdout += text;
			} else {
				capturedStderr += text;
			}
			return buf.length;
		}
		return origWriteSync.call(this, fd, buf);
	};

	// ------------------------------------------------------------------
	// Default example program.
	// ------------------------------------------------------------------
	const DEFAULT_PROGRAM = `# welcome to the blue playground!
# edit the code and hit Run (or press Ctrl/Cmd+Enter)

fun fib(n) {
    if n < 2 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

for i in 1..10 {
    println("fib(#{i}) = #{fib(i)}")
}

# std modules work too:
import math
println("sqrt(144) = #{math.sqrt(144)}")
`;

	// ------------------------------------------------------------------
	// Wasm bootstrapping.
	// ------------------------------------------------------------------
	async function loadRuntime() {
		const go = new Go();
		go.argv = ['blue'];
		let result;
		if (typeof WebAssembly.instantiateStreaming === 'function') {
			try {
				result = await WebAssembly.instantiateStreaming(fetch('blue.wasm'), go.importObject);
			} catch (err) {
				// Fall back (eg. server sends a wrong content-type).
				result = await WebAssembly.instantiate(await fetch('blue.wasm').then((r) => r.arrayBuffer()), go.importObject);
			}
		} else {
			result = await WebAssembly.instantiate(await fetch('blue.wasm').then((r) => r.arrayBuffer()), go.importObject);
		}
		// The go program parks itself forever, keeping blueRun/blueVersion
		// registered; do not await completion.
		go.run(result.instance);
	}

	// ------------------------------------------------------------------
	// Syntax highlighting.
	//
	// Deliberately basic: a single pass tokenizer over the source that
	// colors comments, strings, numbers and keywords (keyword list mirrors
	// token/token.go). The output is written into a <pre> sitting behind
	// the transparent text textarea.
	// ------------------------------------------------------------------
	const KEYWORDS = [
		'fun', 'var', 'val', 'true', 'false', 'if', 'else', 'return', 'for',
		'in', 'notin', 'and', 'or', 'not', 'const', 'match', 'null', 'import',
		'from', 'as', 'try', 'catch', 'finally', 'eval', 'spawn', 'defer',
		'self', 'break', 'continue',
	];

	const HL_RE = new RegExp(
		[
			// 1: comments. ### multiline ###, ## docstring, # line.
			'(###(?:[\\s\\S]*?)###|##[^\\n]*|#[^\\n]*)',
			// 2: strings, single or double quoted with backslash escapes.
			'|("(?:[^"\\\\\\n]|\\\\.)*"|\'(?:[^\'\\\\\\n]|\\\\.)*\')',
			// 3: numbers, hex/octal/binary/floats with underscores.
			'|(\\b0[xXoObB][0-9a-fA-F_]+|\\b\\d[\\d_]*(?:\\.\\d+)?(?:[eE][+-]?\\d+)?)',
			// 4: keywords.
			'|\\b(' + KEYWORDS.join('|') + ')\\b',
		].join(''),
		'g'
	);

	function escapeHtml(s) {
		return s
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;');
	}

	function highlightBlue(src) {
		let out = '';
		let last = 0;
		HL_RE.lastIndex = 0;
		let m;
		while ((m = HL_RE.exec(src)) !== null) {
			if (m.index > last) {
				out += escapeHtml(src.slice(last, m.index));
			}
			const cls = m[1] ? 'tok-comment' : m[2] ? 'tok-string' : m[3] ? 'tok-number' : 'tok-keyword';
			out += `<span class="${cls}">${escapeHtml(m[0])}</span>`;
			last = m.index + m[0].length;
			if (m[0].length === 0) {
				HL_RE.lastIndex++; // guard against zero length matches
			}
		}
		out += escapeHtml(src.slice(last));
		return out;
	}

	const highlightEl = document.getElementById('highlight');
	const highlightCodeEl = document.getElementById('highlight-code');

	// renderHighlight repaints the backdrop and keeps it scroll aligned
	// with the textarea. A trailing newline keeps the layer heights equal
	// when the program ends with one.
	function renderHighlight() {
		highlightCodeEl.innerHTML = highlightBlue(editorEl.value) + '\n';
		highlightEl.scrollTop = editorEl.scrollTop;
		highlightEl.scrollLeft = editorEl.scrollLeft;
	}

	// ------------------------------------------------------------------
	// UI wiring.
	// ------------------------------------------------------------------
	const editorEl = document.getElementById('editor');
	const runBtn = document.getElementById('run-btn');
	const versionEl = document.getElementById('version');
	const stdoutEl = document.getElementById('stdout');
	const resultEl = document.getElementById('result');
	const stderrEl = document.getElementById('stderr');
	const errorTitleEl = document.querySelector('.error-title');

	editorEl.value = DEFAULT_PROGRAM;

	function render(res) {
		stdoutEl.textContent = res.output;
		resultEl.textContent = res.result ? `=> ${res.result}` : '';
		stderrEl.textContent = res.error.trim();
		errorTitleEl.classList.toggle('muted', res.error.trim() === '');
	}

	function run() {
		// Reset panes first so stale output never lingers under a crash.
		render({ output: '', error: '', result: '' });
		capturedStdout = '';
		capturedStderr = '';

		let res;
		try {
			res = globalThis.blueRun(editorEl.value);
		} catch (err) {
			res = { error: `runtime call failed: ${err}\n`, output: '', result: '' };
		}
		if (!res) {
			// The wasm module died (a go panic that escaped recovery).
			res = { error: 'the runtime crashed while running this program, please reload the page\n', output: capturedStdout, result: '' };
		}
		// Attach whatever the program printed while it ran.
		res.output = capturedStdout;
		res.error = (res.error || '') + capturedStderr;

		render(res);
	}

	runBtn.addEventListener('click', run);

	document.addEventListener('keydown', (ev) => {
		if ((ev.ctrlKey || ev.metaKey) && ev.key === 'Enter') {
			ev.preventDefault();
			run();
		}
	});

	// Make the Tab key insert spaces instead of moving focus.
	editorEl.addEventListener('keydown', (ev) => {
		if (ev.key !== 'Tab') {
			return;
		}
		ev.preventDefault();
		const start = editorEl.selectionStart;
		const end = editorEl.selectionEnd;
		editorEl.setRangeText('    ', start, end, 'end');
		renderHighlight();
	});

	// Keep the highlighted backdrop in sync with the textarea.
	editorEl.addEventListener('input', renderHighlight);
	editorEl.addEventListener('scroll', () => {
		highlightEl.scrollTop = editorEl.scrollTop;
		highlightEl.scrollLeft = editorEl.scrollLeft;
	});

	// Persist scratch code across reloads (best effort, sandboxed origin storage).
	try {
		const saved = localStorage.getItem('playground-source');
		if (saved !== null) {
			editorEl.value = saved;
		}
		editorEl.addEventListener('input', () => {
			try {
				localStorage.setItem('playground-source', editorEl.value);
			} catch (_err) {
				/* storage may be unavailable; not fatal */
			}
		});
	} catch (_err) {
		/* ignore */
	}

	renderHighlight();

	// ------------------------------------------------------------------
	// Boot.
	// ------------------------------------------------------------------
	loadRuntime()
		.then(() => {
			versionEl.textContent = globalThis.blueVersion();
			runBtn.disabled = false;
			run();
		})
		.catch((err) => {
			versionEl.textContent = 'failed to load blue.wasm';
			stderrEl.textContent = String(err);
			errorTitleEl.classList.remove('muted');
		});
})();
