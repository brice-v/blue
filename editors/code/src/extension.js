"use strict";

const vscode = require("vscode");
const cp = require("child_process");
const fs = require("fs");
const path = require("path");
const { LanguageClient } = require("vscode-languageclient/node");

let client;
let child;

function setting(key) {
  return String(vscode.workspace.getConfiguration("blue").get(key) || "").trim();
}

// Every `blue` on PATH, in order. Scanning all of them matters because a stale
// install earlier on PATH shadows a working build: an old binary does not know
// the `lsp` subcommand and dies with "file not found: lsp".
function pathCandidates() {
  const name = process.platform === "win32" ? "blue.exe" : "blue";
  const seen = new Set();
  const out = [];
  for (const dir of String(process.env.PATH || "").split(path.delimiter)) {
    if (!dir) continue;
    const full = path.join(dir, name);
    if (!seen.has(full)) {
      seen.add(full);
      out.push(full);
    }
  }
  return out;
}

// `blue help` exits 0 and lists the subcommands, so this tells a build that can
// serve LSP apart from one that would treat `lsp` as a file name. Only used for
// PATH discovery; an explicit blue.path is trusted as given.
function supportsLsp(command) {
  try {
    const out = cp.execFileSync(command, ["help"], { stdio: ["ignore", "pipe", "ignore"] }).toString();
    return /^\s*lsp\s/m.test(out);
  } catch {
    return false;
  }
}

function resolveBinary() {
  const explicit = setting("path");
  if (explicit) return explicit;

  for (const candidate of pathCandidates()) {
    if (!fs.existsSync(candidate)) continue;
    if (supportsLsp(candidate)) return candidate;
  }
  return undefined;
}

async function start() {
  const command = resolveBinary();

  if (!command) {
    const pick = "Set blue.path";
    vscode.window
      .showErrorMessage(
        "blue not found on PATH with `lsp` support. Build it (`go build -o blue .`) and set blue.path to the binary.",
        pick
      )
      .then((choice) => {
        if (choice === pick) {
          vscode.commands.executeCommand("workbench.action.openSettings", "blue.path");
        }
      });
    return;
  }

  const args = ["lsp"];
  if (setting("trace.server") === "verbose") {
    args.push("--trace");
  }

  // Function form must resolve to a ChildProcessInfo, MessageTransports or
  // StreamInfo. Spawning here keeps the arguments exactly `lsp`: an Executable
  // with a transport would append flags such as `--stdio` that blue rejects.
  const serverOptions = async () => {
    child = cp.spawn(command, args, { env: Object.assign({}, process.env), shell: false });
    child.on("error", (err) => {
      vscode.window.showErrorMessage(`failed to start blue language server: ${err.message}`);
    });
    return { process: child, detached: false };
  };

  const clientOptions = {
    documentSelector: [
      { scheme: "file", language: "blue" },
      { scheme: "untitled", language: "blue" },
    ],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher("**/*.b"),
    },
    outputChannelName: "Blue Language Server",
  };

  client = new LanguageClient("blue", serverOptions, clientOptions);
  try {
    await client.start();
  } catch (err) {
    vscode.window.showErrorMessage(`blue language server failed to start: ${err}`);
    client = undefined;
  }
}

async function stop() {
  if (client) {
    const running = client;
    client = undefined;
    try {
      await running.stop();
    } catch {
      // A connection that is already gone needs no report.
    }
  }
  if (child && child.exitCode === null) {
    child.kill();
    child = undefined;
  }
}

async function restart() {
  await stop();
  await start();
}

module.exports = {
  activate(context) {
    context.subscriptions.push(
      vscode.commands.registerCommand("blue.restart", () => restart()),
      vscode.workspace.onDidChangeConfiguration((event) => {
        if (event.affectsConfiguration("blue.path")) {
          restart();
        }
      })
    );
    return start();
  },

  deactivate() {
    return stop();
  },
};
