"""Factory for creating various strategy implementations."""

from enum import Enum
from typing import Type, TypeVar

from ..interfaces.connection_interface import (
    ConnectionStrategy,
    DirectConnectionStrategy,
    IConnectionStrategy,
    PooledConnectionStrategy,
)

T = TypeVar("T")


class ExecutionStrategy(Enum):
    """Execution strategy types."""

    SYNCHRONOUS = "synchronous"
    ASYNCHRONOUS = "asynchronous"
    PARALLEL = "parallel"
    STREAMING = "streaming"


class RetryStrategy(Enum):
    """Retry strategy types."""

    NONE = "none"
    FIXED_DELAY = "fixed_delay"
    EXPONENTIAL_BACKOFF = "exponential_backoff"
    LINEAR_BACKOFF = "linear_backoff"


class LoadBalancingStrategy(Enum):
    """Load balancing strategy types."""

    ROUND_ROBIN = "round_robin"
    LEAST_CONNECTIONS = "least_connections"
    WEIGHTED_ROUND_ROBIN = "weighted_round_robin"
    RANDOM = "random"


class StrategyFactory:
    """Factory for creating strategy implementations."""

    def __init__(self):
        self._connection_strategies = {
            ConnectionStrategy.DIRECT: DirectConnectionStrategy,
            ConnectionStrategy.POOLED: PooledConnectionStrategy,
        }
        self._execution_strategies = {}
        self._retry_strategies = {}
        self._load_balancing_strategies = {}

    def create_connection_strategy(
        self, strategy_type: ConnectionStrategy, **kwargs
    ) -> IConnectionStrategy:
        """Create a connection strategy."""
        if strategy_type not in self._connection_strategies:
            raise ValueError(f"Unknown connection strategy: {strategy_type}")

        strategy_class = self._connection_strategies[strategy_type]
        return strategy_class(**kwargs)

    def register_connection_strategy(
        self,
        strategy_type: ConnectionStrategy,
        strategy_class: Type[IConnectionStrategy],
    ) -> None:
        """Register a connection strategy."""
        self._connection_strategies[strategy_type] = strategy_class

    def create_execution_strategy(self, strategy_type: ExecutionStrategy, **kwargs):
        """Create an execution strategy."""
        if strategy_type == ExecutionStrategy.SYNCHRONOUS:
            return SynchronousExecutionStrategy(**kwargs)
        elif strategy_type == ExecutionStrategy.ASYNCHRONOUS:
            return AsynchronousExecutionStrategy(**kwargs)
        elif strategy_type == ExecutionStrategy.PARALLEL:
            return ParallelExecutionStrategy(**kwargs)
        elif strategy_type == ExecutionStrategy.STREAMING:
            return StreamingExecutionStrategy(**kwargs)
        else:
            raise ValueError(f"Unknown execution strategy: {strategy_type}")

    def create_retry_strategy(self, strategy_type: RetryStrategy, **kwargs):
        """Create a retry strategy."""
        if strategy_type == RetryStrategy.NONE:
            return NoRetryStrategy()
        elif strategy_type == RetryStrategy.FIXED_DELAY:
            return FixedDelayRetryStrategy(**kwargs)
        elif strategy_type == RetryStrategy.EXPONENTIAL_BACKOFF:
            return ExponentialBackoffRetryStrategy(**kwargs)
        elif strategy_type == RetryStrategy.LINEAR_BACKOFF:
            return LinearBackoffRetryStrategy(**kwargs)
        else:
            raise ValueError(f"Unknown retry strategy: {strategy_type}")

    def create_load_balancing_strategy(
        self, strategy_type: LoadBalancingStrategy, **kwargs
    ):
        """Create a load balancing strategy."""
        if strategy_type == LoadBalancingStrategy.ROUND_ROBIN:
            return RoundRobinStrategy(**kwargs)
        elif strategy_type == LoadBalancingStrategy.LEAST_CONNECTIONS:
            return LeastConnectionsStrategy(**kwargs)
        elif strategy_type == LoadBalancingStrategy.WEIGHTED_ROUND_ROBIN:
            return WeightedRoundRobinStrategy(**kwargs)
        elif strategy_type == LoadBalancingStrategy.RANDOM:
            return RandomStrategy(**kwargs)
        else:
            raise ValueError(f"Unknown load balancing strategy: {strategy_type}")


# Execution Strategy Implementations
class SynchronousExecutionStrategy:
    """Synchronous execution strategy."""

    def __init__(self, timeout: int = 30):
        self.timeout = timeout

    def execute(self, func, *args, **kwargs):
        """Execute function synchronously."""
        return func(*args, **kwargs)


class AsynchronousExecutionStrategy:
    """Asynchronous execution strategy."""

    def __init__(self, max_concurrent: int = 10):
        self.max_concurrent = max_concurrent
        self._semaphore = None

    def execute(self, func, *args, **kwargs):
        """Execute function asynchronously."""
        import asyncio
        import concurrent.futures

        # Initialize semaphore if needed
        if self._semaphore is None:
            self._semaphore = asyncio.Semaphore(self.max_concurrent)

        async def async_wrapper():
            async with self._semaphore:
                loop = asyncio.get_event_loop()
                with concurrent.futures.ThreadPoolExecutor() as executor:
                    return await loop.run_in_executor(executor, func, *args, **kwargs)

        return asyncio.create_task(async_wrapper())


class ParallelExecutionStrategy:
    """Parallel execution strategy."""

    def __init__(self, max_workers: int = 5):
        self.max_workers = max_workers

    def execute(self, tasks):
        """Execute multiple tasks in parallel."""
        import concurrent.futures

        with concurrent.futures.ThreadPoolExecutor(
            max_workers=self.max_workers
        ) as executor:
            futures = []
            for func, args, kwargs in tasks:
                future = executor.submit(func, *args, **kwargs)
                futures.append(future)

            results = []
            for future in concurrent.futures.as_completed(futures):
                try:
                    result = future.result()
                    results.append(result)
                except Exception as e:
                    results.append(e)

            return results


class StreamingExecutionStrategy:
    """Streaming execution strategy."""

    def __init__(self, chunk_size: int = 1024, buffer_size: int = 10):
        self.chunk_size = chunk_size
        self.buffer_size = buffer_size

    def execute(self, func, *args, **kwargs):
        """Execute function with streaming results."""

        def stream_generator():
            try:
                result = func(*args, **kwargs)

                # If result is iterable, yield chunks
                if hasattr(result, "__iter__") and not isinstance(result, (str, bytes)):
                    for item in result:
                        yield item
                else:
                    # For non-iterable results, yield the result
                    yield result
            except Exception as e:
                yield e

        return stream_generator()


# Retry Strategy Implementations
class NoRetryStrategy:
    """No retry strategy."""

    def execute(self, func, *args, **kwargs):
        """Execute without retries."""
        return func(*args, **kwargs)


class FixedDelayRetryStrategy:
    """Fixed delay retry strategy."""

    def __init__(self, max_retries: int = 3, delay: float = 1.0):
        self.max_retries = max_retries
        self.delay = delay

    def execute(self, func, *args, **kwargs):
        """Execute with fixed delay retries."""
        import time

        last_exception = None
        for attempt in range(self.max_retries + 1):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                last_exception = e
                if attempt < self.max_retries:
                    time.sleep(self.delay)
                else:
                    break

        raise last_exception


class ExponentialBackoffRetryStrategy:
    """Exponential backoff retry strategy."""

    def __init__(
        self, max_retries: int = 3, base_delay: float = 1.0, max_delay: float = 60.0
    ):
        self.max_retries = max_retries
        self.base_delay = base_delay
        self.max_delay = max_delay

    def execute(self, func, *args, **kwargs):
        """Execute with exponential backoff retries."""
        import time

        last_exception = None
        for attempt in range(self.max_retries + 1):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                last_exception = e
                if attempt < self.max_retries:
                    delay = min(self.base_delay * (2**attempt), self.max_delay)
                    time.sleep(delay)
                else:
                    break

        raise last_exception


class LinearBackoffRetryStrategy:
    """Linear backoff retry strategy."""

    def __init__(
        self, max_retries: int = 3, base_delay: float = 1.0, increment: float = 1.0
    ):
        self.max_retries = max_retries
        self.base_delay = base_delay
        self.increment = increment

    def execute(self, func, *args, **kwargs):
        """Execute with linear backoff retries."""
        import time

        last_exception = None
        for attempt in range(self.max_retries + 1):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                last_exception = e
                if attempt < self.max_retries:
                    delay = self.base_delay + (attempt * self.increment)
                    time.sleep(delay)
                else:
                    break

        raise last_exception


# Load Balancing Strategy Implementations
class RoundRobinStrategy:
    """Round robin load balancing strategy."""

    def __init__(self):
        self._current = 0

    def select_target(self, targets):
        """Select target using round robin."""
        if not targets:
            return None

        target = targets[self._current % len(targets)]
        self._current += 1
        return target


class LeastConnectionsStrategy:
    """Least connections load balancing strategy."""

    def __init__(self):
        self._connections = {}

    def select_target(self, targets):
        """Select target with least connections."""
        if not targets:
            return None

        # Initialize connection counts if needed
        for target in targets:
            if target not in self._connections:
                self._connections[target] = 0

        # Select target with minimum connections
        return min(targets, key=lambda t: self._connections.get(t, 0))

    def record_connection(self, target, increment=True):
        """Record connection to target."""
        if target in self._connections:
            if increment:
                self._connections[target] += 1
            else:
                self._connections[target] = max(0, self._connections[target] - 1)


class WeightedRoundRobinStrategy:
    """Weighted round robin load balancing strategy."""

    def __init__(self, weights=None):
        self.weights = weights or {}
        self._current = 0
        self._weighted_targets = []

    def set_weights(self, weights):
        """Set target weights."""
        self.weights = weights
        self._build_weighted_list()

    def _build_weighted_list(self):
        """Build weighted target list."""
        self._weighted_targets = []
        for target, weight in self.weights.items():
            self._weighted_targets.extend([target] * weight)

    def select_target(self, targets):
        """Select target using weighted round robin."""
        if not self._weighted_targets:
            self._build_weighted_list()

        if not self._weighted_targets:
            return targets[0] if targets else None

        target = self._weighted_targets[self._current % len(self._weighted_targets)]
        self._current += 1
        return target


class RandomStrategy:
    """Random load balancing strategy."""

    def select_target(self, targets):
        """Select target randomly."""
        if not targets:
            return None

        import random

        return random.choice(targets)


# Global strategy factory instance
_global_strategy_factory = None


def get_strategy_factory() -> StrategyFactory:
    """Get global strategy factory instance."""
    global _global_strategy_factory
    if _global_strategy_factory is None:
        _global_strategy_factory = StrategyFactory()
    return _global_strategy_factory
