import { formatDateTime, parseDate } from "./date.js";

export function pollDelayMs(meta) {
  const fallback = Math.max(10, meta?.refreshSeconds || 45) * 1000;
  const status = effectiveRefreshStatus(meta);

  if (status === "rate_limited") {
    const nextAllowed = parseDate(meta?.nextAllowedRefreshAt);
    if (nextAllowed && nextAllowed.getTime() > Date.now()) {
      return Math.max(10000, nextAllowed.getTime() - Date.now() + 1000);
    }
    return Math.max(fallback, 30000);
  }

  if (status === "refreshing") {
    return Math.min(fallback, 5000);
  }

  if (meta?.isStale || status === "stale") {
    return Math.max(fallback, 30000);
  }

  return fallback;
}

export function statusClass(status = "idle") {
  return status.replace(/_/g, "-");
}

export function statusLabel(status = "idle") {
  switch (status) {
    case "refreshing":
      return "Đang cập nhật";
    case "rate_limited":
      return "Hết quota";
    case "stale":
      return "Dữ liệu cũ";
    case "error":
      return "Lỗi refresh";
    default:
      return "Ổn định";
  }
}

export function refreshTitle(meta) {
  const status = effectiveRefreshStatus(meta);
  if (status === "refreshing") return "Backend đang cập nhật";
  if (status === "rate_limited" && isFuture(meta?.nextAllowedRefreshAt)) {
    return `Chờ tới ${formatDateTime(meta?.nextAllowedRefreshAt)}`;
  }
  return "Cập nhật ngay";
}

export function isFuture(value) {
  const date = parseDate(value);
  return Boolean(date && date.getTime() > Date.now());
}

export function effectiveRefreshStatus(meta) {
  const status = meta?.refreshStatus || "idle";
  if (status !== "rate_limited") return status;

  const nextAllowed = parseDate(meta?.nextAllowedRefreshAt);
  if (nextAllowed && nextAllowed.getTime() <= Date.now()) {
    return meta?.isStale ? "stale" : "idle";
  }

  return status;
}
