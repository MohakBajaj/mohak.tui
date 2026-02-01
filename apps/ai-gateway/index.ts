import { Hono } from "hono";
import { cors } from "hono/cors";
import { streamText } from "ai";
import { gateway } from "@ai-sdk/gateway";
import { buildSystemPrompt } from "@mohak/shared-content";

const app = new Hono();

// Configuration
const PORT = parseInt(Bun.env.AI_GATEWAY_PORT ?? "3001", 10);
const MODEL = Bun.env.AI_GATEWAY_MODEL ?? "openai/gpt-oss-20b";
const MAX_TOKENS = parseInt(Bun.env.AI_GATEWAY_MAX_TOKENS ?? "1024", 10);

// Rate limiting: simple in-memory store
// In production, use Redis or similar
interface RateLimitEntry {
  count: number;
  resetAt: number;
}

const rateLimitStore = new Map<string, RateLimitEntry>();
const RATE_LIMIT_WINDOW = 60 * 1000; // 1 minute
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

// Clean up old rate limit entries periodically
setInterval(() => {
  const now = Date.now();
  for (const [key, entry] of rateLimitStore.entries()) {
    if (now > entry.resetAt) {
      rateLimitStore.delete(key);
    }
  }
}, 60 * 1000);

// Middleware
app.use("*", cors());

// Health check
app.get("/health", (c) => {
  return c.json({ status: "ok", model: MODEL });
});

// Chat endpoint
app.post("/chat", async (c) => {
  try {
    const body = await c.req.json<{
      message: string;
      sessionId?: string;
      history?: { role: "user" | "assistant"; content: string }[];
    }>();

    const { message, sessionId = "anonymous", history = [] } = body;

    if (!message || typeof message !== "string") {
      return c.json({ error: "Message is required" }, 400);
    }

    // Rate limit check
    const rateLimit = checkRateLimit(sessionId);
    if (!rateLimit.allowed) {
      return c.json(
        {
          error:
            "Rate limit exceeded. Please wait before sending more messages.",
          retryAfter: Math.ceil(RATE_LIMIT_WINDOW / 1000),
        },
        429,
      );
    }

    // Build messages array
    const messages: {
      role: "user" | "assistant" | "system";
      content: string;
    }[] = [
      { role: "system", content: buildSystemPrompt() },
      ...history,
      { role: "user", content: message },
    ];

    // Stream response
    const result = streamText({
      model: gateway(MODEL),
      messages,
      maxOutputTokens: MAX_TOKENS,
    });

    // Return streaming response
    c.header("Content-Type", "text/event-stream");
    c.header("Cache-Control", "no-cache");
    c.header("Connection", "keep-alive");
    c.header("X-RateLimit-Remaining", rateLimit.remaining.toString());

    return result.toTextStreamResponse();
  } catch (error) {
    console.error("Chat error:", error);

    if (error instanceof Error) {
      return c.json({ error: error.message }, 500);
    }

    return c.json({ error: "An unexpected error occurred" }, 500);
  }
});

// Start server
console.log(`🚀 AI Gateway starting on port ${PORT}`);
console.log(`📡 Model: ${MODEL}`);
console.log(`🔒 Rate limit: ${RATE_LIMIT_MAX} requests per minute`);

export default {
  port: PORT,
  fetch: app.fetch,
};
