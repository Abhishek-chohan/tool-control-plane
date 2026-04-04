"""Constants used across Toolplane client implementations."""

# Default configuration values
DEFAULT_HEARTBEAT_INTERVAL = 60
DEFAULT_MAX_WORKERS = 10
DEFAULT_POLL_INTERVAL = 0.5
DEFAULT_REQUEST_TIMEOUT = 60
DEFAULT_MAX_RETRIES = 3
DEFAULT_BUFFER_SIZE = 4 * 1024 * 1024  # 4MB
DEFAULT_RETRY_BACKOFF_MS = 250

# Cache settings
CACHE_TTL_SECONDS = 0.5

# Protocol-specific defaults
GRPC_DEFAULT_PORT = 50051
HTTP_DEFAULT_PORT = 8080

# Session settings
MAX_SESSION_NAME_LENGTH = 100
MAX_TOOL_NAME_LENGTH = 50
MAX_DESCRIPTION_LENGTH = 500

# Tool execution settings
MAX_TOOL_EXECUTION_TIME = 300  # 5 minutes
STREAM_CHUNK_SIZE = 1024
BACKPRESSURE_THRESHOLD = 0.8  # 80% of buffer size

# Error messages
ERROR_CONNECTION_FAILED = "Failed to connect to server"
ERROR_TOOL_NOT_FOUND = "Tool not found"
ERROR_SESSION_NOT_FOUND = "Session not found"
ERROR_INVALID_PARAMETERS = "Invalid parameters"
ERROR_TIMEOUT = "Operation timed out"
ERROR_STREAM_INTERRUPTED = "Stream was interrupted"
