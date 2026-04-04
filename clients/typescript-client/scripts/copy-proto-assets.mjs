import { cpSync, mkdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageRoot = path.resolve(scriptDir, '..');
const sourceProtoDir = path.join(packageRoot, 'src', 'proto');
const distProtoDir = path.join(packageRoot, 'dist', 'proto');

mkdirSync(distProtoDir, { recursive: true });
cpSync(sourceProtoDir, distProtoDir, { recursive: true });