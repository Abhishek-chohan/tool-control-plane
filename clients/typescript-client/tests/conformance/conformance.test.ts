import { after, before, test } from 'node:test';

import { startConformanceEnvironment, type ConformanceEnvironment } from './environment';
import { executeCase, loadCasesSync } from './runner';
import type { Transport } from './types';

const transports: Transport[] = ['http', 'grpc'];
const cases = loadCasesSync();

let environment: ConformanceEnvironment | null = null;

before(async () => {
  environment = await startConformanceEnvironment();
});

after(async () => {
  if (environment) {
    await environment.cleanup();
  }
});

for (const caseObject of cases) {
  for (const transport of transports) {
    test(`${transport}: ${caseObject.id}`, async () => {
      await executeCase(caseObject, transport);
    });
  }
}