# Walkthrough — Backend Refactor & Architecture Alignment

> Branch: `refractor` • 12 tasks completed • Build: ✅ PASS

---

## Tổng quan

Toàn bộ 12 tasks trong kế hoạch đã hoàn thành. Mục tiêu chính:

1. **Loại bỏ hoàn toàn fake-data** — Kiến trúc ban đầu dùng để test, giờ không cần nữa
2. **Chuyển kiến trúc client-server** — Client chỉ GET, không POST (server tự refresh)
3. **Dọn dẹp dead code** — Xóa tất cả code thừa không còn sử dụng
4. **Tối ưu hiệu năng** — Fingerprint + cache ViewState, bỏ `reflect.DeepEqual`

---

## Chi tiết từng commit

### P0 — Critical (Ảnh hưởng trực tiếp đến architecture)

#### P0-1 + P0-2: Lightweight fingerprint & cache ViewState
**Commit:** `f27213d`

| File | Thay đổi |
|------|----------|
| [fingerprint.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/service/fingerprint.go) | **[NEW]** Tạo `matchFingerprint()` dựa trên ID+Status+Score+LastUpdated |
| [service.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/service/service.go) | Thêm `cachedView`/`cachedViewAt`/`cachedViewMu` — chỉ gọi `MapViewState()` khi data thực sự đổi |

**Lý do:** `reflect.DeepEqual` trên toàn bộ `RawState` rất tốn CPU. Fingerprint nhẹ hơn nhiều lần, chỉ so sánh các field quan trọng. Cache tránh rebuild ViewState mỗi lần client GET `/api/state`.

---

#### P0-3: Xóa `POST /api/refresh`
**Commit:** `7a51a55`

| File | Thay đổi |
|------|----------|
| [router.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/router/router.go) | Xóa route `api.POST("/refresh", h.Refresh)` |
| [handler.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/handler/handler.go) | Xóa `Refresh()` handler + `refreshStatusCode()` helper + xóa `Refresh` khỏi interface |

**Lý do:** Kiến trúc mới: **server tự refresh theo interval** (scheduler), client **chỉ GET** dữ liệu qua SSE/REST. Client không được phép trigger refresh — đảm bảo quota API không bị abuse.

---

#### P0-4: Bỏ `X-Unfold-*` headers
**Commit:** `c81fc6a`

| File | Thay đổi |
|------|----------|
| [client.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/football/client.go) | Xóa 4 headers: `X-Unfold-Lineups`, `X-Unfold-Bookings`, `X-Unfold-Subs`, `X-Unfold-Goals` |

**Lý do:** Dashboard không hiển thị lineup/booking/substitution/goal detail. Gửi headers này khiến response từ football-data.org nặng hơn cần thiết → tốn bandwidth + parse time.

---

### P1 — Cleanup (Dead code removal)

#### P1-1: Xóa 140 dòng comment cũ trong `api_type.go`
**Commit:** `797d739` • −140 dòng

Khối `/* ... */` chứa các struct cũ đã được refactor — chỉ gây confusion khi đọc code.

#### P1-2: Xóa `currentToken()` trong `client.go`
**Commit:** `beb1019` • −10 dòng

Method không còn được gọi từ đâu cả (đã refactor sang inline lock trong `callAPI`).

#### P1-3: Xóa `service.RefreshMeta` trong `view.go`
**Commit:** `920009b` • −9 dòng

Struct duplicate — `store.RefreshMeta` đã tồn tại và được sử dụng. Service-level copy không ai dùng.

#### P1-4: Xóa thư mục `internal/server/`
Đã xóa trong commit trước đó (architecture refactor). Thư mục trống.

#### P1-5: Xóa `LiveMatchCandidates`, `Live` endpoint, `MatchDetail`
**Commit:** `b9f912d` • −43 dòng

| File | Thay đổi |
|------|----------|
| [business.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/service/business.go) | Xóa `LiveMatchCandidates()` |
| [endpoint_handler.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/football/endpoint_handler.go) | Xóa `Live` struct + `liveTimeEndpoint()` |
| [mapper.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/service/mapper.go) | Xóa `MatchDetail` RequestKind + default case |

**Lý do (Hướng A):** Dashboard chỉ cần snapshot toàn bộ matches. Không cần fetch detail từng match riêng → giảm API calls, đơn giản hóa logic.

#### P1-6: Xóa hoàn toàn fake-data
**Commit:** `cc9cf21` • −17,667 dòng (bao gồm JSON files)

| File | Thay đổi |
|------|----------|
| [client.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/football/client.go) | Xóa `Mode` type, `Fake`/`Real` const, `FetchFake()`, `readJSON()`, `fakeDir` field, `mode` field |
| [config.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/config/config.go) | Xóa `FakeDir`, `ForceFake`, `USE_FAKE_DATA` env var |
| [app.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/app/app.go) | Xóa `FakeDir`/`ForceFake` từ `ClientOptions` |
| `fake-data/` | **[DELETED]** toàn bộ thư mục (matches.json + teams.json) |

**Lý do:** Fake-data là kiến trúc ban đầu để test hệ thống và cho AI đọc cấu trúc JSON thật. Giờ không cần nữa — hệ thống chỉ dùng real API.

---

### P2 — Enhancement

#### P2-1: Doc comment cho shallow clone
**Commit:** `bd0f3e3`

Thêm doc comments cảnh báo callers **MUST NOT** mutate pointer fields/slice elements trên data trả về từ `Snapshot()` và `Save()`.

#### P2-4: Xử lý AWARDED match
**Commit:** `9ae8b41`

| File | Thay đổi |
|------|----------|
| [business.go](file:///d:/File%20code/Go/World%20Cup%20web/2nd-world-cup-website/internal/service/business.go) | Thêm `applyAwardedResult()` — khi match status = `AWARDED` và không có `FullTime` scores, dùng `Score.Winner` để xác định thắng/thua |

**Lý do:** Match bị awarded (xử thua do vi phạm) có thể không có score FullTime — logic cũ skip hoàn toàn → standings sai.

---

## Imports cleanup (kèm trong commit cuối)

| File | Import bị xóa | Lý do |
|------|---------------|-------|
| `client.go` | `os`, `path/filepath` | Chỉ dùng cho `readJSON`/`FetchFake` đã xóa |
| `handler.go` | `context`, `errors`, `store` | Chỉ dùng cho `Refresh` handler đã xóa |
| `view.go` | `time` | Chỉ dùng cho `RefreshMeta` struct đã xóa |

---

## Verification

| Check | Result |
|-------|--------|
| `go build ./...` | ✅ Pass — zero errors |
| Grep `fake` in `*.go` | ✅ Zero matches |
| Grep `ForceFake` in `*.go` | ✅ Zero matches |
| Grep `X-Unfold` in `*.go` | ✅ Zero matches |
| Grep `LiveMatchCandidates` in `*.go` | ✅ Zero matches |
| `fake-data/` directory | ✅ Deleted |
| `internal/server/` directory | ✅ Deleted |

---

## Tóm tắt thống kê

| Metric | Value |
|--------|-------|
| Commits | 12 |
| Files modified | ~15 |
| Lines removed | ~17,800+ |
| Lines added | ~60 |
| New files | 1 (`fingerprint.go`) |
| Deleted files/dirs | 3 (`fake-data/matches.json`, `fake-data/teams.json`, `internal/server/`) |
