export function formatGoalDiff(value) {
  return value > 0 ? `+${value}` : `${value}`;
}

export function shortPosition(position = "") {
  return position
    .replace("Goalkeeper", "GK")
    .replace("Defender", "DEF")
    .replace("Defence", "DEF")
    .replace("Midfielder", "MID")
    .replace("Midfield", "MID")
    .replace("Attacker", "FWD")
    .replace("Forward", "FWD")
    .replace("Offence", "FWD");
}
