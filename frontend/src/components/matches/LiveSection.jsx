import { MatchTeams } from "./MatchTeams.jsx";
import { ScoreButton } from "./ScoreButton.jsx";

export function LiveSection({ matches, onNavigate }) {
  if (!matches.length) return null;

  return (
    <section className="live-panel">
      <div className="panel-heading row-heading compact-heading">
        <div>
          <p className="eyebrow">Live</p>
          <h2>Đang trực tiếp</h2>
        </div>
      </div>
      <div className="live-grid">
        {matches.map((match) => (
          <LiveCard key={match.id} match={match} onNavigate={onNavigate} />
        ))}
      </div>
    </section>
  );
}

function LiveCard({ match, onNavigate }) {
  return (
    <article className="live-card">
      <div className="match-topline">
        <span className="status-pill live">{match.minute || match.statusLabel}</span>
        <span>{match.group || match.stage}</span>
      </div>
      <MatchTeams match={match} onNavigate={onNavigate} large />
      <ScoreButton match={match} onNavigate={onNavigate} large />
    </article>
  );
}
