import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { createAssetRegistry, replaceAssetDirectory, writeTextFileAtomically } from "./assets.mjs";

async function withTemporaryPublicDirectory(run) {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "patch-assets-test-"));
  try {
    await run(root);
  } finally {
    await fs.rm(root, { recursive: true, force: true });
  }
}

test("replaceAssetDirectory preserves known-good assets after a download failure", async () => {
  await withTemporaryPublicDirectory(async (webPublicDir) => {
    const targetDir = path.join(webPublicDir, "assets", "patches", "test");
    await fs.mkdir(targetDir, { recursive: true });
    await fs.writeFile(path.join(targetDir, "old.png"), "old");
    const registry = createAssetRegistry();
    registry.register("https://assets.deadlock-api.com/new.png", "/assets/patches/test/new.png");
    const failingFetch = async () => new Response("unavailable", { status: 503 });

    await assert.rejects(replaceAssetDirectory(registry, webPublicDir, targetDir, failingFetch), /503/);
    assert.equal(await fs.readFile(path.join(targetDir, "old.png"), "utf8"), "old");
    await assert.rejects(fs.access(path.join(targetDir, "new.png")));
  });
});

test("replaceAssetDirectory installs a complete staged asset set", async () => {
  await withTemporaryPublicDirectory(async (webPublicDir) => {
    const targetDir = path.join(webPublicDir, "assets", "patches", "test");
    await fs.mkdir(targetDir, { recursive: true });
    await fs.writeFile(path.join(targetDir, "stale.png"), "stale");
    const registry = createAssetRegistry();
    registry.register("https://assets.deadlock-api.com/new.png", "/assets/patches/test/new.png");
    const png = Buffer.from("89504e470d0a1a0a", "hex");
    const successfulFetch = async () => new Response(png, { status: 200, headers: { "content-type": "image/png" } });

    await replaceAssetDirectory(registry, webPublicDir, targetDir, successfulFetch);
    assert.deepEqual(await fs.readFile(path.join(targetDir, "new.png")), png);
    await assert.rejects(fs.access(path.join(targetDir, "stale.png")));
  });
});

test("writeTextFileAtomically replaces an existing file", async () => {
  await withTemporaryPublicDirectory(async (root) => {
    const outputPath = path.join(root, "manifest.json");
    await fs.writeFile(outputPath, "old");

    await writeTextFileAtomically(outputPath, "new");

    assert.equal(await fs.readFile(outputPath, "utf8"), "new");
  });
});
