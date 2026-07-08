import { formatDateTime } from "../../utils/date.js";
import { effectiveRefreshStatus } from "../../utils/status.js";

export function StatusBanner({ meta, error }) {
  const status = effectiveRefreshStatus(meta);

  if (status === "refreshing") {
    return <div className="status-banner refreshing">Backend đang cập nhật snapshot World Cup.</div>;
  }

  if (status === "rate_limited") {
    return (
      <div className="status-banner rate-limited">
        Đã chạm quota upstream. Lần refresh tiếp theo: {formatDateTime(meta?.nextAllowedRefreshAt)}.
      </div>
    );
  }

  if (status === "error" || meta?.lastError || error) {
    return (
      <div className="status-banner error">
        Refresh lỗi: {meta?.lastError || error}. Dashboard vẫn dùng dữ liệu gần nhất.
      </div>
    );
  }

  if (meta?.isStale || status === "stale") {
    return <div className="status-banner stale">Dữ liệu đang cũ; dashboard vẫn hiển thị snapshot gần nhất từ backend.</div>;
  }

  return null;
}
