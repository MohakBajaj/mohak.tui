import { Hono } from "hono";
import { cors } from "hono/cors";
import { streamText } from "ai";
import { gateway } from "@ai-sdk/gateway";
import {
  buildSystemPrompt,
  preprocessMessage,
  detectQueryIntent,
  generateFollowUps,
} from "@mohak/shared-content";
import { logger } from "./lib/logger";
import { analytics } from "./lib/analytics";

const app = new Hono();

// Configuration
const PORT = parseInt(Bun.env.AI_GATEWAY_PORT ?? "3001", 10);
const MODEL = Bun.env.AI_GATEWAY_MODEL ?? "openai/gpt-oss-20b";
const MAX_TOKENS = parseInt(Bun.env.AI_GATEWAY_MAX_TOKENS ?? "1024", 10);
const TEMPERATURE = parseFloat(Bun.env.AI_TEMPERATURE ?? "0.7");
const MAX_HISTORY_LENGTH = 10; // Keep last N messages for context

// Rate limiting
interface RateLimitEntry {
  count: number;
  resetAt: number;
}

const rateLimitStore = new Map<string, RateLimitEntry>();
const RATE_LIMIT_WINDOW = 60 * 1000;
const RATE_LIMIT_MAX = parseInt(Bun.env.AI_GATEWAY_RATE_LIMIT ?? "10", 10);

function checkRateLimit(sessionId: string): {
  allowed: boolean;
  remaining: number;
} {
  const now = Date.now();
  const entry = rateLimitStore.get(sessionId);

  if (!entry || now > entry.resetAt) {
    rateLimitStore.set(sessionId, {
      count: 1,
      resetAt: now + RATE_LIMIT_WINDOW,
    });
    return { allowed: true, remaining: RATE_LIMIT_MAX - 1 };
  }

  if (entry.count >= RATE_LIMIT_MAX) {
    return { allowed: false, remaining: 0 };
  }

  entry.count++;
  return { allowed: true, remaining: RATE_LIMIT_MAX - entry.count };
}

// Cleanup old rate limit entries
setInterval(() => {
  const now = Date.now();
  let cleaned = 0;
  for (const [key, entry] of rateLimitStore.entries()) {
    if (now > entry.resetAt) {
      rateLimitStore.delete(key);
      cleaned++;
    }
  }
  if (cleaned > 0) {
    logger.debug("Cleaned rate limit entries", { count: cleaned });
  }
}, 60 * 1000);

// Middleware
app.use("*", cors());

app.use("*", async (c, next) => {
  const start = performance.now();
  await next();
  const duration = performance.now() - start;
  logger.request(c.req.method, c.req.path, c.res.status, duration, {
    userAgent: c.req.header("user-agent"),
  });
});

// Health check
app.get("/health", (c) => {
  logger.debug("Health check requested");
  return c.json({
    status: "ok",
    model: MODEL,
    temperature: TEMPERATURE,
    maxTokens: MAX_TOKENS,
    timestamp: new Date().toISOString(),
  });
});

// Chat endpoint
app.post("/chat", async (c) => {
  const requestStart = performance.now();

  try {
    const body = await c.req.json<{
      message: string;
      sessionId?: string;
      history?: { role: "user" | "assistant"; content: string }[];
    }>();

    const { message: rawMessage, sessionId = "anonymous", history = [] } = body;

    // Validate input
    if (!rawMessage || typeof rawMessage !== "string") {
      logger.warn("Invalid chat request - missing message", { sessionId });
      return c.json({ error: "Message is required" }, 400);
    }

    if (rawMessage.length > 2000) {
      logger.warn("Message too long", { sessionId, length: rawMessage.length });
      return c.json({ error: "Message too long (max 2000 characters)" }, 400);
    }

    // Preprocess message
    const message = preprocessMessage(rawMessage);
    const intent = detectQueryIntent(message);

    // Track request
    analytics.trackChatRequest({
      sessionId,
      messageLength: message.length,
      historyLength: history.length,
      model: MODEL,
    });

    logger.info("Chat request received", {
      sessionId,
      messageLength: message.length,
      historyLength: history.length,
      intent,
    });

    // Rate limit check
    const rateLimit = checkRateLimit(sessionId);
    if (!rateLimit.allowed) {
      logger.warn("Rate limit exceeded", { sessionId });
      analytics.trackRateLimit({ sessionId, remaining: 0 });
      return c.json(
        {
          error:
            "Rate limit exceeded. Please wait before sending more messages.",
          retryAfter: Math.ceil(RATE_LIMIT_WINDOW / 1000),
        },
        429,
      );
    }

    // Trim history to prevent context overflow
    const trimmedHistory = history.slice(-MAX_HISTORY_LENGTH);

    // Build dynamic system prompt based on query intent
    const systemPrompt = buildSystemPrompt(message);

    // Build messages array
    const messages: {
      role: "user" | "assistant" | "system";
      content: string;
    }[] = [
      { role: "system", content: systemPrompt },
      ...trimmedHistory,
      { role: "user", content: message },
    ];

    logger.debug("Streaming AI response", {
      sessionId,
      model: MODEL,
      intent,
      temperature: TEMPERATURE,
      historyLength: trimmedHistory.length,
    });

    // Stream response with advanced settings
    const result = streamText({
      model: gateway(MODEL),
      messages,
      maxOutputTokens: MAX_TOKENS,
      temperature: TEMPERATURE,
      topP: 0.9,
      frequencyPenalty: 0.3, // Reduce repetition
      presencePenalty: 0.1, // Encourage topic diversity
      stopSequences: [
        "\n\n\n", // Prevent excessive newlines
        "---END---",
        "Human:",
        "User:",
      ],
      onFinish: async ({ text, usage, finishReason }) => {
        const duration = performance.now() - requestStart;

        analytics.trackChatResponse({
          sessionId,
          durationMs: Math.round(duration),
          model: MODEL,
          success: true,
        });

        logger.info("Chat response completed", {
          sessionId,
          durationMs: Math.round(duration),
          responseLength: text?.length ?? 0,
          finishReason,
          tokenUsage: usage.raw,
          intent,
        });
      },
    });

    // Set streaming headers
    c.header("Content-Type", "text/event-stream");
    c.header("Cache-Control", "no-cache");
    c.header("Connection", "keep-alive");
    c.header("X-RateLimit-Remaining", rateLimit.remaining.toString());
    c.header("X-Query-Intent", intent);

    return result.toTextStreamResponse();
  } catch (error) {
    const duration = performance.now() - requestStart;
    const errorMessage =
      error instanceof Error ? error.message : "Unknown error";
    const errorType =
      error instanceof Error ? error.constructor.name : "Unknown";

    logger.error("Chat error", {
      error: errorMessage,
      errorType,
      durationMs: Math.round(duration),
    });

    // Track error
    try {
      const body = await c.req.json();
      analytics.trackChatError({
        sessionId: body?.sessionId ?? "anonymous",
        error: errorMessage,
        errorType,
      });
    } catch {
      analytics.trackChatError({
        sessionId: "anonymous",
        error: errorMessage,
        errorType,
      });
    }

    if (error instanceof Error) {
      return c.json({ error: error.message }, 500);
    }

    return c.json({ error: "An unexpected error occurred" }, 500);
  }
});

// Follow-up suggestions endpoint
app.post("/suggestions", async (c) => {
  try {
    const body = await c.req.json<{ message: string }>();
    const { message } = body;

    if (!message) {
      return c.json({ suggestions: [] });
    }

    const intent = detectQueryIntent(message);
    const suggestions = generateFollowUps(intent);

    return c.json({ suggestions, intent });
  } catch (error) {
    logger.error("Suggestions error", {
      error: error instanceof Error ? error.message : "Unknown",
    });
    return c.json({ suggestions: [] });
  }
});

// Intent detection endpoint (for debugging/UI hints)
app.post("/intent", async (c) => {
  try {
    const body = await c.req.json<{ message: string }>();
    const { message } = body;

    if (!message) {
      return c.json({ intent: "general" });
    }

    const intent = detectQueryIntent(message);
    return c.json({ intent });
  } catch (error) {
    return c.json({ intent: "general" });
  }
});

// Start server
logger.info("AI Gateway starting", {
  port: PORT,
  model: MODEL,
  temperature: TEMPERATURE,
  maxTokens: MAX_TOKENS,
  rateLimit: RATE_LIMIT_MAX,
  maxHistory: MAX_HISTORY_LENGTH,
});
analytics.trackServerStart();

// Graceful shutdown
process.on("SIGTERM", async () => {
  logger.info("SIGTERM received, shutting down...");
  analytics.trackServerStop();
  await analytics.shutdown();
  process.exit(0);
});

process.on("SIGINT", async () => {
  logger.info("SIGINT received, shutting down...");
  analytics.trackServerStop();
  await analytics.shutdown();
  process.exit(0);
});

export default {
  port: PORT,
  fetch: app.fetch,
};
