import fs from 'node:fs';
import path from 'node:path';

import { GrpcConformanceAdapter } from './adapters/grpc_adapter';
import { HttpConformanceAdapter } from './adapters/http_adapter';
import {
  assertApiKeyFieldEquals,
  assertApiKeyCapabilitiesEqual,
  assertApiKeyIdNonEmpty,
  assertApiKeyListContains,
  assertApiKeyListExcludes,
  assertApiKeyPreviewNonEmpty,
  assertApiKeyValueEmpty,
  assertApiKeyValueNonEmpty,
  assertChunkWindowEdge,
  assertChunkWindowFieldEquals,
  assertChunkWindowLength,
  assertContainsSession,
  assertErrorCodeEquals,
  assertFinalMarker,
  assertMachineFieldEquals,
  assertMachineIdNonEmpty,
  assertMachineListContains,
  assertMachineListExcludes,
  assertRequestFieldEquals,
  assertRequestFieldNonEmpty,
  assertRequestIdNonEmpty,
  assertRequestListContains,
  assertRequestStatus,
  assertSessionContextPresent,
  assertSessionFieldEquals,
  assertSessionIdNonEmpty,
  assertSessionsArray,
  assertStreamChunks,
  assertSuccessTrue,
  assertToolFieldEquals,
  assertToolIdNonEmpty,
  assertToolListContains,
  assertToolListExcludes,
  assertUnaryResult,
} from './assertions';
import type { ConformanceAdapter, ConformanceCase, SupportedFeature, Transport } from './types';

export const SUPPORTED_FEATURES = new Set<SupportedFeature>([
  'session_create',
  'session_list',
  'invoke_unary',
  'invoke_stream',
  'tool_discovery',
  'session_update',
  'request_create',
  'request_recovery',
  'api_key_lifecycle',
  'machine_lifecycle',
  'provider_runtime',
]);

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function numberValue(value: unknown, fallback: number): number {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string') {
    const parsed = Number.parseInt(value, 10);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return fallback;
}

async function waitForRunningRequest(
  adapter: ConformanceAdapter,
  sessionId: string,
  toolName: string,
  caseId: string,
  transport: Transport,
): Promise<void> {
  const deadline = Date.now() + 5_000;
  while (Date.now() < deadline) {
    const requests = await adapter.listRequests(sessionId, {
      list_status: 'running',
      tool_name_filter: toolName,
      limit: 20,
    });

    if (requests.length > 0) {
      return;
    }

    await sleep(100);
  }

  throw new Error(`[${transport}] ${caseId}: timed out waiting for running request for tool '${toolName}'`);
}

async function waitForChunkWindowProgress(
  adapter: ConformanceAdapter,
  sessionId: string,
  requestId: string,
  minimumNextSeq: number,
  caseId: string,
  transport: Transport,
): Promise<void> {
  const deadline = Date.now() + 10_000;
  while (Date.now() < deadline) {
    const chunkWindow = await adapter.getRequestChunksWindow(sessionId, requestId);
    const nextSeq = numberValue(chunkWindow.nextSeq, 0);
    if (nextSeq >= minimumNextSeq) {
      return;
    }

    const requestStatus = await adapter.getRequestStatus(sessionId, requestId);
    if (requestStatus.status === 'failure') {
      throw new Error(
        `[${transport}] ${caseId}: request ${requestId} failed while waiting for chunk window progress`,
      );
    }

    await sleep(100);
  }

  throw new Error(
    `[${transport}] ${caseId}: timed out waiting for retained chunk window to reach next_seq ${minimumNextSeq}`,
  );
}

async function executeRequestRecoveryCase(
  adapter: ConformanceAdapter,
  sessionId: string,
  request: Record<string, unknown>,
  expected: Record<string, unknown>,
  caseId: string,
  transport: Transport,
): Promise<void> {
  const toolName = String(request.tool_name ?? '');
  await adapter.registerStreamTool(
    sessionId,
    toolName,
    String(request.tool_description ?? 'conformance request recovery tool'),
  );

  const requestId = await adapter.startStreamingRequest(
    sessionId,
    toolName,
    (request.params as Record<string, unknown>) ?? {},
  );
  assertRequestIdNonEmpty(requestId, caseId, transport);

  const hasResume = Object.prototype.hasOwnProperty.call(request, 'resume_from_seq');
  const hasMidStreamGate = Object.prototype.hasOwnProperty.call(request, 'wait_for_next_seq_at_least');
  const resumeFromSeq = numberValue(request.resume_from_seq, 0);

  if (hasMidStreamGate) {
    await waitForChunkWindowProgress(
      adapter,
      sessionId,
      requestId,
      numberValue(request.wait_for_next_seq_at_least, 0),
      caseId,
      transport,
    );
  }

  const requestStatus = await adapter.waitForRequestCompletion(sessionId, requestId);
  if ('final_status_equals' in expected) {
    assertRequestStatus(requestStatus, String(expected.final_status_equals), caseId, transport);
  }

  const chunkWindow = await adapter.getRequestChunksWindow(sessionId, requestId);

  if ('ordered_chunks' in expected) {
    assertStreamChunks(chunkWindow.chunks as unknown[], (expected.ordered_chunks as unknown[]) ?? [], caseId, transport);
  }
  if ('chunk_count_equals' in expected) {
    assertChunkWindowLength(chunkWindow, numberValue(expected.chunk_count_equals, 0), caseId, transport);
  }
  if ('start_seq_equals' in expected) {
    assertChunkWindowFieldEquals(chunkWindow, 'startSeq', expected.start_seq_equals, caseId, transport);
  }
  if ('next_seq_equals' in expected) {
    assertChunkWindowFieldEquals(chunkWindow, 'nextSeq', expected.next_seq_equals, caseId, transport);
  }
  if ('first_chunk_equals' in expected) {
    assertChunkWindowEdge(chunkWindow, 'first', expected.first_chunk_equals, caseId, transport);
  }
  if ('last_chunk_equals' in expected) {
    assertChunkWindowEdge(chunkWindow, 'last', expected.last_chunk_equals, caseId, transport);
  }

  let resumeResult: Record<string, unknown> | null = null;
  if (hasResume) {
    resumeResult = await adapter.resumeStream(requestId, resumeFromSeq);
  }

  if (resumeResult) {
    if ('resume_ordered_chunks' in expected) {
      assertStreamChunks(
        (resumeResult.chunks as unknown[]) ?? [],
        (expected.resume_ordered_chunks as unknown[]) ?? [],
        caseId,
        transport,
      );
    }
    if (expected.final_marker === true) {
      assertFinalMarker(resumeResult.sawFinal === true, caseId, transport);
    }
    if ('final_seq_equals' in expected) {
      assertChunkWindowFieldEquals(
        { finalSeq: resumeResult.finalSeq },
        'finalSeq',
        expected.final_seq_equals,
        caseId,
        transport,
      );
    }
    if ('resume_error_code_equals' in expected) {
      assertErrorCodeEquals(resumeResult, String(expected.resume_error_code_equals), caseId, transport);
    }
  }
}

function repoRoot(): string {
  return path.resolve(process.cwd(), '../..');
}

export function loadCasesSync(): ConformanceCase[] {
  const caseDir = path.join(repoRoot(), 'conformance', 'cases');
  const fileNames = fs.readdirSync(caseDir).filter((fileName) => fileName.endsWith('.json')).sort();

  const cases = fileNames.map((fileName) => {
    const filePath = path.join(caseDir, fileName);
    const caseObject = JSON.parse(fs.readFileSync(filePath, 'utf8')) as ConformanceCase;
    validateCaseShape(caseObject, filePath);
    return caseObject;
  });

  if (cases.length === 0) {
    throw new Error(`No conformance cases found in ${caseDir}`);
  }

  return cases;
}

function validateCaseShape(caseObject: ConformanceCase, source: string): void {
  for (const field of ['id', 'feature', 'description', 'request', 'expected']) {
    if (!(field in caseObject)) {
      throw new Error(`Case ${source} missing required field '${field}'`);
    }
  }

  if (!SUPPORTED_FEATURES.has(caseObject.feature)) {
    throw new Error(`Case ${source} has unsupported feature '${caseObject.feature}'`);
  }
}

function adapterForTransport(transport: Transport, userId: string): ConformanceAdapter {
  const apiKey = process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? '';
  if (transport === 'http') {
    const host = process.env.TOOLPLANE_CONFORMANCE_HTTP_HOST ?? 'localhost';
    const port = Number.parseInt(process.env.TOOLPLANE_CONFORMANCE_HTTP_PORT ?? '8080', 10);
    return new HttpConformanceAdapter(host, port, userId, apiKey);
  }

  if (transport === 'grpc') {
    const host = process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ?? 'localhost';
    const port = Number.parseInt(process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ?? '9001', 10);
    return new GrpcConformanceAdapter(host, port, userId, apiKey);
  }

  throw new Error(`Unsupported transport: ${transport}`);
}

export async function executeCase(caseObject: ConformanceCase, transport: Transport): Promise<void> {
  const caseId = caseObject.id;
  const feature = caseObject.feature;
  const request = caseObject.request;
  const expected = caseObject.expected;
  const userId = String(request.user_id ?? process.env.TOOLPLANE_CONFORMANCE_USER_ID ?? 'conformance-user');

  const adapter = adapterForTransport(transport, userId);
  let sessionId = '';

  try {
    await adapter.connect();

    if (feature === 'session_create') {
      sessionId = await adapter.createSession(request);
      assertSessionIdNonEmpty(sessionId, caseId, transport);
      if (expected.session_context_available === true) {
        const context = await adapter.getSessionContext(sessionId);
        assertSessionContextPresent(context, caseId, transport);
      }
      return;
    }

    const sessionRequest = {
      user_id: userId,
      name: String(request.name ?? 'conformance-session'),
      description: String(request.description ?? 'conformance session'),
      namespace: String(request.namespace ?? 'conformance'),
    };

    sessionId = await adapter.createSession(sessionRequest);
    assertSessionIdNonEmpty(sessionId, caseId, transport);

    if (feature === 'session_list') {
      const listResponse = await adapter.listUserSessions(request);
      if (expected.sessions_array_present === true) {
        assertSessionsArray(listResponse, caseId, transport);
      }
      if (expected.contains_created_session === true) {
        assertContainsSession(listResponse, sessionId, caseId, transport);
      }
      return;
    }

    if (feature === 'session_update') {
      const updatedSession = await adapter.updateSession(sessionId, request);
      if (expected.session_id_matches_created === true) {
        assertSessionFieldEquals(updatedSession, 'id', sessionId, caseId, transport);
      }
      if ('name_equals' in expected) {
        assertSessionFieldEquals(updatedSession, 'name', expected.name_equals, caseId, transport);
      }
      if ('description_equals' in expected) {
        assertSessionFieldEquals(updatedSession, 'description', expected.description_equals, caseId, transport);
      }
      if ('namespace_equals' in expected) {
        assertSessionFieldEquals(updatedSession, 'namespace', expected.namespace_equals, caseId, transport);
      }
      return;
    }

    if (feature === 'tool_discovery') {
      const toolName = String(request.tool_name ?? '');
      await adapter.registerUnaryEchoTool(
        sessionId,
        toolName,
        String(request.tool_description ?? 'conformance tool discovery tool'),
      );

      const tools = await adapter.listTools(sessionId);
      const listedTool = tools.find((tool) => tool.name === toolName);

      if (expected.listed_after_register === true) {
        if (!listedTool) {
          throw new Error(`[${transport}] ${caseId}: expected tool '${toolName}' in listed tools`);
        }
      }

      const toolId = String(listedTool?.id ?? '');
      if (expected.tool_id_non_empty === true) {
        assertToolIdNonEmpty(toolId, caseId, transport);
      }
      if (listedTool && expected.session_id_matches_created === true) {
        assertToolFieldEquals(listedTool, 'session_id', sessionId, caseId, transport);
      }
      if (listedTool && 'name_equals' in expected) {
        assertToolFieldEquals(listedTool, 'name', expected.name_equals, caseId, transport);
      }
      if (listedTool && 'description_equals' in expected) {
        assertToolFieldEquals(listedTool, 'description', expected.description_equals, caseId, transport);
      }

      if (expected.lookup_by_id === true) {
        const toolById = await adapter.getToolById(sessionId, toolId);
        assertToolFieldEquals(toolById, 'id', toolId, caseId, transport);
        if ('name_equals' in expected) {
          assertToolFieldEquals(toolById, 'name', expected.name_equals, caseId, transport);
        }
      }

      if (expected.lookup_by_name === true) {
        const toolByName = await adapter.getToolByName(sessionId, toolName);
        assertToolFieldEquals(toolByName, 'id', toolId, caseId, transport);
        if ('description_equals' in expected) {
          assertToolFieldEquals(toolByName, 'description', expected.description_equals, caseId, transport);
        }
      }

      const deleted = await adapter.deleteTool(sessionId, toolId);
      if (expected.delete_success === true) {
        assertSuccessTrue(deleted, 'tool delete result', caseId, transport);
      }

      const toolsAfterDelete = await adapter.listTools(sessionId);
      if (expected.absent_after_delete === true) {
        assertToolListExcludes(toolsAfterDelete, toolId, caseId, transport);
      } else {
        assertToolListContains(toolsAfterDelete, toolId, caseId, transport);
      }
      return;
    }

    if (feature === 'request_create') {
      const toolName = String(request.tool_name ?? '');
      await adapter.registerUnaryEchoTool(
        sessionId,
        toolName,
        String(request.tool_description ?? 'conformance request tool'),
      );

      const requestId = await adapter.createRequest(sessionId, toolName, (request.params as Record<string, unknown>) ?? {});
      assertRequestIdNonEmpty(requestId, caseId, transport);

      const requestStatus = await adapter.getRequestStatus(sessionId, requestId);
      if ('status_equals' in expected) {
        assertRequestStatus(requestStatus, String(expected.status_equals), caseId, transport);
      }
      if ('tool_name_equals' in expected) {
        assertRequestFieldEquals(requestStatus, 'toolName', expected.tool_name_equals, caseId, transport);
      }
      if (expected.request_id_matches_created === true) {
        assertRequestFieldEquals(requestStatus, 'id', requestId, caseId, transport);
      }

      const listedRequests = await adapter.listRequests(sessionId, request);
      if (expected.listed_request_present === true) {
        assertRequestListContains(listedRequests, requestId, caseId, transport);
      }
      return;
    }

    if (feature === 'request_recovery') {
      await executeRequestRecoveryCase(adapter, sessionId, request, expected, caseId, transport);
      return;
    }

    if (feature === 'api_key_lifecycle') {
    const requestedCapabilities = Array.isArray(request.api_key_capabilities)
      ? request.api_key_capabilities.map((value) => String(value))
      : undefined;
    const apiKey = await adapter.createApiKey(
      sessionId,
      String(request.api_key_name ?? 'conformance-key'),
      requestedCapabilities,
    );
      if (expected.api_key_id_non_empty === true) {
        assertApiKeyIdNonEmpty(apiKey, caseId, transport);
      }
      if (expected.api_key_value_non_empty === true) {
        assertApiKeyValueNonEmpty(apiKey, caseId, transport);
      }
    if (expected.key_preview_non_empty === true) {
      assertApiKeyPreviewNonEmpty(apiKey, caseId, transport);
    }
      if (expected.session_id_matches_created === true) {
        assertApiKeyFieldEquals(apiKey, 'session_id', sessionId, caseId, transport);
      }
      if ('name_equals' in expected) {
        assertApiKeyFieldEquals(apiKey, 'name', expected.name_equals, caseId, transport);
      }
    if (Array.isArray(expected.capabilities_equal)) {
      assertApiKeyCapabilitiesEqual(apiKey, expected.capabilities_equal.map((value) => String(value)), caseId, transport);
    }

      const listedApiKeys = await adapter.listApiKeys(sessionId);
      if (expected.listed_after_create === true) {
        assertApiKeyListContains(listedApiKeys, String(apiKey.id ?? ''), caseId, transport);
      }
    const listedApiKey = listedApiKeys.find((entry) => entry.id === apiKey.id);
    if (expected.listed_key_value_empty === true && listedApiKey) {
      assertApiKeyValueEmpty(listedApiKey, caseId, transport);
    }
    if (expected.listed_key_preview_non_empty === true && listedApiKey) {
      assertApiKeyPreviewNonEmpty(listedApiKey, caseId, transport);
    }
    if (Array.isArray(expected.listed_capabilities_equal) && listedApiKey) {
      assertApiKeyCapabilitiesEqual(
        listedApiKey,
        expected.listed_capabilities_equal.map((value) => String(value)),
        caseId,
        transport,
      );
    }

      const revokeSuccess = await adapter.revokeApiKey(sessionId, String(apiKey.id ?? ''));
      if (expected.revoke_success === true) {
        assertSuccessTrue(revokeSuccess, 'api key revoke result', caseId, transport);
      }

      const listedAfterRevoke = await adapter.listApiKeys(sessionId);
      if (expected.absent_after_revoke === true) {
        assertApiKeyListExcludes(listedAfterRevoke, String(apiKey.id ?? ''), caseId, transport);
      }
      return;
    }

    if (feature === 'machine_lifecycle') {
      const machine = await adapter.registerMachine(sessionId, request);
      const machineId = String(machine.id ?? '');

      if (expected.machine_id_non_empty === true) {
        assertMachineIdNonEmpty(machine, caseId, transport);
      }
      if (expected.session_id_matches_created === true) {
        assertMachineFieldEquals(machine, 'session_id', sessionId, caseId, transport);
      }
      if ('sdk_version_equals' in expected) {
        assertMachineFieldEquals(machine, 'sdk_version', expected.sdk_version_equals, caseId, transport);
      }
      if ('sdk_language_equals' in expected) {
        assertMachineFieldEquals(machine, 'sdk_language', expected.sdk_language_equals, caseId, transport);
      }

      const listedMachines = await adapter.listMachines(sessionId);
      if (expected.listed_after_register === true) {
        assertMachineListContains(listedMachines, machineId, caseId, transport);
      }

      const fetchedMachine = await adapter.getMachine(sessionId, machineId);
      if (expected.retrieved_by_id === true) {
        assertMachineFieldEquals(fetchedMachine, 'id', machineId, caseId, transport);
      }

      if ('inflight_result_equals' in expected) {
        const toolName = String(request.tool_name ?? '');
        await adapter.registerUnaryEchoTool(
          sessionId,
          toolName,
          String(request.tool_description ?? 'conformance busy drain tool'),
        );

        const invokePromise = adapter.invoke(
          sessionId,
          toolName,
          (request.invoke_params as Record<string, unknown>) ?? {},
        );

        await waitForRunningRequest(adapter, sessionId, toolName, caseId, transport);

        const drainPromise = adapter.drainMachine(sessionId, machineId);

        let invokeResult: unknown;
        try {
          invokeResult = await invokePromise;
        } catch (error) {
          await drainPromise.catch(() => undefined);
          throw error;
        }

        assertUnaryResult(
          invokeResult,
          (expected.inflight_result_equals as Record<string, unknown>) ?? {},
          caseId,
          transport,
        );

        const drained = await drainPromise;
        if (expected.drain_success === true) {
          assertSuccessTrue(drained, 'machine drain result', caseId, transport);
        }

        const listedAfterDrain = await adapter.listMachines(sessionId);
        if (expected.absent_after_drain === true) {
          assertMachineListExcludes(listedAfterDrain, machineId, caseId, transport);
        }
        return;
      }

      const drained = await adapter.drainMachine(sessionId, machineId);
      if (expected.drain_success === true) {
        assertSuccessTrue(drained, 'machine drain result', caseId, transport);
      }

      const listedAfterDrain = await adapter.listMachines(sessionId);
      if (expected.absent_after_drain === true) {
        assertMachineListExcludes(listedAfterDrain, machineId, caseId, transport);
      }
      return;
    }

    if (feature === 'provider_runtime') {
      const mode = String(request.mode ?? 'unary');
      const toolName = String(request.tool_name ?? '');

      if (mode === 'unary') {
        await adapter.registerUnaryEchoTool(
          sessionId,
          toolName,
          String(request.tool_description ?? 'conformance provider unary tool'),
        );
        await adapter.startProviderRuntime(sessionId);
        const requestId = await adapter.createRequest(sessionId, toolName, (request.params as Record<string, unknown>) ?? {});
        await adapter.startRequestProcessing(sessionId, requestId);

        if (expected.request_id_non_empty === true) {
          assertRequestIdNonEmpty(requestId, caseId, transport);
        }

        const requestStatus = await adapter.waitForRequestCompletion(sessionId, requestId);
        if ('status_equals' in expected) {
          assertRequestStatus(requestStatus, String(expected.status_equals), caseId, transport);
        }
        if (expected.executing_machine_present === true) {
          assertRequestFieldNonEmpty(requestStatus, 'executingMachineId', caseId, transport);
        }
        if ('result_equals' in expected) {
          assertUnaryResult(
            requestStatus.result,
            (expected.result_equals as Record<string, unknown>) ?? {},
            caseId,
            transport,
          );
        }
        return;
      }

      if (mode === 'stream') {
        await adapter.registerStreamTool(
          sessionId,
          toolName,
          String(request.tool_description ?? 'conformance provider stream tool'),
        );
        await adapter.startProviderRuntime(sessionId);
        const requestId = await adapter.createRequest(sessionId, toolName, (request.params as Record<string, unknown>) ?? {});
        await adapter.startRequestProcessing(sessionId, requestId);

        if (expected.request_id_non_empty === true) {
          assertRequestIdNonEmpty(requestId, caseId, transport);
        }

        const requestStatus = await adapter.waitForRequestCompletion(sessionId, requestId);
        if ('status_equals' in expected) {
          assertRequestStatus(requestStatus, String(expected.status_equals), caseId, transport);
        }
        if (expected.executing_machine_present === true) {
          assertRequestFieldNonEmpty(requestStatus, 'executingMachineId', caseId, transport);
        }
        assertStreamChunks(
          (requestStatus.streamResults as unknown[]) ?? [],
          (expected.ordered_chunks as unknown[]) ?? [],
          caseId,
          transport,
        );
        return;
      }

      if (mode === 'drain') {
        await adapter.registerUnaryEchoTool(
          sessionId,
          toolName,
          String(request.tool_description ?? 'conformance provider drain tool'),
        );
        await adapter.startProviderRuntime(sessionId);
        const machines = await adapter.listMachines(sessionId);
        const machineId = String(machines.at(0)?.id ?? '');
        if (!machineId.trim()) {
          throw new Error(`[${transport}] ${caseId}: provider runtime did not attach a machine`);
        }

        const requestId = await adapter.createRequest(sessionId, toolName, (request.params as Record<string, unknown>) ?? {});
        await adapter.startRequestProcessing(sessionId, requestId);
        if (expected.request_id_non_empty === true) {
          assertRequestIdNonEmpty(requestId, caseId, transport);
        }

        await waitForRunningRequest(adapter, sessionId, toolName, caseId, transport);
        const drained = await adapter.drainMachine(sessionId, machineId);
        const requestStatus = await adapter.waitForRequestCompletion(sessionId, requestId);

        if ('status_equals' in expected) {
          assertRequestStatus(requestStatus, String(expected.status_equals), caseId, transport);
        }
        if (expected.executing_machine_present === true) {
          assertRequestFieldNonEmpty(requestStatus, 'executingMachineId', caseId, transport);
        }
        if ('result_equals' in expected) {
          assertUnaryResult(
            requestStatus.result,
            (expected.result_equals as Record<string, unknown>) ?? {},
            caseId,
            transport,
          );
        }
        if (expected.drain_success === true) {
          assertSuccessTrue(drained, 'provider drain result', caseId, transport);
        }

        const listedAfterDrain = await adapter.listMachines(sessionId);
        if (expected.absent_after_drain === true) {
          assertMachineListExcludes(listedAfterDrain, machineId, caseId, transport);
        }
        return;
      }

      throw new Error(`[${transport}] ${caseId}: unsupported provider runtime mode ${mode}`);
    }

    if (feature === 'invoke_unary') {
      const toolName = String(request.tool_name ?? '');
      await adapter.registerUnaryEchoTool(
        sessionId,
        toolName,
        String(request.tool_description ?? 'conformance unary tool'),
      );
      const result = await adapter.invoke(sessionId, toolName, (request.params as Record<string, unknown>) ?? {});
      assertUnaryResult(result, (expected.result_equals as Record<string, unknown>) ?? {}, caseId, transport);
      return;
    }

    if (feature === 'invoke_stream') {
      const toolName = String(request.tool_name ?? '');
      await adapter.registerStreamTool(
        sessionId,
        toolName,
        String(request.tool_description ?? 'conformance stream tool'),
      );
      const [chunks, sawFinal] = await adapter.stream(sessionId, toolName, (request.params as Record<string, unknown>) ?? {});
      assertStreamChunks(chunks, (expected.ordered_chunks as unknown[]) ?? [], caseId, transport);
      if (expected.final_marker === true) {
        assertFinalMarker(sawFinal, caseId, transport);
      }
      return;
    }

    throw new Error(`[${transport}] ${caseId}: unsupported feature ${feature}`);
  } finally {
    await adapter.close();
  }
}