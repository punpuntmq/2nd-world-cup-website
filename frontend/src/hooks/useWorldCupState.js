import { useEffect, useState } from "react";
import { fetchWorldCupState, refreshWorldCupState } from "../api/worldCupApi.js";
import { pollDelayMs } from "../utils/status.js";

export function useWorldCupState() {
  const [data, setData] = useState(null);
  const [error, setError] = useState("");
  const [activeGroup, setActiveGroup] = useState("");
  const [scheduleFilter, setScheduleFilter] = useState("all");
  const [isRefreshing, setRefreshing] = useState(false);

  useEffect(() => {
    let alive = true;
    let timer = 0;

    const load = async () => {
      try {
        const nextData = await fetchWorldCupState();
        if (!alive) return;

        setData(nextData);
        setError("");
        setActiveGroup((current) => current || nextData.standings?.[0]?.group || "");
        timer = window.setTimeout(load, pollDelayMs(nextData.meta));
      } catch (err) {
        if (!alive) return;

        setError(err.message);
        timer = window.setTimeout(load, 30000);
      }
    };

    load();
    return () => {
      alive = false;
      window.clearTimeout(timer);
    };
  }, []);

  const refreshNow = async () => {
    setRefreshing(true);
    try {
      const payload = await refreshWorldCupState();
      setData(payload.state);
      setError(payload.error || "");
    } catch (err) {
      setError(err.message);
    } finally {
      setRefreshing(false);
    }
  };

  return {
    data,
    error,
    activeGroup,
    scheduleFilter,
    isRefreshing,
    refreshNow,
    setActiveGroup,
    setScheduleFilter,
  };
}
