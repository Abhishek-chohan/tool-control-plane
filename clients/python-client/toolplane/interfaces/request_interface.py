"""Request management interface definitions."""

import uuid
from abc import ABC, abstractmethod
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Protocol, runtime_checkable


class RequestStatus(Enum):
    """Request status enumeration."""

    PENDING = "pending"
    RUNNING = "running"
    DONE = "done"
    FAILURE = "failure"
    PROCESSING = "running"
    COMPLETED = "done"
    FAILED = "failure"
    TIMEOUT = "timeout"
    CANCELLED = "cancelled"


class RequestPriority(Enum):
    """Request priority levels."""

    LOW = "low"
    NORMAL = "normal"
    HIGH = "high"
    URGENT = "urgent"


@dataclass
class RequestInfo:
    """Information about a request."""

    request_id: str
    session_id: str
    tool_name: str
    parameters: Dict[str, Any]
    status: RequestStatus
    priority: RequestPriority = RequestPriority.NORMAL
    created_at: datetime = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    result: Any = None
    error: Optional[str] = None
    metadata: Dict[str, Any] = None
    retry_count: int = 0
    max_retries: int = 3

    def __post_init__(self):
        if self.created_at is None:
            self.created_at = datetime.now()
        if self.metadata is None:
            self.metadata = {}
        if not self.request_id:
            self.request_id = str(uuid.uuid4())

    @property
    def execution_time(self) -> Optional[float]:
        """Get execution time in seconds."""
        if self.started_at and self.completed_at:
            return (self.completed_at - self.started_at).total_seconds()
        return None

    @property
    def wait_time(self) -> Optional[float]:
        """Get wait time in seconds."""
        if self.started_at:
            return (self.started_at - self.created_at).total_seconds()
        return None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "request_id": self.request_id,
            "session_id": self.session_id,
            "tool_name": self.tool_name,
            "parameters": self.parameters,
            "status": self.status.value,
            "priority": self.priority.value,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "started_at": self.started_at.isoformat() if self.started_at else None,
            "completed_at": (
                self.completed_at.isoformat() if self.completed_at else None
            ),
            "result": self.result,
            "error": self.error,
            "metadata": self.metadata,
            "retry_count": self.retry_count,
            "max_retries": self.max_retries,
            "execution_time": self.execution_time,
            "wait_time": self.wait_time,
        }


@runtime_checkable
class IRequestProcessor(Protocol):
    """Protocol interface for request processing."""

    def process_request(self, request: RequestInfo) -> RequestInfo:
        """Process a single request."""
        ...

    def can_process(self, request: RequestInfo) -> bool:
        """Check if processor can handle this request."""
        ...

    def get_processing_capacity(self) -> int:
        """Get current processing capacity."""
        ...


@runtime_checkable
class IRequestManager(Protocol):
    """Protocol interface for request management."""

    def submit_request(
        self,
        session_id: str,
        tool_name: str,
        parameters: Dict[str, Any],
        priority: RequestPriority = RequestPriority.NORMAL,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> str:
        """Submit a new request."""
        ...

    def get_request_status(self, request_id: str) -> Optional[RequestInfo]:
        """Get request status."""
        ...

    def cancel_request(self, request_id: str) -> bool:
        """Cancel a request."""
        ...

    def list_requests(
        self, session_id: Optional[str] = None, status: Optional[RequestStatus] = None
    ) -> List[RequestInfo]:
        """List requests with optional filtering."""
        ...

    def start_polling(self, interval: float) -> None:
        """Start request polling."""
        ...

    def stop_polling(self) -> None:
        """Stop request polling."""
        ...


class IRequestQueue(ABC):
    """Abstract interface for request queuing."""

    @abstractmethod
    def enqueue(self, request: RequestInfo) -> None:
        """Add request to queue."""
        pass

    @abstractmethod
    def dequeue(self) -> Optional[RequestInfo]:
        """Remove and return next request from queue."""
        pass

    @abstractmethod
    def peek(self) -> Optional[RequestInfo]:
        """Look at next request without removing it."""
        pass

    @abstractmethod
    def size(self) -> int:
        """Get queue size."""
        pass

    @abstractmethod
    def is_empty(self) -> bool:
        """Check if queue is empty."""
        pass

    @abstractmethod
    def clear(self) -> None:
        """Clear all requests from queue."""
        pass

    @abstractmethod
    def remove_request(self, request_id: str) -> bool:
        """Remove specific request from queue."""
        pass


class PriorityRequestQueue(IRequestQueue):
    """Priority-based request queue implementation."""

    def __init__(self):
        import heapq

        self._queue = []
        self._entry_finder = {}  # Maps request_id to entry
        self._counter = 0  # Unique sequence count
        self.heapq = heapq

    def enqueue(self, request: RequestInfo) -> None:
        """Add request to priority queue."""
        # Lower number = higher priority
        priority_map = {
            RequestPriority.URGENT: 0,
            RequestPriority.HIGH: 1,
            RequestPriority.NORMAL: 2,
            RequestPriority.LOW: 3,
        }

        priority = priority_map.get(request.priority, 2)

        # Remove existing entry if updating
        if request.request_id in self._entry_finder:
            self.remove_request(request.request_id)

        count = self._counter
        self._counter += 1

        # Entry format: [priority, count, request]
        entry = [priority, count, request]
        self._entry_finder[request.request_id] = entry
        self.heapq.heappush(self._queue, entry)

    def dequeue(self) -> Optional[RequestInfo]:
        """Remove and return highest priority request."""
        while self._queue:
            priority, count, request = self.heapq.heappop(self._queue)

            # Check if this entry was marked as removed
            if request.request_id not in self._entry_finder:
                continue

            del self._entry_finder[request.request_id]
            return request

        return None

    def peek(self) -> Optional[RequestInfo]:
        """Look at highest priority request without removing it."""
        while self._queue:
            priority, count, request = self._queue[0]

            # Check if this entry was marked as removed
            if request.request_id not in self._entry_finder:
                self.heapq.heappop(self._queue)
                continue

            return request

        return None

    def size(self) -> int:
        """Get queue size."""
        return len(self._entry_finder)

    def is_empty(self) -> bool:
        """Check if queue is empty."""
        return len(self._entry_finder) == 0

    def clear(self) -> None:
        """Clear all requests from queue."""
        self._queue.clear()
        self._entry_finder.clear()
        self._counter = 0

    def remove_request(self, request_id: str) -> bool:
        """Remove specific request from queue."""
        if request_id in self._entry_finder:
            self._entry_finder.pop(request_id)
            # Mark as removed by clearing the request_id
            # The actual removal happens during dequeue
            return True
        return False


class FIFORequestQueue(IRequestQueue):
    """First-in-first-out request queue implementation."""

    def __init__(self):
        from collections import deque

        self._queue = deque()
        self._request_ids = set()

    def enqueue(self, request: RequestInfo) -> None:
        """Add request to end of queue."""
        if request.request_id not in self._request_ids:
            self._queue.append(request)
            self._request_ids.add(request.request_id)

    def dequeue(self) -> Optional[RequestInfo]:
        """Remove and return first request from queue."""
        if self._queue:
            request = self._queue.popleft()
            self._request_ids.remove(request.request_id)
            return request
        return None

    def peek(self) -> Optional[RequestInfo]:
        """Look at first request without removing it."""
        return self._queue[0] if self._queue else None

    def size(self) -> int:
        """Get queue size."""
        return len(self._queue)

    def is_empty(self) -> bool:
        """Check if queue is empty."""
        return len(self._queue) == 0

    def clear(self) -> None:
        """Clear all requests from queue."""
        self._queue.clear()
        self._request_ids.clear()

    def remove_request(self, request_id: str) -> bool:
        """Remove specific request from queue."""
        if request_id in self._request_ids:
            # Find and remove the request
            for i, request in enumerate(self._queue):
                if request.request_id == request_id:
                    del self._queue[i]
                    self._request_ids.remove(request_id)
                    return True
        return False


class RequestQueueFactory:
    """Factory for creating request queues."""

    _queue_types = {
        "priority": PriorityRequestQueue,
        "fifo": FIFORequestQueue,
    }

    @classmethod
    def create_queue(cls, queue_type: str = "priority") -> IRequestQueue:
        """Create a request queue of specified type."""
        if queue_type not in cls._queue_types:
            raise ValueError(f"Unknown queue type: {queue_type}")

        queue_class = cls._queue_types[queue_type]
        return queue_class()

    @classmethod
    def register_queue_type(cls, name: str, queue_class: type) -> None:
        """Register a new queue type."""
        cls._queue_types[name] = queue_class


class RequestMetrics:
    """Metrics collection for request processing."""

    def __init__(self):
        self.reset()

    def reset(self) -> None:
        """Reset all metrics."""
        self.total_requests = 0
        self.completed_requests = 0
        self.failed_requests = 0
        self.cancelled_requests = 0
        self.timeout_requests = 0
        self.total_execution_time = 0.0
        self.total_wait_time = 0.0
        self.requests_by_tool = {}
        self.requests_by_session = {}
        self.requests_by_priority = {p: 0 for p in RequestPriority}

    def record_request(self, request: RequestInfo) -> None:
        """Record request metrics."""
        self.total_requests += 1

        # Count by status
        if request.status == RequestStatus.COMPLETED:
            self.completed_requests += 1
        elif request.status == RequestStatus.FAILED:
            self.failed_requests += 1
        elif request.status == RequestStatus.CANCELLED:
            self.cancelled_requests += 1
        elif request.status == RequestStatus.TIMEOUT:
            self.timeout_requests += 1

        # Record timing
        if request.execution_time:
            self.total_execution_time += request.execution_time
        if request.wait_time:
            self.total_wait_time += request.wait_time

        # Count by tool
        tool_name = request.tool_name
        self.requests_by_tool[tool_name] = self.requests_by_tool.get(tool_name, 0) + 1

        # Count by session
        session_id = request.session_id
        self.requests_by_session[session_id] = (
            self.requests_by_session.get(session_id, 0) + 1
        )

        # Count by priority
        self.requests_by_priority[request.priority] += 1

    def get_metrics(self) -> Dict[str, Any]:
        """Get all metrics."""
        success_rate = (
            self.completed_requests / self.total_requests
            if self.total_requests > 0
            else 0
        )

        avg_execution_time = (
            self.total_execution_time / self.completed_requests
            if self.completed_requests > 0
            else 0
        )

        avg_wait_time = (
            self.total_wait_time / self.total_requests if self.total_requests > 0 else 0
        )

        return {
            "total_requests": self.total_requests,
            "completed_requests": self.completed_requests,
            "failed_requests": self.failed_requests,
            "cancelled_requests": self.cancelled_requests,
            "timeout_requests": self.timeout_requests,
            "success_rate": success_rate,
            "avg_execution_time": avg_execution_time,
            "avg_wait_time": avg_wait_time,
            "requests_by_tool": self.requests_by_tool.copy(),
            "requests_by_session": self.requests_by_session.copy(),
            "requests_by_priority": {
                p.value: count for p, count in self.requests_by_priority.items()
            },
        }
