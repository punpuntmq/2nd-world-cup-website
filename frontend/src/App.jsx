import { DashboardPage } from "./pages/DashboardPage.jsx";
import { MatchPage } from "./pages/MatchPage.jsx";
import { TeamPage } from "./pages/TeamPage.jsx";
import { StatusBanner } from "./components/shell/StatusBanner.jsx";
import { Topbar } from "./components/shell/Topbar.jsx";
import { useAppRoute } from "./hooks/useAppRoute.js";
import { useWorldCupState } from "./hooks/useWorldCupState.js";

export default function App() {
  const { route, navigate } = useAppRoute();
  const {
    data,
    error,
    activeGroup,
    scheduleFilter,
    setActiveGroup,
    setScheduleFilter,
  } = useWorldCupState();

  if (!data) {
    return (
      <div className="loading-screen">
        {error ? `Không tải được dữ liệu: ${error}` : "Đang tải World Cup..."}
      </div>
    );
  }

  return (
    <div className="app-shell">
      <Topbar
        meta={data.meta}
        error={error}
        onNavigate={navigate}
      />
      <StatusBanner meta={data.meta} error={error} />
      {route.name === "team" ? (
        <TeamPage data={data} teamId={route.id} onNavigate={navigate} />
      ) : route.name === "match" ? (
        <MatchPage data={data} matchId={route.id} onNavigate={navigate} />
      ) : (
        <DashboardPage
          data={data}
          activeGroup={activeGroup}
          scheduleFilter={scheduleFilter}
          setActiveGroup={setActiveGroup}
          setScheduleFilter={setScheduleFilter}
          onNavigate={navigate}
        />
      )}
    </div>
  );
}
