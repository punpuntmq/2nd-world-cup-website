import { RefreshCw } from "lucide-react";
import { formatDateTime } from "../../utils/date.js";
import {
  effectiveRefreshStatus,
  isFuture,
  refreshTitle,
  statusClass,
  statusLabel,
} from "../../utils/status.js";

export function Topbar({ meta, error, isRefreshing, onRefresh, onNavigate }) {
  const status = effectiveRefreshStatus(meta);
  const rateLimited = status === "rate_limited" && isFuture(meta?.nextAllowedRefreshAt);
  const disabled = isRefreshing || status === "refreshing" || rateLimited;

  return (
    <header className="topbar">
      <button className="brand" type="button" onClick={() => onNavigate("/")}>
        <span className="brand-mark">WC</span>
        <span className="brand-text">World Cup</span>
      </button>
      <div className="topbar-status">
        <span className={`source-tag ${meta?.source === "football-data" ? "live" : ""}`}>
          {meta?.source === "football-data" ? "API" : "FAKE"}
        </span>
        <span className={`sync-pill ${statusClass(status)}`}>{statusLabel(status)}</span>
        <span className="quota-pill">
          Quota {meta?.remainingCalls ?? "-"} / {meta?.quotaLimit ?? "-"}
        </span>
        <span>{error ? `${formatDateTime(meta?.fetchedAt)} · fallback` : formatDateTime(meta?.fetchedAt)}</span>
        <button
          className="icon-button"
          type="button"
          title={refreshTitle(meta)}
          disabled={disabled}
          onClick={onRefresh}
          aria-label="Cập nhật dữ liệu"
        >
          <RefreshCw size={18} aria-hidden="true" />
        </button>
      </div>
    </header>
  );
}
