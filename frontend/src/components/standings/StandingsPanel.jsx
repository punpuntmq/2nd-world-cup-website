import { Crest } from "../common/Crest.jsx";
import { formatGoalDiff } from "../../utils/team.js";

export function StandingsPanel({ groups, activeGroup, onGroupChange }) {
  const selected = groups.find((group) => group.group === activeGroup) || groups[0];

  return (
    <aside className="panel standings-panel" aria-label="Bảng xếp hạng">
      <div className="panel-heading">
        <p className="eyebrow">Standings</p>
        <h2>Group table</h2>
      </div>
      {groups.length ? (
        <>
          <div className="group-tabs">
            {groups.map((group) => (
              <button
                key={group.group}
                type="button"
                className={`group-tab ${group.group === selected.group ? "active" : ""}`}
                onClick={() => onGroupChange(group.group)}
              >
                <span>{group.group}</span>
                <strong>{group.memberCount ?? group.rows?.length ?? 0}</strong>
              </button>
            ))}
          </div>
          <div className="group-info">
            {selected.group} · {(selected.rows || []).length} đội trong group
          </div>
          <div className="table-wrap">
            <table className="standings-table">
              <thead>
                <tr>
                  {["#", "Team", "P", "W", "D", "L", "GD", "Pts"].map((label) => (
                    <th key={label}>{label}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(selected.rows || []).map((row) => (
                  <tr key={row.team.id}>
                    <td>{row.position}</td>
                    <td>
                      <span className="standing-team">
                        <Crest src={row.team.crest} fallback={row.team.tla} />
                        <strong>{row.team.tla || row.team.shortName}</strong>
                      </span>
                    </td>
                    <td>{row.playedGames}</td>
                    <td>{row.won}</td>
                    <td>{row.draw}</td>
                    <td>{row.lost}</td>
                    <td className={row.goalDifference >= 0 ? "gd-positive" : "gd-negative"}>
                      {formatGoalDiff(row.goalDifference)}
                    </td>
                    <td className="points-cell">{row.points}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : (
        <div className="empty-state">Chưa có bảng xếp hạng.</div>
      )}
    </aside>
  );
}
