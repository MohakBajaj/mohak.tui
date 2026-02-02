/**
 * PostHog analytics integration for AI Gateway
 */

import { PostHog } from "posthog-node";
import { logger } from "./logger";

// Initialize PostHog client
const POSTHOG_API_KEY = Bun.env.POSTHOG_API_KEY;
const POSTHOG_HOST = Bun.env.POSTHOG_HOST ?? "https://us.i.posthog.com";

let posthogClient: PostHog | null = null;

if (POSTHOG_API_KEY) {
  posthogClient = new PostHog(POSTHOG_API_KEY, {
    host: POSTHOG_HOST,
    flushAt: 10,
    flushInterval: 5000,
  });
  logger.info("PostHog analytics initialized", { host: POSTHOG_HOST });
} else {
  logger.warn("PostHog API key not set, analytics disabled");
}

// Event types for type safety
export type AnalyticsEvent =
  | "chat_request"
  | "chat_response"
  | "chat_error"
  | "rate_limit_hit"
  | "health_check"
  | "server_start"
  | "server_stop";

interface ChatRequestProperties {
  sessionId: string;
  messageLength: number;
  historyLength: number;
  model: string;
}

interface ChatResponseProperties {
  sessionId: string;
  durationMs: number;
  model: string;
  success: boolean;
}

interface ChatErrorProperties {
  sessionId: string;
  error: string;
  errorType: string;
}

interface RateLimitProperties {
  sessionId: string;
  remaining: number;
}

type EventProperties =
  | ChatRequestProperties
  | ChatResponseProperties
  | ChatErrorProperties
  | RateLimitProperties
  | Record<string, unknown>;

export const analytics = {
  /**
   * Capture an analytics event
   */
  capture(
    event: AnalyticsEvent,
    distinctId: string,
    properties?: EventProperties,
  ) {
    if (!posthogClient) return;

    try {
      posthogClient.capture({
        distinctId,
        event: `ai_gateway_${event}`,
        properties: {
          ...properties,
          service: "ai-gateway",
          environment: Bun.env.NODE_ENV ?? "development",
        },
      });
    } catch (error) {
      logger.error("Failed to capture analytics event", {
        event,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  },

  /**
   * Track a chat request
   */
  trackChatRequest(props: ChatRequestProperties) {
    this.capture("chat_request", props.sessionId, props);
  },

  /**
   * Track a chat response
   */
  trackChatResponse(props: ChatResponseProperties) {
    this.capture("chat_response", props.sessionId, props);
  },

  /**
   * Track a chat error
   */
  trackChatError(props: ChatErrorProperties) {
    this.capture("chat_error", props.sessionId, props);
  },

  /**
   * Track rate limit hit
   */
  trackRateLimit(props: RateLimitProperties) {
    this.capture("rate_limit_hit", props.sessionId, props);
  },

  /**
   * Track server lifecycle
   */
  trackServerStart() {
    this.capture("server_start", "system", {
      port: Bun.env.AI_GATEWAY_PORT ?? "3001",
      model: Bun.env.AI_GATEWAY_MODEL ?? "openai/gpt-oss-20b",
    });
  },

  trackServerStop() {
    this.capture("server_stop", "system", {});
  },

  /**
   * Identify a user/session
   */
  identify(distinctId: string, properties: Record<string, unknown>) {
    if (!posthogClient) return;

    try {
      posthogClient.capture({
        distinctId,
        event: "$identify",
        properties: {
          $set: properties,
        },
      });
    } catch (error) {
      logger.error("Failed to identify user", {
        distinctId,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  },

  /**
   * Flush pending events and shutdown
   */
  async shutdown() {
    if (!posthogClient) return;

    try {
      await posthogClient.shutdown();
      logger.info("PostHog client shutdown complete");
    } catch (error) {
      logger.error("Error shutting down PostHog client", {
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  },
};

export type Analytics = typeof analytics;
