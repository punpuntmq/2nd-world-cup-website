import { ChevronRight } from "lucide-react";
import { Crest } from "../common/Crest.jsx";

export function TeamsPanel({ teams, onNavigate }) {
  return (
    <aside className="panel teams-panel" aria-label="Danh sách đội bóng">
      <div className="panel-heading">
        <p className="eyebrow">Teams</p>
        <h1>FIFA World Cup</h1>
      </div>
      <div className="team-list">
        {teams.length ? (
          teams.map((team) => (
            <button
              key={team.id}
              type="button"
              className="team-button"
              onClick={() => onNavigate(`/team/${team.id}`)}
            >
              <Crest src={team.crest} fallback={team.tla} />
              <span>
                <span className="team-name">{team.shortName || team.name}</span>
                <span className="team-meta">{team.group || "World Cup"}</span>
              </span>
              <span className="open-chip" aria-hidden="true">
                <ChevronRight size={18} />
              </span>
            </button>
          ))
        ) : (
          <div className="empty-state">Chưa có đội bóng.</div>
        )}
      </div>
    </aside>
  );
}
