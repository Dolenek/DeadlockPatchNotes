#!/usr/bin/env node
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const MAX_FILE_LINES = 500;
const WARN_FILE_LINES = 400;
const MAX_FUNCTION_LINES = 40;

const INCLUDED_EXTENSIONS = new Set([".go", ".ts", ".tsx", ".js", ".mjs", ".css", ".sql", ".md"]);
const SOURCE_ROOTS = ["api", "web", "scripts", "docs"];
const EXCLUDED_PREFIXES = [
  "web/node_modules/",
  "web/.next/",
  "web/public/assets/",
  "api/internal/patches/data/",
  "web/package-lock.json",
];

const FUNCTION_START_PATTERNS = [
  /^\s*(export\s+)?(async\s+)?function\b/,
  /^\s*function\b/,
  /^\s*const\s+[A-Za-z0-9_]+\s*=\s*(async\s*)?\([^)]*\)\s*=>/,
  /^\s*func\b/,
];

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, "..");

function shouldExclude(relPath) {
  return EXCLUDED_PREFIXES.some((prefix) => relPath === prefix || relPath.startsWith(prefix));
}

function isSourceFile(relPath) {
  if (shouldExclude(relPath)) {
    return false;
  }
  return INCLUDED_EXTENSIONS.has(path.extname(relPath));
}

async function walk(currentDir, found) {
  const entries = await fs.readdir(currentDir, { withFileTypes: true });
  for (const entry of entries) {
    const absPath = path.join(currentDir, entry.name);
    const relPath = path.relative(ROOT, absPath).replace(/\\/g, "/");

    if (shouldExclude(relPath)) {
      continue;
    }

    if (entry.isDirectory()) {
      await walk(absPath, found);
      continue;
    }

    if (isSourceFile(relPath)) {
      found.push({ absPath, relPath });
    }
  }
}

async function gatherSourceFiles() {
  const files = [];
  for (const root of SOURCE_ROOTS) {
    await walk(path.join(ROOT, root), files);
  }
  return files;
}

function collectLongFunctions(content, relPath) {
  const lines = content.split(/\r?\n/);
  const starts = [];

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    if (FUNCTION_START_PATTERNS.some((pattern) => pattern.test(line))) {
      starts.push(index + 1);
    }
  }

  const findings = [];
  for (let i = 0; i < starts.length; i += 1) {
    const start = starts[i];
    const end = i < starts.length - 1 ? starts[i + 1] - 1 : lines.length;
    const length = end - start + 1;
    if (length > MAX_FUNCTION_LINES) {
      findings.push(`${relPath}:${start} function length ${length} > ${MAX_FUNCTION_LINES}`);
    }
  }

  return findings;
}

async function analyzeFiles(files) {
  const warnings = [];
  const errors = [];

  for (const file of files) {
    const content = await fs.readFile(file.absPath, "utf8");
    const lineCount = content.split(/\r?\n/).length;

    if (lineCount > MAX_FILE_LINES) {
      errors.push(`${file.relPath} has ${lineCount} lines (max ${MAX_FILE_LINES})`);
    } else if (lineCount > WARN_FILE_LINES) {
      warnings.push(`${file.relPath} has ${lineCount} lines (warn threshold ${WARN_FILE_LINES})`);
    }

    errors.push(...collectLongFunctions(content, file.relPath));
  }

  return { warnings, errors };
}

function printReport(warnings, errors) {
  if (warnings.length > 0) {
    console.log("Warnings:");
    for (const warning of warnings) {
      console.log(`- ${warning}`);
    }
  }

  if (errors.length > 0) {
    console.log("Violations:");
    for (const error of errors) {
      console.log(`- ${error}`);
    }
    process.exitCode = 1;
    return;
  }

  console.log("Source limits check passed.");
}

async function main() {
  const files = await gatherSourceFiles();
  const { warnings, errors } = await analyzeFiles(files);
  printReport(warnings, errors);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
