import { useEffect, useState } from "react";
import { fetchWorldCupState, refreshWorldCupState, subscribeWorldCupEvents } from "../api/worldCupApi.js";

export function useWorldCupState() {
  const [data, setData] = useState(null);
  const [error, setError] = useState("");
  const [activeGroup, setActiveGroup] = useState("");
  const [scheduleFilter, setScheduleFilter] = useState("all");
  const [isRefreshing, setRefreshing] = useState(false);

  useEffect(() => {
    let alive = true;
    let openedOnce = false;

    const applyData = (nextData) => {
      setData(nextData);
      setError("");
      setActiveGroup((current) => current || nextData.standings?.[0]?.group || "");
    };

    const load = async () => {
      try {
        const nextData = await fetchWorldCupState();
        if (!alive) return;

        applyData(nextData);
      } catch (err) {
        if (!alive) return;

        setError(err.message);
      }
    };

    load();
    const unsubscribe = subscribeWorldCupEvents({
      onOpen: () => {
        if (!alive) return;
        if (openedOnce) {
          load();
        }
        openedOnce = true;
      },
      onState: (nextData) => {
        if (!alive) return;
        applyData(nextData);
      },
      onError: () => {
        if (!alive) return;
        setError("Mất kết nối realtime, đang kết nối lại...");
      },
    });

    return () => {
      alive = false;
      unsubscribe();
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
