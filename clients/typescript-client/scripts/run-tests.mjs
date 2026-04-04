import { spawnSync } from 'node:child_process';
import { readdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageRoot = path.resolve(scriptDir, '..');
const targetDir = process.argv[2];

if (!targetDir) {
  console.error('Usage: node scripts/run-tests.mjs <test-directory>');
  process.exit(1);
}

function collectTestFiles(directoryPath) {
  const testFiles = [];

  for (const entry of readdirSync(directoryPath, { withFileTypes: true })) {
    const entryPath = path.join(directoryPath, entry.name);

    if (entry.isDirectory()) {
      testFiles.push(...collectTestFiles(entryPath));
      continue;
    }

    if (entry.isFile() && entry.name.endsWith('.test.ts')) {
      testFiles.push(entryPath);
    }
  }

  return testFiles;
}

const absoluteTargetDir = path.resolve(packageRoot, targetDir);
const testFiles = collectTestFiles(absoluteTargetDir).sort();

if (testFiles.length === 0) {
  console.error(`No test files found under ${targetDir}`);
  process.exit(1);
}

const tsxBinary = path.join(
  packageRoot,
  'node_modules',
  '.bin',
  process.platform === 'win32' ? 'tsx.cmd' : 'tsx'
);

const result = spawnSync(tsxBinary, ['--test', ...testFiles], {
  cwd: packageRoot,
  stdio: 'inherit',
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
