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