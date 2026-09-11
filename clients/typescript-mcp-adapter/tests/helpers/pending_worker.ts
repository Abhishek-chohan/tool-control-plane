/**
 * Claims pending requests for a session through a provider adapter and
 * completes them, standing in for the provider machine inside tests.
 */
import type {GrpcConformanceAdapter} from '../../../typescript-client/tests/conformance/adapters/grpc_adapter';

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export class PendingRequestWorker {
  private stopped = false;

  private loopPromise?: Promise<void>;

  private readonly inFlight = new Set<string>();

  constructor(
    private readonly provider: GrpcConformanceAdapter,
    private readonly sessionId: string,
  ) {}

  start(): void {
    if (!this.loopPromise) {
      this.loopPromise = this.loop();
    }
  }

  async stop(): Promise<void> {
    this.stopped = true;

    if (this.loopPromise) {
      await this.loopPromise;
    }

    while (this.inFlight.size > 0) {
      await sleep(25);
    }
  }

  private async loop(): Promise<void> {
    while (!this.stopped) {
      const pendingRequests = await this.provider.listRequests(this.sessionId, {
        list_status: 'pending',
        limit: 20,
      });

      for (const request of pendingRequests) {
        const requestId = String(request.id ?? '');
        if (!requestId || this.inFlight.has(requestId)) {
          continue;
        }

        this.inFlight.add(requestId);
        void this.provider
          .startRequestProcessing(this.sessionId, requestId)
          .finally(() => {
            this.inFlight.delete(requestId);
          });
      }

      await sleep(50);
    }
  }
}
