"""Utility functions shared across Toolplane client implementations."""

import functools
import json
import re
import time
import uuid
from threading import RLock
from typing import Any, Callable, Dict, Optional, TypeVar

from .constants import (
    CACHE_TTL_SECONDS,
    MAX_DESCRIPTION_LENGTH,
    MAX_SESSION_NAME_LENGTH,
    MAX_TOOL_NAME_LENGTH,
)

T = TypeVar("T")


def generate_session_id() -> str:
    """Generate a unique session ID."""
    return str(uuid.uuid4())


def validate_tool_name(name: str) -> bool:
    """Validate tool name format."""
    if not name or len(name) > MAX_TOOL_NAME_LENGTH:
        return False

    # Tool names should be alphanumeric with underscores/hyphens
    return re.match(r"^[a-zA-Z0-9_-]+$", name) is not None


def validate_session_name(name: str) -> bool:
    """Validate session name format."""
    if not name or len(name) > MAX_SESSION_NAME_LENGTH:
        return False

    # Session names allow more characters
    return re.match(r"^[a-zA-Z0-9_\-\s\.]+$", name) is not None


def validate_description(description: str) -> bool:
    """Validate description length."""
    return len(description) <= MAX_DESCRIPTION_LENGTH


def parse_json_safe(json_str: str, default: Any = None) -> Any:
    """Safely parse JSON string."""
    try:
        return json.loads(json_str)
    except (json.JSONDecodeError, TypeError):
        return default


def format_error_message(error: Exception, context: str = "") -> str:
    """Format error message with context."""
    if context:
        return f"{context}: {str(error)}"
    return str(error)


def with_retry(
    max_retries: int = 3, backoff_ms: int = 250, exceptions: tuple = (Exception,)
) -> Callable:
    """Decorator for retrying operations."""

    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        @functools.wraps(func)
        def wrapper(*args, **kwargs) -> T:
            last_exception = None

            for attempt in range(max_retries + 1):
                try:
                    return func(*args, **kwargs)
                except exceptions as e:
                    last_exception = e
                    if attempt < max_retries:
                        # Exponential backoff
                        sleep_time = (backoff_ms * (2**attempt)) / 1000
                        time.sleep(sleep_time)
                    else:
                        break

            # If all retries failed, raise the last exception
            raise last_exception

        return wrapper

    return decorator


class TTLCache:
    """Thread-safe cache with TTL (Time To Live)."""

    def __init__(self, ttl_seconds: int = CACHE_TTL_SECONDS):
        self.ttl_seconds = ttl_seconds
        self.cache: Dict[str, tuple] = {}  # key -> (timestamp, value)
        self.lock = RLock()

    def get(self, key: str) -> Optional[Any]:
        """Get value from cache if not expired."""
        with self.lock:
            if key in self.cache:
                timestamp, value = self.cache[key]
                if time.time() - timestamp < self.ttl_seconds:
                    return value
                else:
                    # Remove expired entry
                    del self.cache[key]
            return None

    def set(self, key: str, value: Any) -> None:
        """Set value in cache with current timestamp."""
        with self.lock:
            self.cache[key] = (time.time(), value)

    def clear(self) -> None:
        """Clear all cache entries."""
        with self.lock:
            self.cache.clear()

    def cleanup_expired(self) -> None:
        """Remove expired entries from cache."""
        now = time.time()
        with self.lock:
            expired_keys = [
                key
                for key, (timestamp, _) in self.cache.items()
                if now - timestamp >= self.ttl_seconds
            ]
            for key in expired_keys:
                del self.cache[key]


def cache_with_ttl(ttl_seconds: int = CACHE_TTL_SECONDS) -> Callable:
    """Decorator for caching function results with TTL."""

    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        cache = TTLCache(ttl_seconds)

        @functools.wraps(func)
        def wrapper(*args, **kwargs) -> T:
            # Create cache key from function name and arguments
            key = f"{func.__name__}:{hash((args, tuple(sorted(kwargs.items()))))}"

            # Try to get from cache first
            cached_result = cache.get(key)
            if cached_result is not None:
                return cached_result

            # If not in cache, execute function and cache result
            result = func(*args, **kwargs)
            cache.set(key, result)
            return result

        # Add cache management methods
        wrapper.cache_clear = cache.clear
        wrapper.cache_cleanup = cache.cleanup_expired

        return wrapper

    return decorator


def normalize_url(url: str, default_protocol: str = "http") -> str:
    """Normalize URL by adding protocol if missing."""
    if not url.startswith(("http://", "https://")):
        url = f"{default_protocol}://{url}"
    return url.rstrip("/")


def extract_host_port(url: str) -> tuple[str, int]:
    """Extract host and port from URL."""
    # Remove protocol if present
    if "://" in url:
        url = url.split("://")[1]

    # Remove path if present
    if "/" in url:
        url = url.split("/")[0]

    # Split host and port
    if ":" in url:
        host, port_str = url.rsplit(":", 1)
        try:
            port = int(port_str)
        except ValueError:
            port = 80 if "http" in url else 443
    else:
        host = url
        port = 80 if "http" in url else 443

    return host, port


def merge_dicts(base: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
    """Merge two dictionaries, with override taking precedence."""
    result = base.copy()
    result.update(override)
    return result


def sanitize_input(value: str, max_length: int = 1000) -> str:
    """Sanitize input string by limiting length and removing control characters."""
    if not isinstance(value, str):
        value = str(value)

    # Remove control characters except newlines and tabs
    sanitized = re.sub(r"[\x00-\x08\x0b\x0c\x0e-\x1f\x7f-\x9f]", "", value)

    # Limit length
    if len(sanitized) > max_length:
        sanitized = sanitized[:max_length]

    return sanitized


def is_valid_json(json_str: str) -> bool:
    """Check if string is valid JSON."""
    try:
        json.loads(json_str)
        return True
    except (json.JSONDecodeError, TypeError):
        return False


def deep_merge(dict1: Dict[str, Any], dict2: Dict[str, Any]) -> Dict[str, Any]:
    """Deep merge two dictionaries."""
    result = dict1.copy()

    for key, value in dict2.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = value

    return result


def timeout_wrapper(timeout_seconds: int) -> Callable:
    """Decorator to add timeout to function execution."""

    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        @functools.wraps(func)
        def wrapper(*args, **kwargs) -> T:
            import threading

            result = [None]
            exception = [None]

            def target():
                try:
                    result[0] = func(*args, **kwargs)
                except Exception as e:
                    exception[0] = e

            thread = threading.Thread(target=target)
            thread.daemon = True
            thread.start()
            thread.join(timeout_seconds)

            if thread.is_alive():
                raise TimeoutError(
                    f"Function {func.__name__} timed out after {timeout_seconds} seconds"
                )

            if exception[0]:
                raise exception[0]

            return result[0]

        return wrapper

    return decorator
