import { Crest } from "../common/Crest.jsx";

export function TeamButton({ team, className, onNavigate }) {
  return (
    <button
      type="button"
      className={className}
      onClick={() => team?.id && onNavigate(`/team/${team.id}`)}
      disabled={!team?.id}
    >
      <Crest src={team?.crest} fallback={team?.tla} />
      <span className="row-team-name">{team?.shortName || team?.name || "TBD"}</span>
    </button>
  );
}
