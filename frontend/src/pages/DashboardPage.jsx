import { LiveSection } from "../components/matches/LiveSection.jsx";
import { NextMatchesSection } from "../components/matches/NextMatchesSection.jsx";
import { SchedulePanel } from "../components/matches/SchedulePanel.jsx";
import { StandingsPanel } from "../components/standings/StandingsPanel.jsx";
import { TeamsPanel } from "../components/teams/TeamsPanel.jsx";

export function DashboardPage({
  data,
  activeGroup,
  scheduleFilter,
  setActiveGroup,
  setScheduleFilter,
  onNavigate,
}) {
  return (
    <main className="layout-shell">
      <TeamsPanel teams={data.teams || []} onNavigate={onNavigate} />
      <section className="center-stage" aria-label="Khu vực trận đấu">
        <LiveSection matches={data.liveMatches || []} onNavigate={onNavigate} />
        <NextMatchesSection matches={data.upcomingMatches || []} onNavigate={onNavigate} />
        <SchedulePanel
          data={data}
          filter={scheduleFilter}
          onFilterChange={setScheduleFilter}
          onNavigate={onNavigate}
        />
      </section>
      <StandingsPanel groups={data.standings || []} activeGroup={activeGroup} onGroupChange={setActiveGroup} />
    </main>
  );
}
