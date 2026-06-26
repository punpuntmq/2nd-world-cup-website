import { parseDate } from "./date.js";

export function getNextWindow(matches) {
  const upcoming = (matches || []).filter((match) => match.statusGroup === "upcoming");
  return upcoming.slice(0, 2);
}

export function getAllMatches(data) {
  return [
    ...(data.liveMatches || []),
    ...(data.upcomingMatches || []),
    ...(data.finishedMatches || []),
    ...(data.specialMatches || []),
  ];
}

export function findMatch(data, id) {
  return getAllMatches(data).find((match) => match.id === id) || null;
}
