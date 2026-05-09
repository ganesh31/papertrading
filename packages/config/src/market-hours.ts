/**
 * NSE equity cash session model (Phase 1).
 * Align with `services/go/md/internal/marketstatus` + `infra/seed/holidays.json`.
 * Pass replay virtual time as `now` when driving UI from `GET /replay/status.virtualTime`.
 */

export type MarketSession = "PREOPEN" | "OPEN" | "CLOSED" | "POSTCLOSE";

export type MarketSegment = "NSE_EQ";

export type HolidaysFile = {
  nse_eq?: string[];
};

/** Parse `holidays.json` body (same schema as Go). */
export function parseNseEqHolidays(raw: unknown): Set<string> {
  if (!raw || typeof raw !== "object") throw new Error("holidays: expected JSON object");
  const arr = (raw as HolidaysFile).nse_eq;
  if (!Array.isArray(arr)) throw new Error("holidays: missing nse_eq array");
  return new Set(
    arr.filter((x): x is string => typeof x === "string").map((s) => s.trim()).filter(Boolean),
  );
}

function istCalendarDateKey(d: Date): string {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Kolkata" }).format(d);
}

function istWeekdayLong(d: Date): string {
  return new Intl.DateTimeFormat("en-US", { timeZone: "Asia/Kolkata", weekday: "long" }).format(d);
}

function istMinuteOfDay(d: Date): number {
  const parts = new Intl.DateTimeFormat("en-GB", {
    timeZone: "Asia/Kolkata",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).formatToParts(d);
  const hour = Number(parts.find((p) => p.type === "hour")?.value ?? "0");
  const minute = Number(parts.find((p) => p.type === "minute")?.value ?? "0");
  return hour * 60 + minute;
}

/**
 * @param now wall clock or replay virtual instant
 * @param holidays IST calendar dates `YYYY-MM-DD` when NSE_EQ is closed (plus weekends in logic)
 */
export function getSession(now: Date, segment: MarketSegment, holidays: ReadonlySet<string>): MarketSession {
  if (segment !== "NSE_EQ") return "CLOSED";
  const ymd = istCalendarDateKey(now);
  if (holidays.has(ymd)) return "CLOSED";

  const wd = istWeekdayLong(now);
  if (wd === "Saturday" || wd === "Sunday") return "CLOSED";

  const min = istMinuteOfDay(now);
  if (min >= 9 * 60 && min < 9 * 60 + 8) return "PREOPEN";
  if (min >= 9 * 60 + 8 && min < 9 * 60 + 15) return "CLOSED";
  if (min >= 9 * 60 + 15 && min < 15 * 60 + 30) return "OPEN";
  if (min >= 15 * 60 + 30 && min < 16 * 60) return "POSTCLOSE";
  return "CLOSED";
}
