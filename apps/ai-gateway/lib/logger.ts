/**
 * Production-grade structured logger for the AI Gateway
 */

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogContext {
  [key: string]: unknown;
}

interface LogEntry {
  timestamp: string;
  level: LogLevel;
  message: string;
  service: string;
  context?: LogContext;
}

const LOG_LEVELS: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

const currentLevel: LogLevel = (Bun.env.LOG_LEVEL as LogLevel) ?? "info";

function shouldLog(level: LogLevel): boolean {
  return LOG_LEVELS[level] >= LOG_LEVELS[currentLevel];
}

function formatLog(entry: LogEntry): string {
  const { timestamp, level, message, service, context } = entry;

  if (Bun.env.LOG_FORMAT === "json") {
    return JSON.stringify(entry);
  }

  // Pretty format for development
  const levelColors: Record<LogLevel, string> = {
    debug: "\x1b[36m", // cyan
    info: "\x1b[32m", // green
    warn: "\x1b[33m", // yellow
    error: "\x1b[31m", // red
  };
  const reset = "\x1b[0m";
  const dim = "\x1b[2m";

  let output = `${dim}${timestamp}${reset} ${levelColors[level]}${level.toUpperCase().padEnd(5)}${reset} ${dim}[${service}]${reset} ${message}`;

  if (context && Object.keys(context).length > 0) {
    output += ` ${dim}${JSON.stringify(context)}${reset}`;
  }

  return output;
}

function createLogEntry(
  level: LogLevel,
  message: string,
  context?: LogContext,
): LogEntry {
  return {
    timestamp: new Date().toISOString(),
    level,
    message,
    service: "ai-gateway",
    context,
  };
}

export const logger = {
  debug(message: string, context?: LogContext) {
    if (shouldLog("debug")) {
      console.log(formatLog(createLogEntry("debug", message, context)));
    }
  },

  info(message: string, context?: LogContext) {
    if (shouldLog("info")) {
      console.log(formatLog(createLogEntry("info", message, context)));
    }
  },

  warn(message: string, context?: LogContext) {
    if (shouldLog("warn")) {
      console.warn(formatLog(createLogEntry("warn", message, context)));
    }
  },

  error(message: string, context?: LogContext) {
    if (shouldLog("error")) {
      console.error(formatLog(createLogEntry("error", message, context)));
    }
  },

  // Log with timing
  time(label: string): () => void {
    const start = performance.now();
    return () => {
      const duration = performance.now() - start;
      this.debug(`${label} completed`, { durationMs: Math.round(duration) });
    };
  },

  // Log request info
  request(
    method: string,
    path: string,
    statusCode: number,
    durationMs: number,
    context?: LogContext,
  ) {
    const level: LogLevel =
      statusCode >= 500 ? "error" : statusCode >= 400 ? "warn" : "info";
    this[level](`${method} ${path} ${statusCode}`, {
      ...context,
      statusCode,
      durationMs: Math.round(durationMs),
    });
  },
};

export type Logger = typeof logger;
