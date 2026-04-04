function assertNonEmptyString(value: unknown, label: string, caseId: string, transport: string): void {
  if (typeof value !== 'string' || value.trim().length === 0) {
    throw new Error(`[${transport}] ${caseId}: ${label} must be non-empty`);
  }
}

function assertFieldEquals(
  payload: Record<string, unknown>,
  field: string,
  expected: unknown,
  label: string,
  caseId: string,
  transport: string,
): void {
  const actual = payload[field];
  if (actual !== expected) {
    throw new Error(
      `[${transport}] ${caseId}: ${label}.${field} mismatch. expected=${String(expected)}, actual=${String(actual)}`,
    );
  }
}

export function assertSessionIdNonEmpty(sessionId: string, caseId: string, transport: string): void {
  assertNonEmptyString(sessionId, 'session_id', caseId, transport);
}

export function assertSessionContextPresent(context: unknown, caseId: string, transport: string): void {
  if (context == null) {
    throw new Error(`[${transport}] ${caseId}: session context should be available`);
  }
}

export function assertSessionsArray(response: Record<string, unknown>, caseId: string, transport: string): void {
  if (!Array.isArray(response.sessions)) {
    throw new Error(`[${transport}] ${caseId}: response.sessions must be a list`);
  }
}

export function assertContainsSession(
  response: Record<string, unknown>,
  sessionId: string,
  caseId: string,
  transport: string,
): void {
  const sessions = Array.isArray(response.sessions) ? response.sessions : [];
  const sessionIds = new Set(
    sessions
      .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
      .map((entry) => entry.id),
  );

  if (!sessionIds.has(sessionId)) {
    throw new Error(`[${transport}] ${caseId}: expected session_id '${sessionId}' in listed sessions`);
  }
}

export function assertUnaryResult(result: unknown, expectedResult: Record<string, unknown>, caseId: string, transport: string): void {
  const actual = JSON.stringify(result);
  const expected = JSON.stringify(expectedResult);
  if (actual !== expected) {
    throw new Error(`[${transport}] ${caseId}: unary result mismatch. expected=${expected}, actual=${actual}`);
  }
}

export function assertStreamChunks(chunks: unknown[], expectedChunks: unknown[], caseId: string, transport: string): void {
  const actual = JSON.stringify(chunks);
  const expected = JSON.stringify(expectedChunks);
  if (actual !== expected) {
    throw new Error(`[${transport}] ${caseId}: stream chunks mismatch. expected=${expected}, actual=${actual}`);
  }
}

export function assertFinalMarker(sawFinal: boolean, caseId: string, transport: string): void {
  if (!sawFinal) {
    throw new Error(`[${transport}] ${caseId}: expected final stream marker`);
  }
}

export function assertSessionFieldEquals(
  session: Record<string, unknown>,
  field: string,
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(session, field, expected, 'session', caseId, transport);
}

export function assertToolIdNonEmpty(toolId: string, caseId: string, transport: string): void {
  assertNonEmptyString(toolId, 'tool.id', caseId, transport);
}

export function assertToolFieldEquals(
  tool: Record<string, unknown>,
  field: string,
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(tool, field, expected, 'tool', caseId, transport);
}

export function assertToolListContains(
  tools: Record<string, unknown>[],
  toolId: string,
  caseId: string,
  transport: string,
): void {
  const toolIds = new Set(tools.map((entry) => entry.id));
  if (!toolIds.has(toolId)) {
    throw new Error(`[${transport}] ${caseId}: expected tool '${toolId}' in listed tools`);
  }
}

export function assertToolListExcludes(
  tools: Record<string, unknown>[],
  toolId: string,
  caseId: string,
  transport: string,
): void {
  const toolIds = new Set(tools.map((entry) => entry.id));
  if (toolIds.has(toolId)) {
    throw new Error(`[${transport}] ${caseId}: did not expect tool '${toolId}' in listed tools`);
  }
}

export function assertRequestIdNonEmpty(requestId: string, caseId: string, transport: string): void {
  assertNonEmptyString(requestId, 'request_id', caseId, transport);
}

export function assertRequestStatus(
  request: Record<string, unknown>,
  expectedStatus: string,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(request, 'status', expectedStatus, 'request', caseId, transport);
}

export function assertRequestFieldEquals(
  request: Record<string, unknown>,
  field: string,
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(request, field, expected, 'request', caseId, transport);
}

export function assertRequestFieldNonEmpty(
  request: Record<string, unknown>,
  field: string,
  caseId: string,
  transport: string,
): void {
  assertNonEmptyString(request[field], `request.${field}`, caseId, transport);
}

export function assertChunkWindowFieldEquals(
  chunkWindow: Record<string, unknown>,
  field: string,
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(chunkWindow, field, expected, 'chunk_window', caseId, transport);
}

export function assertChunkWindowLength(
  chunkWindow: Record<string, unknown>,
  expectedLength: number,
  caseId: string,
  transport: string,
): void {
  const chunks = Array.isArray(chunkWindow.chunks) ? chunkWindow.chunks : [];
  if (chunks.length !== expectedLength) {
    throw new Error(
      `[${transport}] ${caseId}: chunk_window.chunks length mismatch. expected=${expectedLength}, actual=${chunks.length}`,
    );
  }
}

export function assertChunkWindowEdge(
  chunkWindow: Record<string, unknown>,
  edge: 'first' | 'last',
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  const chunks = Array.isArray(chunkWindow.chunks) ? chunkWindow.chunks : [];
  const actual = edge === 'first' ? chunks.at(0) : chunks.at(-1);
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(
      `[${transport}] ${caseId}: chunk_window ${edge} chunk mismatch. expected=${JSON.stringify(expected)}, actual=${JSON.stringify(actual)}`,
    );
  }
}

export function assertErrorCodeEquals(
  payload: Record<string, unknown>,
  expected: string,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(payload, 'errorCode', expected, 'result', caseId, transport);
}

export function assertRequestListContains(
  requests: Record<string, unknown>[],
  requestId: string,
  caseId: string,
  transport: string,
): void {
  const requestIds = new Set(requests.map((entry) => entry.id));
  if (!requestIds.has(requestId)) {
    throw new Error(`[${transport}] ${caseId}: expected request_id '${requestId}' in listed requests`);
  }
}

export function assertApiKeyIdNonEmpty(apiKey: Record<string, unknown>, caseId: string, transport: string): void {
  assertNonEmptyString(apiKey.id, 'api_key.id', caseId, transport);
}

export function assertApiKeyValueNonEmpty(apiKey: Record<string, unknown>, caseId: string, transport: string): void {
  assertNonEmptyString(apiKey.key, 'api_key.key', caseId, transport);
}

export function assertApiKeyFieldEquals(
  apiKey: Record<string, unknown>,
  field: string,
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(apiKey, field, expected, 'api_key', caseId, transport);
}

export function assertApiKeyListContains(
  apiKeys: Record<string, unknown>[],
  keyId: string,
  caseId: string,
  transport: string,
): void {
  const keyIds = new Set(apiKeys.map((entry) => entry.id));
  if (!keyIds.has(keyId)) {
    throw new Error(`[${transport}] ${caseId}: expected api key '${keyId}' in listed api keys`);
  }
}

export function assertApiKeyListExcludes(
  apiKeys: Record<string, unknown>[],
  keyId: string,
  caseId: string,
  transport: string,
): void {
  const keyIds = new Set(apiKeys.map((entry) => entry.id));
  if (keyIds.has(keyId)) {
    throw new Error(`[${transport}] ${caseId}: did not expect api key '${keyId}' in listed api keys`);
  }
}

export function assertMachineIdNonEmpty(machine: Record<string, unknown>, caseId: string, transport: string): void {
  assertNonEmptyString(machine.id, 'machine.id', caseId, transport);
}

export function assertMachineFieldEquals(
  machine: Record<string, unknown>,
  field: string,
  expected: unknown,
  caseId: string,
  transport: string,
): void {
  assertFieldEquals(machine, field, expected, 'machine', caseId, transport);
}

export function assertMachineListContains(
  machines: Record<string, unknown>[],
  machineId: string,
  caseId: string,
  transport: string,
): void {
  const machineIds = new Set(machines.map((entry) => entry.id));
  if (!machineIds.has(machineId)) {
    throw new Error(`[${transport}] ${caseId}: expected machine '${machineId}' in listed machines`);
  }
}

export function assertMachineListExcludes(
  machines: Record<string, unknown>[],
  machineId: string,
  caseId: string,
  transport: string,
): void {
  const machineIds = new Set(machines.map((entry) => entry.id));
  if (machineIds.has(machineId)) {
    throw new Error(`[${transport}] ${caseId}: did not expect machine '${machineId}' in listed machines`);
  }
}

export function assertSuccessTrue(success: boolean, label: string, caseId: string, transport: string): void {
  if (success !== true) {
    throw new Error(`[${transport}] ${caseId}: expected ${label} to be true`);
  }
}