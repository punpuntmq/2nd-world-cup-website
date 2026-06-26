export async function fetchWorldCupState() {
  const response = await fetch("/api/state", { cache: "no-store" });
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

export async function refreshWorldCupState() {
  const response = await fetch("/api/refresh", { method: "POST" });
  return response.json();
}
