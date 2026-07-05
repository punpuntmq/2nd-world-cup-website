import { formatDateTime } from "../../utils/date.js";
import {
  effectiveRefreshStatus,
  isFuture,
  statusClass,
  statusLabel,
} from "../../utils/status.js";

export function Topbar({ meta, error, onNavigate }) {
  const status = effectiveRefreshStatus(meta);

  return (
    <header className="topbar">
      <button className="brand" type="button" onClick={() => onNavigate("/")}>
        <span className="brand-mark">WC</span>
        <span className="brand-text">World Cup</span>
      </button>
      <div className="topbar-status">
        <span className="source-tag live">API</span>
        <span className={`sync-pill ${statusClass(status)}`}>{statusLabel(status)}</span>
        <span className="quota-pill">
          Quota {meta?.remainingCalls ?? "-"} / {meta?.quotaLimit ?? "-"}
        </span>
        <span>{error ? `${formatDateTime(meta?.fetchedAt)} · fallback` : formatDateTime(meta?.fetchedAt)}</span>
      </div>
    </header>
  );
}
