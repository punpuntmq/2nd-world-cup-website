import { TeamButton } from "../teams/TeamButton.jsx";

export function MatchTeams({ match, onNavigate, large }) {
  return (
    <div className={large ? "match-teams large" : "match-teams"}>
      <TeamButton team={match.homeTeam} className="match-team-link" onNavigate={onNavigate} />
      <span className="versus">vs</span>
      <TeamButton team={match.awayTeam} className="match-team-link" onNavigate={onNavigate} />
    </div>
  );
}
