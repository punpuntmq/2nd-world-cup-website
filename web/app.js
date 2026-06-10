const state = {
  data: null,
  selectedTeamId: null,
  activeGroup: null,
  timer: null,
};

const $ = (selector) => document.querySelector(selector);

document.addEventListener("DOMContentLoaded", () => {
  $("#refreshButton").addEventListener("click", () => refreshNow());
  loadState();
});

async function loadState() {
  try {
    const response = await fetch("/api/state", { cache: "no-store" });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    state.data = data;
    if (!state.activeGroup && data.standings?.length) {
      state.activeGroup = data.standings[0].group;
    }
    if (!state.selectedTeamId && data.teams?.length) {
      state.selectedTeamId = data.teams[0].id;
    }
    render();
    scheduleNextPoll(data.meta?.refreshSeconds || 45);
  } catch (error) {
    renderLoadError(error);
    scheduleNextPoll(30);
  }
}

async function refreshNow() {
  const button = $("#refreshButton");
  button.disabled = true;
  try {
    const response = await fetch("/api/refresh", { method: "POST" });
    const payload = await response.json();
    state.data = payload.state;
    render();
  } catch (error) {
    renderLoadError(error);
  } finally {
    button.disabled = false;
  }
}

function scheduleNextPoll(seconds) {
  clearTimeout(state.timer);
  state.timer = setTimeout(loadState, Math.max(10, seconds) * 1000);
}

function render() {
  const data = state.data;
  if (!data) return;
  $("#competitionName").textContent = `${data.meta?.competitionName || "World Cup"} ${data.meta?.competitionCode || ""}`.trim();
  renderMeta(data.meta);
  renderTeams(data.teams || []);
  renderHero(data.currentMatch);
  renderHighlight(data.highlight, data.nextMatch);
  renderTeamDetail(findSelectedTeam());
  renderMatches(data);
  renderStandings(data.standings || []);
}

function renderMeta(meta = {}) {
  const source = $("#sourceTag");
  source.textContent = meta.source === "football-data" ? "API" : "FAKE";
  source.classList.toggle("live", meta.source === "football-data");
  const updatedAt = meta.fetchedAt ? formatDateTime(meta.fetchedAt) : "Đang chờ dữ liệu";
  $("#updatedAt").textContent = meta.lastError ? `${updatedAt} · fallback` : updatedAt;
}

function renderTeams(teams) {
  const list = $("#teamList");
  list.innerHTML = "";
  if (!teams.length) {
    list.innerHTML = `<div class="empty-state">Chưa có đội bóng.</div>`;
    return;
  }

  for (const team of teams) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = `team-button ${team.id === state.selectedTeamId ? "active" : ""}`;
    button.innerHTML = `
      ${crestMarkup(team.crest, team.tla)}
      <span>
        <span class="team-name">${escapeHTML(team.shortName || team.name)}</span>
        <span class="team-meta">${escapeHTML(team.group || "World Cup")} · ${team.points || 0} pts</span>
      </span>
      <span class="rank-chip">${team.rank ? `#${team.rank}` : ""}</span>
    `;
    button.addEventListener("click", () => {
      state.selectedTeamId = team.id;
      renderTeams(teams);
      renderTeamDetail(team);
    });
    list.appendChild(button);
  }
}

function renderHero(match) {
  const root = $("#currentMatch");
  if (!match) {
    root.innerHTML = `<div class="empty-state">Chưa có trận hiện tại.</div>`;
    return;
  }

  root.innerHTML = `
    <div class="match-topline">
      <div>
        <p class="eyebrow">${escapeHTML(match.stage || "World Cup")}</p>
        <h2>${escapeHTML(match.group || "Match")}</h2>
      </div>
      <span class="status-pill ${match.statusGroup}">${escapeHTML(match.statusLabel)}</span>
    </div>
    <div class="scoreboard">
      <div class="match-team">
        ${crestMarkup(match.homeTeam.crest, match.homeTeam.tla)}
        <div class="match-team-name">${escapeHTML(match.homeTeam.name)}</div>
      </div>
      <div class="score-box">
        <div class="score-main">
          <span>${escapeHTML(match.homeScore)}</span>
          <span class="score-separator">-</span>
          <span>${escapeHTML(match.awayScore)}</span>
        </div>
        <span class="status-pill ${match.statusGroup}">${escapeHTML(match.minute)}</span>
      </div>
      <div class="match-team">
        ${crestMarkup(match.awayTeam.crest, match.awayTeam.tla)}
        <div class="match-team-name">${escapeHTML(match.awayTeam.name)}</div>
      </div>
    </div>
    <div class="hero-meta">
      <span>${formatDateTime(match.kickoffUTC)}</span>
      <span>${escapeHTML(match.venue)}</span>
      <span>${escapeHTML(match.matchday ? `Matchday ${match.matchday}` : "World Cup")}</span>
    </div>
  `;
}

function renderHighlight(highlight = {}, nextMatch) {
  const root = $("#matchHighlight");
  root.innerHTML = `
    <p class="highlight-title">${escapeHTML(highlight.title || "Trạng thái")}</p>
    <p class="highlight-message">${escapeHTML(highlight.message || "Đang cập nhật")}</p>
    <p class="highlight-detail">${escapeHTML(highlight.detail || "")}</p>
    ${nextMatch ? `<p class="highlight-detail">Next: ${escapeHTML(nextMatch.homeTeam.shortName)} vs ${escapeHTML(nextMatch.awayTeam.shortName)} · ${formatDateTime(nextMatch.kickoffUTC)}</p>` : ""}
  `;
}

function renderTeamDetail(team) {
  const root = $("#teamDetail");
  if (!team) {
    root.innerHTML = `<div class="empty-state">Chọn một đội để xem chi tiết.</div>`;
    return;
  }

  const squad = (team.squad || []).slice(0, 10).map((player) => `
    <div class="player-row">
      <span class="player-number">${escapeHTML(player.shirtNumber || "-")}</span>
      <span class="player-name">${escapeHTML(player.name)}</span>
      <span>${escapeHTML(shortPosition(player.position))}</span>
    </div>
  `).join("");

  root.innerHTML = `
    <div class="team-detail-header">
      ${crestMarkup(team.crest, team.tla)}
      <div>
        <p class="eyebrow">${escapeHTML(team.group || "World Cup")}</p>
        <h2 class="team-detail-title">${escapeHTML(team.name)}</h2>
      </div>
    </div>
    <div class="detail-grid">
      <div class="detail-item"><div class="detail-label">Rank</div><div class="detail-value">${team.rank ? `#${team.rank}` : "-"}</div></div>
      <div class="detail-item"><div class="detail-label">Points</div><div class="detail-value">${team.points || 0}</div></div>
      <div class="detail-item"><div class="detail-label">Coach</div><div class="detail-value">${escapeHTML(team.coach || "-")}</div></div>
      <div class="detail-item"><div class="detail-label">Venue</div><div class="detail-value">${escapeHTML(team.venue || "-")}</div></div>
    </div>
    <div class="squad-list">${squad || `<div class="empty-state">Chua co du lieu cau thu.</div>`}</div>
  `;
}

function renderMatches(data) {
  const counts = $("#matchCounts");
  counts.innerHTML = `
    <span class="count-pill">Live ${data.meta?.liveCount || 0}</span>
    <span class="count-pill">Upcoming ${data.meta?.upcomingCount || 0}</span>
    <span class="count-pill">FT ${data.meta?.finishedCount || 0}</span>
  `;

  const mainList = $("#matchList");
  const upcoming = data.upcomingMatches || [];
  const finished = data.finishedMatches || [];
  const rows = [...upcoming, ...finished];
  mainList.innerHTML = rows.length ? rows.map(matchRowMarkup).join("") : `<div class="empty-state">Chưa có lịch đấu.</div>`;

  const special = data.specialMatches || [];
  $("#specialList").innerHTML = special.length
    ? `<p class="special-title">Tạm hoãn / hoãn / hủy</p>${special.map(matchRowMarkup).join("")}`
    : "";
}

function matchRowMarkup(match) {
  return `
    <div class="match-row ${match.statusGroup}">
      <div class="match-time">
        <strong>${escapeHTML(match.statusLabel)}</strong><br />
        <span>${formatShortDate(match.kickoffUTC)}</span>
      </div>
      <div class="row-team">${crestMarkup(match.homeTeam.crest, match.homeTeam.tla)}<span class="row-team-name">${escapeHTML(match.homeTeam.shortName || match.homeTeam.name)}</span></div>
      <div class="row-score">${escapeHTML(match.homeScore)}:${escapeHTML(match.awayScore)}</div>
      <div class="row-team">${crestMarkup(match.awayTeam.crest, match.awayTeam.tla)}<span class="row-team-name">${escapeHTML(match.awayTeam.shortName || match.awayTeam.name)}</span></div>
      <div class="row-stage">${escapeHTML(match.group || match.stage || "")}<br />${escapeHTML(match.venue)}</div>
    </div>
  `;
}

function renderStandings(groups) {
  const tabs = $("#groupTabs");
  const body = $("#standingsBody");
  const info = $("#activeGroupInfo");
  tabs.innerHTML = "";

  if (!groups.length) {
    info.textContent = "";
    body.innerHTML = `<tr><td colspan="8" class="empty-state">Chưa có bảng xếp hạng.</td></tr>`;
    return;
  }

  if (!groups.some((group) => group.group === state.activeGroup)) {
    state.activeGroup = groups[0].group;
  }

  for (const group of groups) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = `group-tab ${group.group === state.activeGroup ? "active" : ""}`;
    const memberCount = group.memberCount ?? group.rows?.length ?? 0;
    button.innerHTML = `<span>${escapeHTML(group.group)}</span><strong>${memberCount}</strong>`;
    button.title = `${group.group}: ${memberCount} đội`;
    button.addEventListener("click", () => {
      state.activeGroup = group.group;
      renderStandings(groups);
    });
    tabs.appendChild(button);
  }

  const active = groups.find((group) => group.group === state.activeGroup) || groups[0];
  const activeMemberCount = active.memberCount ?? active.rows?.length ?? 0;
  info.textContent = `${active.group} · ${activeMemberCount} đội trong group`;
  body.innerHTML = active.rows.length ? active.rows.map((row) => `
    <tr>
      <td>${row.position}</td>
      <td>
        <span class="standing-team">
          ${crestMarkup(row.team.crest, row.team.tla)}
          <strong>${escapeHTML(row.team.tla || row.team.shortName)}</strong>
        </span>
      </td>
      <td>${row.playedGames}</td>
      <td>${row.won}</td>
      <td>${row.draw}</td>
      <td>${row.lost}</td>
      <td class="${row.goalDifference >= 0 ? "gd-positive" : "gd-negative"}">${formatGoalDiff(row.goalDifference)}</td>
      <td class="points-cell">${row.points}</td>
    </tr>
  `).join("") : `<tr><td colspan="8" class="empty-state">Group này chưa có đội.</td></tr>`;
}

function findSelectedTeam() {
  return (state.data?.teams || []).find((team) => team.id === state.selectedTeamId) || state.data?.teams?.[0];
}

function crestMarkup(src, fallback) {
  const label = escapeHTML(fallback || "WC");
  if (!src) return `<span class="crest-fallback">${label}</span>`;
  return `<img class="crest" src="${escapeAttr(src)}" alt="${label}" onerror="this.replaceWith(Object.assign(document.createElement('span'), {className: 'crest-fallback', textContent: '${label}'}))" />`;
}

function formatDateTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("vi-VN", {
    weekday: "short",
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatShortDate(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
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
    .replace("Midfielder", "MID")
    .replace("Attacker", "FWD")
    .replace("Forward", "FWD");
}

function renderLoadError(error) {
  $("#currentMatch").innerHTML = `<div class="empty-state">Không tải được dữ liệu: ${escapeHTML(error.message)}</div>`;
}

function escapeHTML(value = "") {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function escapeAttr(value = "") {
  return escapeHTML(value).replaceAll("`", "&#096;");
}
