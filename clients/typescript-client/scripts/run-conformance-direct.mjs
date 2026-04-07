import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { spawn, spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageRoot = path.resolve(scriptDir, '..');
const repoRoot = path.resolve(packageRoot, '..', '..');

const defaultOutputDir = path.join(repoRoot, 'attached_assets', 'typescript-conformance');
const defaultTestFile = path.join('tests', 'conformance', 'conformance.test.ts');
const npmBinary = process.platform === 'win32' ? 'npm.cmd' : 'npm';
const tsxBinary = path.join(
  packageRoot,
  'node_modules',
  '.bin',
  process.platform === 'win32' ? 'tsx.cmd' : 'tsx',
);

function sanitizeLabel(value) {
  return value.replace(/[^a-zA-Z0-9._-]+/g, '-').replace(/^-+|-+$/g, '') || 'conformance';
}

function parseArgs(argv) {
  const options = {
    label: 'full-suite',
    pattern: '',
    reporter: 'tap',
    aggregate: false,
    skipBuild: false,
    outputDir: defaultOutputDir,
    showLines: 80,
    testFile: defaultTestFile,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];

    if (arg === '--label') {
      options.label = argv[index + 1] ?? options.label;
      index += 1;
      continue;
    }

    if (arg === '--pattern') {
      options.pattern = argv[index + 1] ?? options.pattern;
      index += 1;
      continue;
    }

    if (arg === '--reporter') {
      options.reporter = argv[index + 1] ?? options.reporter;
      index += 1;
      continue;
    }

    if (arg === '--aggregate') {
      options.aggregate = true;
      continue;
    }

    if (arg === '--output-dir') {
      options.outputDir = path.resolve(packageRoot, argv[index + 1] ?? options.outputDir);
      index += 1;
      continue;
    }

    if (arg === '--show-lines') {
      const parsed = Number.parseInt(argv[index + 1] ?? '', 10);
      if (Number.isFinite(parsed) && parsed > 0) {
        options.showLines = parsed;
      }
      index += 1;
      continue;
    }

    if (arg === '--test-file') {
      options.testFile = argv[index + 1] ?? options.testFile;
      index += 1;
      continue;
    }

    if (arg === '--skip-build') {
      options.skipBuild = true;
      continue;
    }

    if (arg === '--help') {
      console.log([
        'Usage: node scripts/run-conformance-direct.mjs [options]',
        '',
        'Options:',
        '  --label <name>        Artifact prefix under attached_assets/typescript-conformance',
        '  --pattern <regex>     node:test name pattern filter',
        '  --reporter <name>     node:test reporter (default: tap)',
        '  --aggregate           Build a synthetic full-suite summary from per-family summary files',
        '  --output-dir <path>   Override artifact directory',
        '  --show-lines <count>  Print the last N log lines (default: 80)',
        '  --test-file <path>    Override the conformance test file',
        '  --skip-build          Skip npm run build before test execution',
        '',
        'Without --pattern, the runner executes the full suite one case-family at a time',
        'and persists progress after every case slice.',
      ].join('\n'));
      process.exit(0);
    }
  }

  options.label = sanitizeLabel(options.label);
  return options;
}

function runCommand(command, args, cwd) {
  return spawnSync(command, args, {
    cwd,
    encoding: 'utf8',
    maxBuffer: 32 * 1024 * 1024,
  });
}

function parseTapSummary(output) {
  const lines = output.split(/\r?\n/);
  const subtests = [];

  for (const line of lines) {
    const match = /^(ok|not ok)\s+\d+\s+-\s+(.*)$/.exec(line.trim());
    if (!match) {
      continue;
    }

    subtests.push({
      status: match[1] === 'ok' ? 'passed' : 'failed',
      name: match[2],
    });
  }

  return {
    passed: subtests.filter((entry) => entry.status === 'passed').map((entry) => entry.name),
    failed: subtests.filter((entry) => entry.status === 'failed').map((entry) => entry.name),
    subtests,
  };
}

function loadCaseIds() {
  const caseDir = path.join(repoRoot, 'conformance', 'cases');
  return readdirSync(caseDir)
    .filter((fileName) => fileName.endsWith('.json'))
    .sort()
    .map((fileName) => {
      const caseDocument = JSON.parse(readFileSync(path.join(caseDir, fileName), 'utf8'));
      return String(caseDocument.id ?? '').trim();
    })
    .filter((caseId) => caseId.length > 0);
}

function formatCommand(command, args) {
  return [command, ...args].join(' ');
}

function isCompletedSummary(document) {
  if (document.complete === true) {
    return true;
  }

  return typeof document.exitCode === 'number';
}

function loadPatternSummaryEntries(outputDir, currentLabel) {
  if (!path.isAbsolute(outputDir)) {
    return [];
  }

  return readdirSync(outputDir)
    .filter((fileName) => fileName.endsWith('.summary.json'))
    .map((fileName) => {
      const filePath = path.join(outputDir, fileName);
      try {
        const document = JSON.parse(readFileSync(filePath, 'utf8'));
        return { fileName, filePath, document };
      } catch {
        return null;
      }
    })
    .filter((entry) => entry !== null)
    .filter((entry) => entry.document.label !== currentLabel)
    .filter((entry) => typeof entry.document.pattern === 'string' && entry.document.pattern.length > 0)
    .filter((entry) => entry.document.mode !== 'aggregate');
}

function aggregateSummaries(options, logPath, exitPath, summaryPath) {
  const expectedPatterns = loadCaseIds();
  const entries = loadPatternSummaryEntries(options.outputDir, options.label);
  const byPattern = new Map();

  for (const entry of entries) {
    const existing = byPattern.get(entry.document.pattern);
    if (!existing || (!isCompletedSummary(existing.document) && isCompletedSummary(entry.document))) {
      byPattern.set(entry.document.pattern, entry);
    }
  }

  const missingPatterns = [];
  const incompletePatterns = [];
  const failedPatterns = [];
  const passed = [];
  const failed = [];
  const subtests = [];
  const caseRuns = [];

  for (const pattern of expectedPatterns) {
    const entry = byPattern.get(pattern);
    if (!entry) {
      missingPatterns.push(pattern);
      continue;
    }

    const { document, filePath } = entry;
    const complete = isCompletedSummary(document);
    caseRuns.push({
      pattern,
      label: document.label,
      exitCode: document.exitCode,
      complete,
      passedCount: document.passedCount ?? 0,
      failedCount: document.failedCount ?? 0,
      summaryPath: filePath,
    });

    if (!complete) {
      incompletePatterns.push(pattern);
      continue;
    }

    if (document.exitCode !== 0 || (document.failedCount ?? 0) > 0) {
      failedPatterns.push(pattern);
    }

    passed.push(...(document.passed ?? []));
    failed.push(...(document.failed ?? []));
    subtests.push(...(document.subtests ?? []));
  }

  const exitCode = missingPatterns.length === 0 && incompletePatterns.length === 0 && failedPatterns.length === 0
    ? 0
    : 1;

  const logLines = [
    '# aggregate mode',
    `label: ${options.label}`,
    `expected_families: ${expectedPatterns.length}`,
    `covered_families: ${caseRuns.length}`,
    `missing_families: ${missingPatterns.length}`,
    `incomplete_families: ${incompletePatterns.length}`,
    `failed_families: ${failedPatterns.length}`,
    '',
    'families:',
    ...caseRuns.map((caseRun) => (
      `- ${caseRun.pattern} label=${caseRun.label} exit=${caseRun.exitCode} complete=${caseRun.complete} passed=${caseRun.passedCount} failed=${caseRun.failedCount}`
    )),
  ];

  if (missingPatterns.length > 0) {
    logLines.push('', `missing: ${missingPatterns.join(', ')}`);
  }
  if (incompletePatterns.length > 0) {
    logLines.push('', `incomplete: ${incompletePatterns.join(', ')}`);
  }
  if (failedPatterns.length > 0) {
    logLines.push('', `failed: ${failedPatterns.join(', ')}`);
  }

  const combinedOutput = `${logLines.join('\n')}\n`;
  const summaryDocument = {
    label: options.label,
    mode: 'aggregate',
    reporter: options.reporter,
    pattern: null,
    skippedBuild: true,
    buildExitCode: 0,
    testExitCode: exitCode,
    exitCode,
    complete: exitCode === 0,
    familyCount: expectedPatterns.length,
    coveredFamilyCount: caseRuns.length,
    missingPatterns,
    incompletePatterns,
    failedPatterns,
    caseRuns,
    passedCount: passed.length,
    failedCount: failed.length,
    passed,
    failed,
    subtests,
    logPath,
    exitPath,
  };

  writeArtifacts({
    logPath,
    exitPath,
    summaryPath,
    combinedOutput,
    summaryDocument,
    exitMarker: `EXIT:${exitCode}`,
  });

  return { exitCode, combinedOutput, summaryDocument };
}

function writeArtifacts({
  logPath,
  exitPath,
  summaryPath,
  combinedOutput,
  summaryDocument,
  exitMarker,
}) {
  writeFileSync(logPath, combinedOutput, 'utf8');
  writeFileSync(exitPath, `${exitMarker}\n`, 'utf8');
  writeFileSync(summaryPath, `${JSON.stringify(summaryDocument, null, 2)}\n`, 'utf8');
}

function runTapSlice(pattern, options) {
  return runTapSliceAsync(pattern, options);
}

function runTapSliceAsync(pattern, options) {
  const testArgs = ['--test', '--test-reporter', options.reporter];
  if (pattern) {
    testArgs.push('--test-name-pattern', pattern);
  }
  testArgs.push(options.testFile);

  return new Promise((resolve) => {
    const child = spawn(tsxBinary, testArgs, {
      cwd: packageRoot,
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    let stdout = '';
    let stderr = '';
    let tick = 0;
    const keepAlive = setInterval(() => {
      tick += 1;
      console.log(`Slice ${pattern} still running (${tick * 5}s)`);
    }, 5_000);

    child.stdout.setEncoding('utf8');
    child.stderr.setEncoding('utf8');
    child.stdout.on('data', (chunk) => {
      stdout += chunk;
    });
    child.stderr.on('data', (chunk) => {
      stderr += chunk;
    });

    const finish = (status, extraError = '') => {
      clearInterval(keepAlive);
      const output = [
        `$ ${formatCommand(tsxBinary, testArgs)}\n`,
        stdout,
        stderr,
        extraError,
      ].join('');

      resolve({
        pattern,
        status: status ?? 1,
        output,
        summary: parseTapSummary(stdout + stderr),
      });
    };

    child.on('error', (error) => {
      finish(1, `${error.message}\n`);
    });

    child.on('close', (code) => {
      finish(code ?? 1);
    });
  });
}

const options = parseArgs(process.argv.slice(2));
mkdirSync(options.outputDir, { recursive: true });

const logPath = path.join(options.outputDir, `${options.label}.log`);
const exitPath = path.join(options.outputDir, `${options.label}.exit`);
const summaryPath = path.join(options.outputDir, `${options.label}.summary.json`);

let combinedOutput = '';
let buildStatus = 0;
let testStatus = 0;

const summary = {
  passed: [],
  failed: [],
  subtests: [],
};
const caseRuns = [];

function buildSummaryDocument(complete, exitCode) {
  return {
    label: options.label,
    mode: options.pattern ? 'pattern' : 'full-suite-chunked',
    pattern: options.pattern || null,
    reporter: options.reporter,
    skippedBuild: options.skipBuild,
    buildExitCode: buildStatus,
    testExitCode: testStatus,
    exitCode,
    complete,
    caseRuns,
    passedCount: summary.passed.length,
    failedCount: summary.failed.length,
    passed: summary.passed,
    failed: summary.failed,
    subtests: summary.subtests,
    logPath,
    exitPath,
  };
}

if (!options.skipBuild) {
  const buildArgs = ['run', 'build'];
  const buildResult = runCommand(npmBinary, buildArgs, packageRoot);
  buildStatus = buildResult.status ?? 1;
  combinedOutput += `$ ${formatCommand(npmBinary, buildArgs)}\n`;
  combinedOutput += buildResult.stdout ?? '';
  combinedOutput += buildResult.stderr ?? '';

  if (buildResult.error) {
    combinedOutput += `${buildResult.error.message}\n`;
  }
}

writeArtifacts({
  logPath,
  exitPath,
  summaryPath,
  combinedOutput,
  summaryDocument: buildSummaryDocument(false, buildStatus === 0 ? 'RUNNING' : buildStatus),
  exitMarker: buildStatus === 0 ? 'EXIT:RUNNING' : `EXIT:${buildStatus}`,
});

async function main() {
if (options.aggregate) {
  const aggregateResult = aggregateSummaries(options, logPath, exitPath, summaryPath);
  console.log(`Artifacts written to ${options.outputDir}`);
  console.log(`Log: ${logPath}`);
  console.log(`Exit: ${exitPath}`);
  console.log(`Summary: ${summaryPath}`);
  console.log(`Covered families: ${aggregateResult.summaryDocument.coveredFamilyCount}/${aggregateResult.summaryDocument.familyCount}`);
  console.log(`Missing families: ${aggregateResult.summaryDocument.missingPatterns.length}`);
  console.log(`Incomplete families: ${aggregateResult.summaryDocument.incompletePatterns.length}`);
  console.log(`Failed families: ${aggregateResult.summaryDocument.failedPatterns.length}`);
  console.log(`EXIT:${aggregateResult.exitCode}`);
  process.exit(aggregateResult.exitCode);
}

if (buildStatus === 0) {
  const patterns = options.pattern ? [options.pattern] : loadCaseIds();

  console.log(`Running ${patterns.length} conformance slice(s)...`);

  for (const pattern of patterns) {
    console.log(`Starting slice: ${pattern}`);
    const slice = await runTapSlice(pattern, options);
    combinedOutput += `\n## pattern: ${pattern}\n`;
    combinedOutput += slice.output;

    summary.passed.push(...slice.summary.passed);
    summary.failed.push(...slice.summary.failed);
    summary.subtests.push(...slice.summary.subtests);
    caseRuns.push({
      pattern,
      exitCode: slice.status,
      passedCount: slice.summary.passed.length,
      failedCount: slice.summary.failed.length,
      passed: slice.summary.passed,
      failed: slice.summary.failed,
    });

    if (slice.status !== 0 && testStatus === 0) {
      testStatus = slice.status;
    }

    writeArtifacts({
      logPath,
      exitPath,
      summaryPath,
      combinedOutput,
      summaryDocument: buildSummaryDocument(false, testStatus === 0 ? 'RUNNING' : testStatus),
      exitMarker: testStatus === 0 ? 'EXIT:RUNNING' : `EXIT:${testStatus}`,
    });

    console.log(
      `Finished slice: ${pattern} (exit=${slice.status}, passed=${slice.summary.passed.length}, failed=${slice.summary.failed.length})`,
    );
  }
} else {
  testStatus = buildStatus;
}

const exitCode = buildStatus === 0 ? testStatus : buildStatus;
const summaryDocument = buildSummaryDocument(true, exitCode);

writeArtifacts({
  logPath,
  exitPath,
  summaryPath,
  combinedOutput,
  summaryDocument,
  exitMarker: `EXIT:${exitCode}`,
});

console.log(`Artifacts written to ${options.outputDir}`);
console.log(`Log: ${logPath}`);
console.log(`Exit: ${exitPath}`);
console.log(`Summary: ${summaryPath}`);
console.log(`Passed subtests: ${summary.passed.length}`);
console.log(`Failed subtests: ${summary.failed.length}`);

const outputLines = combinedOutput.trimEnd().split(/\r?\n/);
const tailLines = outputLines.slice(Math.max(0, outputLines.length - options.showLines));
if (tailLines.length > 0) {
  console.log('--- tail ---');
  console.log(tailLines.join('\n'));
}
console.log(`EXIT:${exitCode}`);

process.exit(exitCode);
}

await main();