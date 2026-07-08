import { formatDateTime } from "../../utils/date.js";
import { getNextWindow } from "../../utils/matches.js";
import { MatchTeams } from "./MatchTeams.jsx";
import { ScoreButton } from "./ScoreButton.jsx";

export function NextMatchesSection({ matches, onNavigate }) {
  const nextMatches = getNextWindow(matches);

  if (!nextMatches.length) {
    return (
      <section className="next-panel">
        <div className="empty-state">Chưa có trận kế tiếp.</div>
      </section>
    );
  }

  return (
    <section className="next-panel">
      <div className="panel-heading row-heading compact-heading">
        <div>
          <p className="eyebrow">Next Match</p>
          <h2>Trận kế tiếp</h2>
        </div>
      </div>
      <div className={`next-grid count-${nextMatches.length}`}>
        {nextMatches.map((match) => (
          <NextMatchCard key={match.id} match={match} onNavigate={onNavigate} />
        ))}
      </div>
    </section>
  );
}

function NextMatchCard({ match, onNavigate }) {
  return (
    <article className="next-card">
      <div className="next-time">{formatDateTime(match.kickoffUTC)}</div>
      <MatchTeams match={match} onNavigate={onNavigate} />
      <div className="next-score">
        <ScoreButton match={match} onNavigate={onNavigate} />
      </div>
      <div className="hero-meta">
        <span>{match.group || match.stage}</span>
        <span>{match.matchday ? `Matchday ${match.matchday}` : "World Cup"}</span>
      </div>
    </article>
  );
}
