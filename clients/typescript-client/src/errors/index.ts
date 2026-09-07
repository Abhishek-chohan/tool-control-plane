/**
 * Base Toolplane error class
 */
export class ToolplaneError extends Error {
  public readonly code: number;
  public readonly data?: any;

  constructor(message: string, code: number = -1, data?: any) {
    super(message);
    this.name = 'ToolplaneError';
    this.code = code;
    this.data = data;
    Error.captureStackTrace(this, this.constructor);
  }
}

/**
 * Connection error class
 */
export class ConnectionError extends ToolplaneError {
  constructor(message: string, data?: any) {
    super(message, -1, data);
    this.name = 'ConnectionError';
  }
}

/**
 * Timeout error class
 */
export class TimeoutError extends ToolplaneError {
  constructor(message: string = 'Request timeout', data?: any) {
    super(message, -2, data);
    this.name = 'TimeoutError';
  }
}

/**
 * Protocol error class
 */
export class ProtocolError extends ToolplaneError {
  constructor(message: string, data?: any) {
    super(message, -3, data);
    this.name = 'ProtocolError';
  }
}

/**
 * Validation error class
 */
export class ValidationError extends ToolplaneError {
  constructor(message: string, data?: any) {
    super(message, -4, data);
    this.name = 'ValidationError';
  }
}

/**
 * Context attached to a typed API error: the request the failing call
 * operated on (when known) and the last-known request status for poll-style
 * calls.
 */
export interface APIErrorContext {
  requestId?: string;
  status?: string;
  data?: any;
}

// Status names where re-issuing the same call can plausibly succeed (the
// server was momentarily unreachable, or a capacity limit will clear).
// Everything else — credentials, arguments, state conflicts, missing
// entities — is deterministic.
const RETRYABLE_CODE_NAMES = new Set(['UNAVAILABLE', 'RESOURCE_EXHAUSTED']);

/**
 * A server-reported failure carrying the gRPC status code of the failed
 * call. `code` is the numeric gRPC status and `codeName` its name (e.g.
 * "NOT_FOUND", "FAILED_PRECONDITION"); `retryable` states whether a retry
 * can help.
 */
export class APIError extends ToolplaneError {
  public readonly codeName: string;
  public readonly retryable: boolean;
  public readonly requestId?: string;
  public readonly status?: string;

  constructor(message: string, code: number, codeName: string, context: APIErrorContext = {}) {
    super(message, code, context.data);
    this.name = 'ToolplaneAPIError';
    this.codeName = codeName;
    this.retryable = RETRYABLE_CODE_NAMES.has(codeName);
    this.requestId = context.requestId;
    this.status = context.status;
  }
}

/** The targeted session, request, machine, tool, key, or task is missing. */
export class NotFoundError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 5, 'NOT_FOUND', context);
    this.name = 'NotFoundError';
  }
}

/** The server rejected the request payload (INVALID_ARGUMENT / OUT_OF_RANGE). */
export class InvalidArgumentError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 3, 'INVALID_ARGUMENT', context);
    this.name = 'InvalidArgumentError';
  }
}

/**
 * The operation conflicts with server state: a lost claim race, a stale
 * lease, a draining machine, a terminal request, or no provider registered
 * for the tool.
 */
export class FailedPreconditionError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 9, 'FAILED_PRECONDITION', context);
    this.name = 'FailedPreconditionError';
  }
}

/** A capacity limit was hit; retryable with backoff. */
export class ResourceExhaustedError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 8, 'RESOURCE_EXHAUSTED', context);
    this.name = 'ResourceExhaustedError';
  }
}

/** The API key is missing, unknown, or revoked; retrying cannot succeed. */
export class UnauthenticatedError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 16, 'UNAUTHENTICATED', context);
    this.name = 'UnauthenticatedError';
  }
}

/** The caller lacks the required capability or machine credential. */
export class PermissionDeniedError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 7, 'PERMISSION_DENIED', context);
    this.name = 'PermissionDeniedError';
  }
}

/** The created entity already exists. */
export class AlreadyExistsError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 6, 'ALREADY_EXISTS', context);
    this.name = 'AlreadyExistsError';
  }
}

/** The server was momentarily unreachable; retryable with backoff. */
export class UnavailableError extends APIError {
  constructor(message: string, context: APIErrorContext = {}) {
    super(message, 14, 'UNAVAILABLE', context);
    this.name = 'UnavailableError';
  }
}