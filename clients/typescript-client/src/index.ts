/**
 * Toolplane TypeScript Client
 * 
 * A TypeScript client for the Toolplane system with a maintained gRPC
 * control-plane surface.
 */

// Core client
export { ToolplaneClient } from './core/toolplane_client';
export { ProviderRuntime } from './provider_runtime';

// Interfaces and types
export * from './interfaces';

// Error classes
export {
  ToolplaneError,
  APIError,
  AlreadyExistsError,
  ConnectionError,
  FailedPreconditionError,
  InvalidArgumentError,
  NotFoundError,
  PermissionDeniedError,
  ProtocolError,
  ResourceExhaustedError,
  TimeoutError,
  UnauthenticatedError,
  UnavailableError,
  ValidationError,
} from './errors';

// Version info
export const VERSION = '1.0.0';
export const SUPPORTED_PROTOCOLS = ['grpc'] as const;