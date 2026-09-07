import assert from 'node:assert/strict';
import test from 'node:test';

import {
  AlreadyExistsError,
  APIError,
  ConnectionError,
  FailedPreconditionError,
  InvalidArgumentError,
  NotFoundError,
  PermissionDeniedError,
  ResourceExhaustedError,
  ToolplaneError,
  UnauthenticatedError,
  UnavailableError,
} from '../../src/errors';

test('typed API errors carry code, request id, and status', () => {
  const err = new NotFoundError('request req_1 not found', { requestId: 'req_1', status: 'pending' });
  assert.ok(err instanceof ToolplaneError);
  assert.ok(err instanceof APIError);
  assert.equal(err.codeName, 'NOT_FOUND');
  assert.equal(err.code, 5);
  assert.equal(err.requestId, 'req_1');
  assert.equal(err.status, 'pending');
  assert.equal(err.retryable, false);
});

test('only transport and capacity codes are retryable', () => {
  assert.equal(new UnavailableError('down').retryable, true);
  assert.equal(new ResourceExhaustedError('at capacity').retryable, true);
  assert.equal(new UnauthenticatedError('bad key').retryable, false);
  assert.equal(new PermissionDeniedError('no').retryable, false);
  assert.equal(new InvalidArgumentError('bad').retryable, false);
  assert.equal(new FailedPreconditionError('conflict').retryable, false);
  assert.equal(new AlreadyExistsError('dup').retryable, false);
  assert.equal(new NotFoundError('gone').retryable, false);
});

test('legacy error classes keep their identity for existing callers', () => {
  const conn = new ConnectionError('no route');
  assert.ok(conn instanceof ToolplaneError);
  assert.equal(conn.name, 'ConnectionError');
});
