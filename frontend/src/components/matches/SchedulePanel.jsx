import { MatchRow } from "./MatchRow.jsx";

export function SchedulePanel({ data, filter, onFilterChange, onNavigate }) {
  const upcoming = data.upcomingMatches || [];
  const finished = data.finishedMatches || [];
  const special = data.specialMatches || [];
  const rows = filter === "upcoming" ? upcoming : filter === "finished" ? finished : [...upcoming, ...finished];

  return (
    <section className="schedule-panel">
      <div className="panel-heading row-heading">
        <div>
          <p className="eyebrow">Schedule</p>
          <h2>All matches</h2>
        </div>
        <div className="schedule-filters" role="tablist" aria-label="Lọc lịch đấu">
          <FilterButton active={filter === "all"} onClick={() => onFilterChange("all")} label={`All ${upcoming.length + finished.length}`} />
          <FilterButton active={filter === "upcoming"} onClick={() => onFilterChange("upcoming")} label={`Upcoming ${upcoming.length}`} />
          <FilterButton active={filter === "finished"} onClick={() => onFilterChange("finished")} label={`Finished ${finished.length}`} />
        </div>
      </div>
      <div className="match-list">
        {rows.length ? (
          rows.map((match) => <MatchRow key={match.id} match={match} onNavigate={onNavigate} />)
        ) : (
          <div className="empty-state">Chưa có lịch đấu.</div>
        )}
      </div>
      {special.length ? (
        <div className="special-list">
          <p className="special-title">Tạm hoãn / hoãn / hủy</p>
          {special.map((match) => (
            <MatchRow key={match.id} match={match} onNavigate={onNavigate} />
          ))}
        </div>
      ) : null}
    </section>
  );
}

function FilterButton({ active, onClick, label }) {
  return (
    <button type="button" className={`filter-button ${active ? "active" : ""}`} onClick={onClick}>
      {label}
    </button>
  );
}
