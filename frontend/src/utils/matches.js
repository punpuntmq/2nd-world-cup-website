import { parseDate } from "./date.js";

export function getNextWindow(matches) {
  const upcoming = (matches || [])
    .filter((match) => match.statusGroup === "upcoming" && parseDate(match.kickoffUTC))
    .sort((a, b) => parseDate(a.kickoffUTC) - parseDate(b.kickoffUTC));

  if (!upcoming.length) return [];

  const firstTime = parseDate(upcoming[0].kickoffUTC).getTime();
  const sixHours = 6 * 60 * 60 * 1000;

  return upcoming
    .filter((match) => Math.abs(parseDate(match.kickoffUTC).getTime() - firstTime) <= sixHours)
    .slice(0, 2);
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
