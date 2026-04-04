"""Event system interfaces for decoupled component communication."""

import weakref
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Protocol, runtime_checkable


class EventType(Enum):
    """Standard event types for Toolplane system."""

    # Connection events
    CONNECTION_ESTABLISHED = "connection.established"
    CONNECTION_LOST = "connection.lost"
    CONNECTION_ERROR = "connection.error"

    # Session events
    SESSION_CREATED = "session.created"
    SESSION_DESTROYED = "session.destroyed"
    SESSION_ERROR = "session.error"

    # Tool events
    TOOL_REGISTERED = "tool.registered"
    TOOL_UNREGISTERED = "tool.unregistered"
    TOOL_EXECUTED = "tool.executed"
    TOOL_ERROR = "tool.error"

    # Machine events
    MACHINE_REGISTERED = "machine.registered"
    MACHINE_UNREGISTERED = "machine.unregistered"
    MACHINE_HEARTBEAT_FAILED = "machine.heartbeat_failed"
    MACHINE_RECOVERY_STARTED = "machine.recovery_started"
    MACHINE_RECOVERY_SUCCEEDED = "machine.recovery_succeeded"
    MACHINE_RECOVERY_FAILED = "machine.recovery_failed"

    # Request events
    REQUEST_STARTED = "request.started"
    REQUEST_COMPLETED = "request.completed"
    REQUEST_FAILED = "request.failed"
    REQUEST_TIMEOUT = "request.timeout"

    # Client events
    CLIENT_STARTED = "client.started"
    CLIENT_STOPPED = "client.stopped"
    CLIENT_ERROR = "client.error"


@dataclass
class Event:
    """Represents an event in the system."""

    type: EventType
    source: str
    data: Dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=datetime.now)
    correlation_id: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert event to dictionary."""
        return {
            "type": self.type.value,
            "source": self.source,
            "data": self.data,
            "timestamp": self.timestamp.isoformat(),
            "correlation_id": self.correlation_id,
        }


@runtime_checkable
class IEventHandler(Protocol):
    """Protocol for event handlers."""

    def handle_event(self, event: Event) -> None:
        """Handle an event."""
        ...


class IEventEmitter(ABC):
    """Abstract interface for event emission."""

    @abstractmethod
    def emit(self, event: Event) -> None:
        """Emit an event."""
        pass

    @abstractmethod
    def subscribe(
        self,
        event_type: EventType,
        handler: IEventHandler,
        filter_func: Optional[Callable[[Event], bool]] = None,
    ) -> str:
        """Subscribe to events of a specific type."""
        pass

    @abstractmethod
    def unsubscribe(self, subscription_id: str) -> bool:
        """Unsubscribe from events."""
        pass

    @abstractmethod
    def get_subscribers(self, event_type: EventType) -> List[IEventHandler]:
        """Get subscribers for an event type."""
        pass


class EventBus(IEventEmitter):
    """Central event bus implementation with weak references."""

    def __init__(self):
        self._subscribers: Dict[EventType, Dict[str, weakref.ReferenceType]] = {}
        self._filters: Dict[str, Callable[[Event], bool]] = {}
        self._subscription_counter = 0

    def emit(self, event: Event) -> None:
        """Emit an event to all subscribers."""
        if event.type not in self._subscribers:
            return

        # Clean up dead references and call live handlers
        dead_refs = []
        for sub_id, handler_ref in self._subscribers[event.type].items():
            handler = handler_ref()
            if handler is None:
                dead_refs.append(sub_id)
                continue

            # Apply filter if present
            if sub_id in self._filters:
                if not self._filters[sub_id](event):
                    continue

            try:
                handler.handle_event(event)
            except Exception as e:
                # Log error but don't stop other handlers
                print(f"Error in event handler {sub_id}: {e}")

        # Clean up dead references
        for sub_id in dead_refs:
            self._remove_subscription(event.type, sub_id)

    def subscribe(
        self,
        event_type: EventType,
        handler: IEventHandler,
        filter_func: Optional[Callable[[Event], bool]] = None,
    ) -> str:
        """Subscribe to events of a specific type."""
        if event_type not in self._subscribers:
            self._subscribers[event_type] = {}

        # Generate unique subscription ID
        self._subscription_counter += 1
        sub_id = f"sub_{self._subscription_counter}"

        # Store weak reference to handler
        self._subscribers[event_type][sub_id] = weakref.ref(handler)

        # Store filter if provided
        if filter_func:
            self._filters[sub_id] = filter_func

        return sub_id

    def unsubscribe(self, subscription_id: str) -> bool:
        """Unsubscribe from events."""
        for event_type in self._subscribers:
            if subscription_id in self._subscribers[event_type]:
                self._remove_subscription(event_type, subscription_id)
                return True
        return False

    def get_subscribers(self, event_type: EventType) -> List[IEventHandler]:
        """Get live subscribers for an event type."""
        if event_type not in self._subscribers:
            return []

        subscribers = []
        dead_refs = []

        for sub_id, handler_ref in self._subscribers[event_type].items():
            handler = handler_ref()
            if handler is None:
                dead_refs.append(sub_id)
            else:
                subscribers.append(handler)

        # Clean up dead references
        for sub_id in dead_refs:
            self._remove_subscription(event_type, sub_id)

        return subscribers

    def _remove_subscription(self, event_type: EventType, sub_id: str) -> None:
        """Remove a subscription."""
        if event_type in self._subscribers and sub_id in self._subscribers[event_type]:
            del self._subscribers[event_type][sub_id]

        if sub_id in self._filters:
            del self._filters[sub_id]


class EventHandlerMixin:
    """Mixin to add event handling capabilities to classes."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._event_subscriptions: List[str] = []

    def subscribe_to_events(
        self,
        event_bus: IEventEmitter,
        event_types: List[EventType],
        filter_func: Optional[Callable[[Event], bool]] = None,
    ) -> None:
        """Subscribe to multiple event types."""
        for event_type in event_types:
            sub_id = event_bus.subscribe(event_type, self, filter_func)
            self._event_subscriptions.append(sub_id)

    def unsubscribe_from_events(self, event_bus: IEventEmitter) -> None:
        """Unsubscribe from all events."""
        for sub_id in self._event_subscriptions:
            event_bus.unsubscribe(sub_id)
        self._event_subscriptions.clear()

    def handle_event(self, event: Event) -> None:
        """Default event handler - override in subclasses."""
        pass


# Global event bus instance (singleton pattern)
_global_event_bus: Optional[EventBus] = None


def get_global_event_bus() -> EventBus:
    """Get the global event bus instance."""
    global _global_event_bus
    if _global_event_bus is None:
        _global_event_bus = EventBus()
    return _global_event_bus


class EventLogger(IEventHandler):
    """Event handler that logs events."""

    def __init__(self, log_func: Optional[Callable[[str], None]] = None):
        self.log_func = log_func or print

    def handle_event(self, event: Event) -> None:
        """Log the event."""
        self.log_func(
            f"Event: {event.type.value} from {event.source} "
            f"at {event.timestamp.isoformat()}"
        )


class EventMetrics(IEventHandler):
    """Event handler that collects metrics."""

    def __init__(self):
        self.event_counts: Dict[EventType, int] = {}
        self.error_counts: Dict[str, int] = {}

    def handle_event(self, event: Event) -> None:
        """Collect metrics from the event."""
        # Count events by type
        self.event_counts[event.type] = self.event_counts.get(event.type, 0) + 1

        # Count errors by source
        if "error" in event.type.value:
            source = event.source
            self.error_counts[source] = self.error_counts.get(source, 0) + 1

    def get_metrics(self) -> Dict[str, Any]:
        """Get collected metrics."""
        return {
            "event_counts": {
                et.value: count for et, count in self.event_counts.items()
            },
            "error_counts": self.error_counts.copy(),
            "total_events": sum(self.event_counts.values()),
            "total_errors": sum(self.error_counts.values()),
        }
