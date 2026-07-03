const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || "").replace(/\/$/, "");

function apiURL(path) {
  return `${API_BASE_URL}${path}`;
}

export async function fetchWorldCupState() {
  const response = await fetch(apiURL("/api/state"), { cache: "no-store" });
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

export async function refreshWorldCupState() {
  const response = await fetch(apiURL("/api/refresh"), { method: "POST" });
  const payload = await response.json();
  if (!response.ok && !payload.state) throw new Error(payload.error || `HTTP ${response.status}`);
  return payload;
}

export function subscribeWorldCupEvents({ onOpen, onState, onError }) {
  const source = new EventSource(apiURL("/api/events"));

  source.onopen = () => {
    onOpen?.();
  };

  source.addEventListener("state", (event) => {
    try {
      onState?.(JSON.parse(event.data));
    } catch (err) {
      onError?.(err);
    }
  });

  source.onerror = () => {
    onError?.(new Error("Realtime connection lost"));
  };

  return () => source.close();
}
