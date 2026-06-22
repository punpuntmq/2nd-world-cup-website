const { createElement: h, useEffect, useState } = window.React || {};

if (!window.React || !window.ReactDOM) {
  document.getElementById("root").innerHTML = '<div class="app-error">Không tải được React runtime.</div>';
} else {
  const root = window.ReactDOM.createRoot(document.getElementById("root"));
  root.render(h(App));
}

function App() {
  const [data, setData] = useState(null);
  const [error, setError] = useState("");
  const [route, setRoute] = useState(parseRoute());
  const [activeGroup, setActiveGroup] = useState("");
  const [scheduleFilter, setScheduleFilter] = useState("all");
  const [isRefreshing, setRefreshing] = useState(false);

  useEffect(() => {
    const onPop = () => setRoute(parseRoute());
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  useEffect(() => {
    let alive = true;
    let timer = 0;

    const load = async () => {
      try {
        const response = await fetch("/api/state", { cache: "no-store" });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const nextData = await response.json();
        if (!alive) return;
        setData(nextData);
        setError("");
        if (!activeGroup && nextData.standings?.length) {
          setActiveGroup(nextData.standings[0].group);
        }
        timer = window.setTimeout(load, pollDelayMs(nextData.meta));
      } catch (err) {
        if (!alive) return;
        setError(err.message);
        timer = window.setTimeout(load, 30000);
      }
    };

    load();
    return () => {
      alive = false;
      window.clearTimeout(timer);
    };
  }, [activeGroup]);

  const navigate = (path) => {
    if (window.location.pathname !== path) {
      window.history.pushState({}, "", path);
    }
    setRoute(parseRoute(path));
  };

  const refreshNow = async () => {
    setRefreshing(true);
    try {
      const response = await fetch("/api/refresh", { method: "POST" });
      const payload = await response.json();
      setData(payload.state);
      setError(payload.error || "");
    } catch (err) {
      setError(err.message);
    } finally {
      setRefreshing(false);
    }
  };

  if (!data) {
    return h("div", { className: "loading-screen" }, error ? `Không tải được dữ liệu: ${error}` : "Đang tải World Cup...");
  }

  return h(
    "div",
    { className: "app-shell" },
    h(Topbar, { meta: data.meta, error, isRefreshing, onRefresh: refreshNow, onNavigate: navigate }),
    h(StatusBanner, { meta: data.meta, error }),
    route.name === "team"
      ? h(TeamPage, { data, teamId: route.id, onNavigate: navigate })
      : route.name === "match"
        ? h(MatchPage, { data, matchId: route.id, onNavigate: navigate })
        : h(DashboardPage, {
            data,
            activeGroup,
            scheduleFilter,
            setActiveGroup,
            setScheduleFilter,
            onNavigate: navigate,
          })
  );
}

function Topbar({ meta, error, isRefreshing, onRefresh, onNavigate }) {
  const status = meta?.refreshStatus || "idle";
  const rateLimited = status === "rate_limited" && isFuture(meta?.nextAllowedRefreshAt);
  const disabled = isRefreshing || status === "refreshing" || rateLimited;
  return h(
    "header",
    { className: "topbar" },
    h(
      "button",
      { className: "brand", type: "button", onClick: () => onNavigate("/") },
      h("span", { className: "brand-mark" }, "WC"),
      h("span", { className: "brand-text" }, "World Cup")
    ),
    h(
      "div",
      { className: "topbar-status" },
      h("span", { className: `source-tag ${meta?.source === "football-data" ? "live" : ""}` }, meta?.source === "football-data" ? "API" : "FAKE"),
      h("span", { className: `sync-pill ${statusClass(status)}` }, statusLabel(status)),
      h("span", { className: "quota-pill" }, `Quota ${meta?.remainingCalls ?? "-"} / 10`),
      h("span", null, error ? `${formatDateTime(meta?.fetchedAt)} · fallback` : formatDateTime(meta?.fetchedAt)),
      h("button", { className: "icon-button", type: "button", title: refreshTitle(meta), disabled, onClick: onRefresh }, "↻")
    )
  );
}

function StatusBanner({ meta, error }) {
  const status = meta?.refreshStatus || "idle";
  if (status === "refreshing") {
    return h("div", { className: "status-banner refreshing" }, "Backend đang cập nhật snapshot World Cup.");
  }
  if (status === "rate_limited") {
    return h("div", { className: "status-banner rate-limited" }, `Đã chạm quota upstream. Lần refresh tiếp theo: ${formatDateTime(meta?.nextAllowedRefreshAt)}.`);
  }
  if (status === "error" || meta?.lastError || error) {
    return h("div", { className: "status-banner error" }, `Refresh lỗi: ${meta?.lastError || error}. Dashboard vẫn dùng dữ liệu gần nhất.`);
  }
  if (meta?.isStale || status === "stale") {
    return h("div", { className: "status-banner stale" }, "Dữ liệu đang cũ; dashboard vẫn hiển thị snapshot gần nhất từ backend.");
  }
  return null;
}

function DashboardPage({ data, activeGroup, scheduleFilter, setActiveGroup, setScheduleFilter, onNavigate }) {
  return h(
    "main",
    { className: "layout-shell" },
    h(TeamsPanel, { teams: data.teams || [], onNavigate }),
    h(
      "section",
      { className: "center-stage", "aria-label": "Khu vực trận đấu" },
      h(LiveSection, { matches: data.liveMatches || [], onNavigate }),
      h(NextMatchesSection, { matches: data.upcomingMatches || [], onNavigate }),
      h(SchedulePanel, {
        data,
        filter: scheduleFilter,
        onFilterChange: setScheduleFilter,
        onNavigate,
      })
    ),
    h(StandingsPanel, { groups: data.standings || [], activeGroup, onGroupChange: setActiveGroup })
  );
}

function TeamsPanel({ teams, onNavigate }) {
  return h(
    "aside",
    { className: "panel teams-panel", "aria-label": "Danh sách đội bóng" },
    h("div", { className: "panel-heading" }, h("p", { className: "eyebrow" }, "Teams"), h("h1", null, "FIFA World Cup")),
    h(
      "div",
      { className: "team-list" },
      teams.length
        ? teams.map((team) =>
            h(
              "button",
              { key: team.id, type: "button", className: "team-button", onClick: () => onNavigate(`/team/${team.id}`) },
              h(Crest, { src: team.crest, fallback: team.tla }),
              h("span", null, h("span", { className: "team-name" }, team.shortName || team.name), h("span", { className: "team-meta" }, team.group || "World Cup")),
              h("span", { className: "open-chip" }, "›")
            )
          )
        : h("div", { className: "empty-state" }, "Chưa có đội bóng.")
    )
  );
}

function LiveSection({ matches, onNavigate }) {
  if (!matches.length) return null;
  return h(
    "section",
    { className: "live-panel" },
    h("div", { className: "panel-heading row-heading compact-heading" }, h("div", null, h("p", { className: "eyebrow" }, "Live"), h("h2", null, "Đang trực tiếp"))),
    h(
      "div",
      { className: "live-grid" },
      matches.map((match) => h(LiveCard, { key: match.id, match, onNavigate }))
    )
  );
}

function LiveCard({ match, onNavigate }) {
  return h(
    "article",
    { className: "live-card" },
    h("div", { className: "match-topline" }, h("span", { className: "status-pill live" }, match.minute || match.statusLabel), h("span", null, match.group || match.stage)),
    h(MatchTeams, { match, onNavigate, large: true }),
    h(ScoreButton, { match, onNavigate, large: true })
  );
}

function NextMatchesSection({ matches, onNavigate }) {
  const nextMatches = getNextWindow(matches);
  if (!nextMatches.length) {
    return h("section", { className: "next-panel" }, h("div", { className: "empty-state" }, "Chưa có trận kế tiếp."));
  }

  return h(
    "section",
    { className: "next-panel" },
    h("div", { className: "panel-heading row-heading compact-heading" }, h("div", null, h("p", { className: "eyebrow" }, "Next Match"), h("h2", null, "Trận kế tiếp"))),
    h(
      "div",
      { className: `next-grid count-${nextMatches.length}` },
      nextMatches.map((match) => h(NextMatchCard, { key: match.id, match, onNavigate }))
    )
  );
}

function NextMatchCard({ match, onNavigate }) {
  return h(
    "article",
    { className: "next-card" },
    h("div", { className: "next-time" }, formatDateTime(match.kickoffUTC)),
    h(MatchTeams, { match, onNavigate, large: false }),
    h(ScoreButton, { match, onNavigate, large: false }),
    h("div", { className: "hero-meta" }, h("span", null, match.group || match.stage), h("span", null, match.matchday ? `Matchday ${match.matchday}` : "World Cup"))
  );
}

function SchedulePanel({ data, filter, onFilterChange, onNavigate }) {
  const upcoming = data.upcomingMatches || [];
  const finished = data.finishedMatches || [];
  const special = data.specialMatches || [];
  const rows = filter === "upcoming" ? upcoming : filter === "finished" ? finished : [...upcoming, ...finished];

  return h(
    "section",
    { className: "schedule-panel" },
    h(
      "div",
      { className: "panel-heading row-heading" },
      h("div", null, h("p", { className: "eyebrow" }, "Schedule"), h("h2", null, "All matches")),
      h(
        "div",
        { className: "schedule-filters", role: "tablist", "aria-label": "Lọc lịch đấu" },
        h(FilterButton, { active: filter === "all", onClick: () => onFilterChange("all"), label: `All ${upcoming.length + finished.length}` }),
        h(FilterButton, { active: filter === "upcoming", onClick: () => onFilterChange("upcoming"), label: `Upcoming ${upcoming.length}` }),
        h(FilterButton, { active: filter === "finished", onClick: () => onFilterChange("finished"), label: `Finished ${finished.length}` })
      )
    ),
    h("div", { className: "match-list" }, rows.length ? rows.map((match) => h(MatchRow, { key: match.id, match, onNavigate })) : h("div", { className: "empty-state" }, "Chưa có lịch đấu.")),
    special.length
      ? h("div", { className: "special-list" }, h("p", { className: "special-title" }, "Tạm hoãn / hoãn / hủy"), special.map((match) => h(MatchRow, { key: match.id, match, onNavigate })))
      : null
  );
}

function FilterButton({ active, onClick, label }) {
  return h("button", { type: "button", className: `filter-button ${active ? "active" : ""}`, onClick }, label);
}

function MatchRow({ match, onNavigate }) {
  return h(
    "div",
    { className: `match-row ${match.statusGroup}` },
    h("div", { className: "match-time" }, h("strong", null, match.statusLabel), h("br"), h("span", null, formatShortDate(match.kickoffUTC))),
    h(TeamButton, { team: match.homeTeam, className: "row-team home-team", onNavigate }),
    h(ScoreButton, { match, onNavigate }),
    h(TeamButton, { team: match.awayTeam, className: "row-team away-team", onNavigate }),
    h("div", { className: "row-stage" }, match.group || match.stage || "")
  );
}

function MatchTeams({ match, onNavigate, large }) {
  return h(
    "div",
    { className: large ? "match-teams large" : "match-teams" },
    h(TeamButton, { team: match.homeTeam, className: "match-team-link", onNavigate }),
    h("span", { className: "versus" }, "vs"),
    h(TeamButton, { team: match.awayTeam, className: "match-team-link", onNavigate })
  );
}

function TeamButton({ team, className, onNavigate }) {
  return h(
    "button",
    { type: "button", className, onClick: () => team?.id && onNavigate(`/team/${team.id}`), disabled: !team?.id },
    h(Crest, { src: team?.crest, fallback: team?.tla }),
    h("span", { className: "row-team-name" }, team?.shortName || team?.name || "TBD")
  );
}

function ScoreButton({ match, onNavigate, large }) {
  return h(
    "button",
    { type: "button", className: large ? "score-button large" : "score-button", onClick: () => onNavigate(`/match/${match.id}`), title: "Xem chi tiết trận đấu" },
    h("span", null, match.homeScore),
    h("span", { className: "score-separator" }, "-"),
    h("span", null, match.awayScore)
  );
}

function StandingsPanel({ groups, activeGroup, onGroupChange }) {
  const selected = groups.find((group) => group.group === activeGroup) || groups[0];
  return h(
    "aside",
    { className: "panel standings-panel", "aria-label": "Bảng xếp hạng" },
    h("div", { className: "panel-heading" }, h("p", { className: "eyebrow" }, "Standings"), h("h2", null, "Group table")),
    groups.length
      ? h(
          React.Fragment,
          null,
          h(
            "div",
            { className: "group-tabs" },
            groups.map((group) =>
              h(
                "button",
                { key: group.group, type: "button", className: `group-tab ${group.group === selected.group ? "active" : ""}`, onClick: () => onGroupChange(group.group) },
                h("span", null, group.group),
                h("strong", null, group.memberCount ?? group.rows?.length ?? 0)
              )
            )
          ),
          h("div", { className: "group-info" }, `${selected.group} · ${(selected.rows || []).length} đội trong group`),
          h(
            "div",
            { className: "table-wrap" },
            h(
              "table",
              { className: "standings-table" },
              h("thead", null, h("tr", null, ["#", "Team", "P", "W", "D", "L", "GD", "Pts"].map((label) => h("th", { key: label }, label)))),
              h(
                "tbody",
                null,
                (selected.rows || []).map((row) =>
                  h(
                    "tr",
                    { key: row.team.id },
                    h("td", null, row.position),
                    h("td", null, h("span", { className: "standing-team" }, h(Crest, { src: row.team.crest, fallback: row.team.tla }), h("strong", null, row.team.tla || row.team.shortName))),
                    h("td", null, row.playedGames),
                    h("td", null, row.won),
                    h("td", null, row.draw),
                    h("td", null, row.lost),
                    h("td", { className: row.goalDifference >= 0 ? "gd-positive" : "gd-negative" }, formatGoalDiff(row.goalDifference)),
                    h("td", { className: "points-cell" }, row.points)
                  )
                )
              )
            )
          )
        )
      : h("div", { className: "empty-state" }, "Chưa có bảng xếp hạng.")
  );
}

function TeamPage({ data, teamId, onNavigate }) {
  const team = (data.teams || []).find((item) => item.id === teamId);
  const matches = getAllMatches(data).filter((match) => match.homeTeam.id === teamId || match.awayTeam.id === teamId);
  if (!team) return h(NotFoundPage, { title: "Không tìm thấy đội", onNavigate });

  return h(
    "main",
    { className: "detail-layout" },
    h(TeamsPanel, { teams: data.teams || [], onNavigate }),
    h(
      "section",
      { className: "detail-page" },
      h("button", { className: "back-button", type: "button", onClick: () => onNavigate("/") }, "← Dashboard"),
      h(
        "div",
        { className: "detail-hero" },
        h(Crest, { src: team.crest, fallback: team.tla }),
        h("div", null, h("p", { className: "eyebrow" }, team.group || "World Cup"), h("h1", null, team.name), h("p", { className: "detail-muted" }, team.coach ? `HLV: ${team.coach}` : "Chưa có thông tin HLV"))
      ),
      h(
        "div",
        { className: "detail-grid" },
        h(DetailItem, { label: "Rank", value: team.rank ? `#${team.rank}` : "-" }),
        h(DetailItem, { label: "Points", value: team.points || 0 }),
        h(DetailItem, { label: "Played", value: team.played || 0 }),
        h(DetailItem, { label: "Goal diff", value: formatGoalDiff(team.goalDiff || 0) })
      ),
      h("h2", { className: "section-title" }, "Trận đấu của đội"),
      h("div", { className: "match-list detail-matches" }, matches.length ? matches.map((match) => h(MatchRow, { key: match.id, match, onNavigate })) : h("div", { className: "empty-state" }, "Chưa có trận đấu.")),
      h("h2", { className: "section-title" }, "Cầu thủ"),
      h(
        "div",
        { className: "squad-grid" },
        (team.squad || []).length
          ? team.squad.map((player) => h("div", { key: player.id, className: "player-card" }, h("strong", null, player.name), h("span", null, shortPosition(player.position)), h("small", null, player.nationality)))
          : h("div", { className: "empty-state" }, "Chưa có dữ liệu cầu thủ.")
      )
    ),
    h(StandingsPanel, { groups: data.standings || [], activeGroup: team.group, onGroupChange: () => {} })
  );
}

function MatchPage({ data, matchId, onNavigate }) {
  const match = findMatch(data, matchId);
  if (!match) return h(NotFoundPage, { title: "Không tìm thấy trận đấu", onNavigate });

  return h(
    "main",
    { className: "detail-layout" },
    h(TeamsPanel, { teams: data.teams || [], onNavigate }),
    h(
      "section",
      { className: "detail-page" },
      h("button", { className: "back-button", type: "button", onClick: () => onNavigate("/") }, "← Dashboard"),
      h(
        "div",
        { className: `match-detail-card ${match.statusGroup}` },
        h("div", { className: "match-topline" }, h("div", null, h("p", { className: "eyebrow" }, match.stage || "World Cup"), h("h1", null, match.group || "Match")), h("span", { className: `status-pill ${match.statusGroup}` }, match.statusLabel)),
        h(MatchTeams, { match, onNavigate, large: true }),
        h(ScoreButton, { match, onNavigate, large: true }),
        h("div", { className: "hero-meta" }, h("span", null, formatDateTime(match.kickoffUTC)), h("span", null, match.matchday ? `Matchday ${match.matchday}` : "World Cup"), h("span", null, match.minute))
      ),
      h("h2", { className: "section-title" }, "Thông tin trận đấu"),
      h(
        "div",
        { className: "detail-grid" },
        h(DetailItem, { label: "Status", value: match.statusLabel }),
        h(DetailItem, { label: "Stage", value: match.stage || "-" }),
        h(DetailItem, { label: "Group", value: match.group || "-" }),
        h(DetailItem, { label: "Kickoff", value: formatDateTime(match.kickoffUTC) })
      )
    ),
    h(StandingsPanel, { groups: data.standings || [], activeGroup: match.group, onGroupChange: () => {} })
  );
}

function DetailItem({ label, value }) {
  return h("div", { className: "detail-item" }, h("div", { className: "detail-label" }, label), h("div", { className: "detail-value" }, value));
}

function NotFoundPage({ title, onNavigate }) {
  return h("main", { className: "detail-layout single" }, h("section", { className: "detail-page" }, h("button", { className: "back-button", type: "button", onClick: () => onNavigate("/") }, "← Dashboard"), h("div", { className: "empty-state" }, title)));
}

function Crest({ src, fallback }) {
  const label = fallback || "WC";
  if (!src) return h("span", { className: "crest-fallback" }, label);
  return h("img", { className: "crest", src, alt: label, onError: (event) => event.currentTarget.replaceWith(Object.assign(document.createElement("span"), { className: "crest-fallback", textContent: label })) });
}

function getNextWindow(matches) {
  const upcoming = (matches || []).filter((match) => match.statusGroup === "upcoming" && parseDate(match.kickoffUTC)).sort((a, b) => parseDate(a.kickoffUTC) - parseDate(b.kickoffUTC));
  if (!upcoming.length) return [];
  const firstTime = parseDate(upcoming[0].kickoffUTC).getTime();
  const sixHours = 6 * 60 * 60 * 1000;
  return upcoming.filter((match) => Math.abs(parseDate(match.kickoffUTC).getTime() - firstTime) <= sixHours).slice(0, 2);
}

function getAllMatches(data) {
  return [...(data.liveMatches || []), ...(data.upcomingMatches || []), ...(data.finishedMatches || []), ...(data.specialMatches || [])];
}

function findMatch(data, id) {
  return getAllMatches(data).find((match) => match.id === id) || null;
}

function pollDelayMs(meta) {
  const fallback = Math.max(10, meta?.refreshSeconds || 45) * 1000;
  const status = meta?.refreshStatus || "idle";
  if (status === "rate_limited") {
    const nextAllowed = parseDate(meta?.nextAllowedRefreshAt);
    if (nextAllowed && nextAllowed.getTime() > Date.now()) {
      return Math.max(10000, nextAllowed.getTime() - Date.now() + 1000);
    }
    return Math.max(fallback, 30000);
  }
  if (status === "refreshing") {
    return Math.min(fallback, 5000);
  }
  if (meta?.isStale || status === "stale") {
    return Math.max(fallback, 30000);
  }
  return fallback;
}

function statusClass(status = "idle") {
  return status.replace(/_/g, "-");
}

function statusLabel(status = "idle") {
  switch (status) {
    case "refreshing":
      return "Đang cập nhật";
    case "rate_limited":
      return "Hết quota";
    case "stale":
      return "Dữ liệu cũ";
    case "error":
      return "Lỗi refresh";
    default:
      return "Ổn định";
  }
}

function refreshTitle(meta) {
  const status = meta?.refreshStatus || "idle";
  if (status === "refreshing") return "Backend đang cập nhật";
  if (status === "rate_limited" && isFuture(meta?.nextAllowedRefreshAt)) {
    return `Chờ tới ${formatDateTime(meta?.nextAllowedRefreshAt)}`;
  }
  return "Cập nhật ngay";
}

function isFuture(value) {
  const date = parseDate(value);
  return Boolean(date && date.getTime() > Date.now());
}

function parseRoute(path = window.location.pathname) {
  const teamMatch = path.match(/^\/(?:server\/)?team\/(\d+)\/?$/);
  if (teamMatch) return { name: "team", id: Number(teamMatch[1]) };
  const matchMatch = path.match(/^\/(?:server\/)?(?:match|matchs|matches)\/(\d+)\/?$/);
  if (matchMatch) return { name: "match", id: Number(matchMatch[1]) };
  return { name: "dashboard" };
}

function parseDate(value) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatDateTime(value) {
  const date = parseDate(value);
  if (!date) return "-";
  return new Intl.DateTimeFormat("vi-VN", {
    weekday: "short",
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatShortDate(value) {
  const date = parseDate(value);
  if (!date) return "-";
  return new Intl.DateTimeFormat("vi-VN", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatGoalDiff(value) {
  return value > 0 ? `+${value}` : `${value}`;
}

function shortPosition(position = "") {
  return position
    .replace("Goalkeeper", "GK")
    .replace("Defender", "DEF")
    .replace("Defence", "DEF")
    .replace("Midfielder", "MID")
    .replace("Midfield", "MID")
    .replace("Attacker", "FWD")
    .replace("Forward", "FWD")
    .replace("Offence", "FWD");
}
