import { formatShortDate } from "../../utils/date.js";
import { Crest } from "../common/Crest.jsx";
import { ScoreButton } from "./ScoreButton.jsx";

export function MatchRow({ match, onNavigate }) {
  return (
    <div className={`match-row ${match.statusGroup}`}>
      <div className="match-row-meta">
        <span className="row-stage">{match.group || match.stage || "World Cup"}</span>
        <span className="match-time">
          <strong>{match.statusLabel}</strong>
          <span>{formatShortDate(match.kickoffUTC)}</span>
        </span>
      </div>
      <div className="match-row-body">
        <RowTeam team={match.homeTeam} side="home" onNavigate={onNavigate} />
        <div className="match-center">
          <ScoreButton match={match} onNavigate={onNavigate} />
        </div>
        <RowTeam team={match.awayTeam} side="away" onNavigate={onNavigate} />
      </div>
    </div>
  );
}

function RowTeam({ team, side, onNavigate }) {
  const disabled = !team?.id;
  const name = team?.shortName || team?.name || "TBD";
  const crest = <Crest src={team?.crest} fallback={team?.tla} />;

  return (
    <button
      type="button"
      className={`row-team ${side}-team`}
      onClick={() => team?.id && onNavigate(`/team/${team.id}`)}
      disabled={disabled}
    >
      {side === "home" ? crest : null}
      <span className="row-team-name">{name}</span>
      {side === "away" ? crest : null}
    </button>
  );
}
