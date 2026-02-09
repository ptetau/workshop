# Workshop Jiu Jitsu — Product Requirements Document

Organised by **feature area**. Each feature includes a description and role access. **User stories and acceptance criteria are tracked in [GitHub Issues](https://github.com/ptetau/workshop/issues)** (labels: `type:epic`, `type:story`; milestones: Phase 0–10).

---

## 1. Foundation

### 1.1 Roles & Permissions

All features are gated by role. Every account belongs to exactly one role.

| Role | Description |
|------|-------------|
| **Admin** | Gym owner/operator. Full system access including schedule, billing, grading, and user management. |
| **Coach** | Instructor. Can manage check-ins, view attendance, manage kiosk mode, propose gradings, curate content, and control curriculum rotors. |
| **Member** | Active paying student. Can check in, flag injuries, view schedule, track mat hours, vote on topics, set personal goals, and access study tools. |
| **Trial** | Prospective student. Can check in, sign waiver, and view schedule. No hard visit limit — Admin manually converts to Member when ready. |
| **Guest** | Drop-in visitor. A lightweight account (name + email) is created during the waiver flow. If they return, they are recognised and prompted to convert to Trial or Member. |

### 1.2 Member Statuses

| Status | Description |
|--------|-------------|
| **Active** | Currently training. Appears in kiosk search, member lists, and all active views. |
| **Inactive** | Has stopped checking in (flagged after a configurable number of days). Still visible in member lists with an "inactive" indicator. |
| **Archived** | Manually archived by Admin. Hidden from all active views and kiosk search. All data preserved. Can be restored to Active at any time. |

### 1.3 Programs, Classes & Schedule

The system uses a **three-tier hierarchy**:

```
Program (Adults, Kids, Youth)
  └── Class (Gi Express, Nuts & Bolts, Workshop Kids, etc.)
       └── Session (a specific occurrence on a given day + time)
```

**Programs** are audience groups. Every member belongs to exactly one program. Programs determine which classes a member sees by default, which term structure applies, and which communications they receive. As kids and youth age up, Admin updates their program assignment.

| Program | Audience | Term-Aligned |
|---------|----------|:------------:|
| **Adults** | 18+ years | No |
| **Kids** | Ages 6–11 | Yes (NZ school terms) |
| **Youth** | Ages 12–16 | Yes (NZ school terms) |

**Classes** are named session types within a program. Each class can have its own curriculum rotor (see §5). Classes are extensible by Admin.

| Class | Program | Description | Duration |
|-------|---------|-------------|----------|
| **Nuts & Bolts** | Adults | Beginner skills development (Gi) | 60 min |
| **Gi Express** | Adults | All-levels skills development and sparring (Gi) | 60 min |
| **NoGi Express** | Adults | All-levels skills development and sparring (NoGi) | 60 min |
| **Feet-to-Floor** | Adults | Skills for taking the match to the ground | 60 min |
| **NoGi Long Games** | Adults | Specific sparring and competition prep (NoGi) | 60 min |
| **BYO Task Based Games** | Adults | Practice game design and specific skills (Gi or NoGi) | 60 min |
| **Open Mat** | Adults | Gi or NoGi sparring | 60–180 min |
| **Workshop Kids** | Kids | Kids program | 45 min |
| **Workshop Youth** | Youth | Youth program | 45–60 min |

**Default Weekly Timetable (from workshopjiujitsu.co.nz):**

| Day | Time | Class | Duration |
|-----|------|-------|----------|
| Monday | 12:00 PM | Gi Express | 60 min |
| Monday | 6:00 PM | Nuts & Bolts | 60 min |
| Monday | 7:15 PM | Feet-to-Floor | 60 min |
| Tuesday | 4:00 PM | Workshop Kids | 45 min |
| Tuesday | 4:45 PM | Workshop Youth | 45 min |
| Tuesday | 6:30 PM | Gi Express | 60 min |
| Tuesday | 7:30 PM | BYO Task Based Games | 60 min |
| Wednesday | 12:00 PM | NoGi Express | 60 min |
| Wednesday | 6:00 PM | Nuts & Bolts | 60 min |
| Wednesday | 7:30 PM | NoGi Long Games | 60 min |
| Thursday | 4:45 PM | Workshop Youth | 60 min |
| Thursday | 6:30 PM | NoGi Express | 60 min |
| Thursday | 7:30 PM | NoGi Sparring | 60 min |
| Friday | 12:00 PM | Gi Sparring | 60 min |
| Saturday | 9:00 AM | Workshop Kids | 45 min |
| Saturday | 10:00 AM | Open Mat | 180 min |

Members see classes from their own program by default. Admin can grant visibility to additional classes (e.g., a Youth member allowed in some Adults classes).

**NZ School Terms (default suggestions):**

| Term | Typical Dates |
|------|--------------|
| Term 1 | Early February – mid-April |
| Term 2 | Late April/early May – early July |
| Term 3 | Mid/late July – late September |
| Term 4 | Mid-October – mid-December |

Kids and Youth programs are term-aligned. The system suggests NZ school term dates but Admin must manually confirm each term before the schedule activates.

**Session Resolution:** Sessions are **not** pre-generated. The system resolves `Schedule + Terms − Holidays` on-the-fly. No attendance rows exist until a member checks in.

### 1.4 Terminology

| Term | Definition |
|------|-----------|
| **Mat hours** | Total accumulated training hours for a member — the canonical metric for adult belt progression. Includes recorded hours (from check-ins) and estimated hours (admin/self-submitted). Displayed to members as **"flight time"** in the UI. |
| **Flight time** | Member-facing display name for mat hours. Used in dashboards and training logs. The underlying data model uses `mat_hours`. |
| **Program** | An audience group: Adults, Kids, or Youth. Every member belongs to exactly one program. Determines default class visibility, term structure, and communications targeting. |
| **Class** | A named session type within a program (Gi Express, Nuts & Bolts, Workshop Kids, etc.). Each class can have its own curriculum rotor. |
| **Session** | A specific occurrence of a class on a given day and time (e.g., "Monday 6 PM Nuts & Bolts"). Resolved on-the-fly from Schedule + Terms − Holidays. |
| **Rotor** | A versioned curriculum structure assigned to a class. Contains N concurrent theme categories, each with its own topic queue. Coaches advance topic queues; the rotor auto-advances by default when a topic's duration expires. |
| **Theme** | A positional/tactical category within a rotor (e.g., Standing, Guard, Pinning, In-Between Game). Multiple themes run concurrently each week. |
| **Topic** | A specific technique or drill within a theme's rotating queue. One topic per theme is scheduled each week; members can vote to bump the scheduled topic. |
| **Promotion** | The act of advancing a member to a new belt. Proposed by Coach, approved by Admin. The ceremony is handled outside the system. |
| **Stripe** | A progress marker within a belt level (0–4). For adults, inferred automatically from mat hours. |

### 1.5 List View Pattern

A **cross-cutting UI pattern** that applies to every list/table view in the system. All list views share the same interaction model for pagination, sorting, searching, and row count.

#### 1.5.1 Applicable Views

The following views implement the list view pattern:

| View | Section | Columns (sortable) | Default Sort | Search/Filter Fields |
|------|---------|-------------------|-------------|---------------------|
| **Member List** | §9.3 | Name, Email, Program, Belt, Status, Joined | Name (A–Z) | Name, email, program, status, belt |
| **Attendance (Today)** | §3.1 | Name, Belt, Class, Time, Injury Flag | Time (newest) | Name, class |
| **Historical Attendance** | §3.2 | Name, Belt, Class, Date, Time | Date (newest) | Name, class, date range |
| **Training Log** | §3.3 | Date, Class, Duration, Mat Hours | Date (newest) | Class, date range |
| **Grading Readiness** | §4.5 | Name, Belt, Mat Hours/Attendance, Progress % | Progress (highest) | Program, belt, name |
| **Grading Proposals** | §4.6 | Member, Current Belt, Target Belt, Proposed By, Status, Date | Date (newest) | Status, program, belt |
| **Schedules** | §9.7 | Day, Time, Class, Program, Coach, Duration | Day + Time | Program, class, day |
| **Holidays** | §9.7 | Name, Start Date, End Date | Start Date (soonest) | Name, date range |
| **Terms** | §1.3 | Term Name, Start Date, End Date, Confirmed | Start Date (soonest) | — |
| **Accounts** | §1.1 | Email, Role, Created | Email (A–Z) | Email, role |
| **Notices** | §8.1 | Title, Type, Status, Author, Date | Date (newest) | Type, status, title |
| **Email History** | §8.2 | Subject, Recipients, Status, Scheduled, Sent | Sent (newest) | Status, subject, recipient |
| **Inactive Radar** | §9.6 | Name, Program, Belt, Last Check-In, Days Inactive | Days Inactive (highest) | Program, name |
| **Milestones** | §3.3 | Name, Threshold, Members Achieved | Threshold (lowest) | Name |
| **Clip Library** | §7.1 | Title, Tags, Topic Link, Added By, Date | Date (newest) | Title, tag, topic |
| **Coach Observations** | §8.3 | Member, Author, Date, Preview | Date (newest) | Member, author |
| **Topic Queue** | §5.3 | Position, Topic Name, Duration, Last Covered | Position (ascending) | Topic name |
| **Audit Log** | §14.1 | Timestamp, Actor, Action, Resource, Detail | Timestamp (newest) | Actor, action, resource type, date range |

#### 1.5.2 Pagination

All list views are **server-side paginated**. The client sends `page` and `per_page` query parameters; the server returns the requested slice plus total count metadata.

**Behaviour:**

- Page numbers are **1-indexed**
- Navigation controls: **First**, **Previous**, **[page numbers]**, **Next**, **Last**
- Show at most **5 page number buttons** around the current page (e.g., `3 4 [5] 6 7`)
- Display: **"Showing X–Y of Z results"** below the table
- If total results ≤ selected row count, pagination controls are hidden
- Navigating pages preserves the current sort, search, and row count settings
- All pagination state is encoded in the URL query string (bookmarkable, shareable)

#### 1.5.3 Row Count Selector

A dropdown above the table (right-aligned) allowing the user to choose how many rows per page.

**Options:** `10`, `20`, `50`, `100`, `200`

**Behaviour:**

- Default: **20** rows per page
- Changing row count resets to page 1
- Selection is persisted per-view in the user's browser (localStorage) so it is remembered across sessions
- The dropdown label reads: **"Rows: [N]"**

#### 1.5.4 Column Sorting

Every column listed as "sortable" in the table above supports click-to-sort.

**Behaviour:**

- Click a column header to sort ascending; click again to toggle descending; click a third time to return to default sort
- **Sort indicator:** `▲` (ascending) or `▼` (descending) displayed next to the active sort column header
- Only **one column** can be sorted at a time (no multi-column sort)
- Sort is applied server-side (not client-side JavaScript re-ordering)
- Sort parameter is encoded in the URL query string: `sort=name&dir=asc`

#### 1.5.5 Search & Filters

Each list view has a **filter bar** above the table with view-specific search and filter controls.

**Behaviour:**

- **Text search** fields use server-side `LIKE` matching (case-insensitive, partial match)
- **Dropdown filters** (e.g., program, status, role, belt) filter by exact match
- **Date range** filters provide "from" and "to" date pickers
- Applying any filter resets to page 1
- A **"Clear filters"** link/button appears when any filter is active, resetting all filters
- Active filters are encoded in the URL query string (bookmarkable)
- Filters are combined with **AND** logic (all must match)
- Search is debounced (250ms) for text inputs to avoid excessive server requests
- Empty filters are ignored (show all)

#### 1.5.6 URL State

All list view state is reflected in the URL query string so that views are **bookmarkable and shareable**:

```
/members?q=smith&program=adults&status=active&sort=name&dir=asc&per_page=50&page=2
```

Navigating back/forward in the browser restores the exact list state.

---

## 2. Kiosk & Check-In

### 2.1 Kiosk Mode

A locked check-in screen for the front-desk tablet. Launched by a Coach or Admin from their own logged-in account. The kiosk is **tied to the account that launched it** — only that person's password can exit it. Prevents members from navigating away.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

**Kiosk Flow:**

1. Coach/Admin authenticates → tablet locks into kiosk view tied to that account
2. Member types name → fuzzy search shows matching Active members (autocomplete)
3. Member selects name → system resolves today's sessions on-the-fly and auto-selects the closest session
4. Member confirms or adjusts → can switch session or select multiple
5. Trial member sees prompt after check-in: *"Enjoying Workshop? Talk to your coach about signing up!"*
6. Coach/Admin exits by entering their password; a different Coach/Admin must authenticate with their own credentials first

For **Guests**: a "Guest Check-In" button launches the waiver flow (creating a lightweight account), then records attendance. Returning guests are recognised and prompted to convert.

### 2.2 Check-In by Name Search

The **only** way to check in is by typing your name. Fuzzy search presents a shortlist of matching Active members as the user types. No member ID, email, or barcode is ever required. Inactive and Archived members are hidden from results.

**Access:** Admin — | Coach — | Member ✓ | Trial ✓ | Guest ✓ (via waiver flow)

### 2.3 Session Auto-Select

After selecting their name, the system auto-selects the session closest to the current time. The member can confirm, change, or select multiple sessions.

**Access:** All check-in users (Member, Trial, Guest)

### 2.4 Multi-Session Check-In

Members can check into more than one session in a single visit. Each session is recorded as a separate attendance entry.

**Access:** Member ✓ | Trial ✓ | Guest —

### 2.5 Un-Check-In

If a member checked into the wrong session by mistake, they can undo their check-in from the kiosk. Only available for the current day.

**Access:** Member ✓ | Trial ✓ | Guest —

---

## 3. Attendance & Training Log

### 3.1 Attendance Records

Each check-in creates an attendance record linking a member to a resolved session (class + date + time). Mat hours are calculated from the session's configured duration. **All session types count equally by default** (1 hour = 1 mat hour). Admin can optionally configure a weighting multiplier per class (e.g., Open Mat = 0.5× because it's unstructured, Competition Prep = 1.5×).

**Access:** Admin ✓ (view all) | Coach ✓ (view all) | Member ✓ (own) | Trial ✓ (own) | Guest —

### 3.2 Historical Attendance View

Admin and Coach can navigate to view attendance for other days. A prominent "Back to Today" button allows quick return.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

### 3.3 Training Log

A member-facing projection of their attendance data: classes attended, mat hours displayed as "flight time" (split into recorded and estimated), streaks, belt/stripe icon, and belt progression progress.

**Access:** Admin — | Coach — | Member ✓ | Trial ✓ | Guest —

### 3.4 Estimated Training Hours

For members with incomplete records, Admin or Coach can bulk-add estimated mat hours by selecting a date range and entering estimated weekly hours. Overlaps with existing recorded training are handled explicitly.

**Access:** Admin ✓ | Coach ✓ | Member — (see §3.5 for self-estimates)

### 3.5 Member Self-Estimates

Members can submit their own estimated training periods. Self-estimates require a note and are flagged for Admin review.

**Access:** Admin ✓ (review) | Coach — | Member ✓ (submit) | Trial — | Guest —

---

## 4. Grading & Belt Progression

### 4.1 Belt Progression System

Tracks each member's current belt and progress toward their next grade. IBJJF-aligned.

**Adults (18+ years):** White → Blue → Purple → Brown → Black. 4 stripes per belt before promotion eligibility. Black belt has degrees per IBJJF.

| Belt | Min. Time at Belt | Stripes | Default Mat Hours Threshold |
|------|-------------------|---------|-------------------------------|
| White → Blue | 2 years | 4 | 150 hours |
| Blue → Purple | 1.5 years | 4 | 300 hours |
| Purple → Brown | 1.5 years | 4 | 500 hours |
| Brown → Black | 1 year | 4 | 750 hours |

**Kids (4–15 years):** White → Grey/White → Grey → Grey/Black → Yellow/White → Yellow → Yellow/Black → Orange/White → Orange → Orange/Black → Green/White → Green → Green/Black → Blue (at 16+). Up to 4 stripes per belt. Configurable by Admin.

**Youth (16–17 years):** May hold Purple or Brown belt (up to 4 stripes). Transitions to adult track at 18.

**Access:** Admin ✓ (approve/override/configure) | Coach ✓ (propose) | Member ✓ (view own) | Trial ✓ (view own) | Guest —

### 4.2 Stripe Inference (Adults)

Stripes are automatically inferred by pro-rating accumulated mat hours across the configured threshold. No manual stripe awards needed — the system calculates stripe count from hours.

Formula: `stripe = floor(hours / (threshold / stripe_count))`

| Threshold | Hours Trained | Inferred Stripe |
|-----------|---------------|-----------------|
| 150h | 0–37h | Stripe 0 |
| 150h | 38–74h | Stripe 1 |
| 150h | 75–111h | Stripe 2 |
| 150h | 112–149h | Stripe 3 |
| 150h | 150h+ | Stripe 4 (eligible for promotion) |

### 4.3 Term-Based Grading (Kids/Youth)

For term-based programs, Admin can toggle the grading metric between sessions attended and hours.

| Mode | Metric | Example |
|------|--------|---------|
| **Sessions** (default for kids) | Sessions attended in the current term | "24 of 30 sessions this term" |
| **Hours** | Accumulated mat hours (same as adults) | "45 hours since last promotion" |

**Sessions mode**: eligibility is based on **% of term attendance** — the system counts total available sessions in the current NZ school term, divides by attendance, and compares to admin-configured thresholds. Attendance resets each term. Admin has ultimate discretion over all kids/youth promotions.

**Access:** Admin ✓ (configure) | Coach ✓ (view) | Member ✓ (view own) | Trial — | Guest —

### 4.4 Belt & Stripe Icons

Visual CSS/SVG icons for all belt levels, displayed wherever a member's rank is shown.

- **Belt colour**: solid (adults) or split (kids dual-colour belts like Grey/White)
- **Stripe count**: 0–4 small markers on the belt tab
- **Black belt degrees**: per IBJJF

**Displayed on:** training logs, grading readiness lists, member profiles, dashboards, attendance lists.

**Access:** Admin ✓ (configure) | Coach ✓ (view) | Member ✓ (view own) | Trial ✓ (view own) | Guest —

### 4.5 Grading Readiness List

Auto-generated list of members approaching promotion eligibility. Shows belt icon, progress metric, and coach notes.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

### 4.6 Grading Proposals & Promotions

Coaches propose promotions; Admin approves to make them official. The workflow prevents unilateral promotions. The promotion ceremony (belt presentation) is handled outside the system — the system only tracks the record.

**Access:** Admin ✓ (approve/reject) | Coach ✓ (propose) | Member — | Trial — | Guest —

### 4.7 Admin Overrides

Admin can bypass normal grading thresholds for special circumstances.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

---

## 5. Curriculum Rotor System

### 5.1 Rotor Structure

Each class can have its own **rotor** — a versioned curriculum structure containing **N concurrent theme categories**. All themes run simultaneously each week, each with its own rotating queue of topics.

**Structure:**

```
Class (e.g., "Gi Express")
  └── Rotor (v2 — active)
       ├── Theme: "Standing" ──── Topic Queue ──→ [Single Leg] → [Double Leg] → [Arm Drag] → ...
       ├── Theme: "Guard"    ──── Topic Queue ──→ [Closed Guard Attacks] → [DLR Sweeps] → ...
       ├── Theme: "Pinning"  ──── Topic Queue ──→ [Side Control Subs] → [Mount Attacks] → ...
       └── Theme: "In-Between Game" ─ Topic Queue ─→ [Turtle Escapes] → [Scrambles] → ...
```

Each week, one topic per theme is scheduled. Topics auto-advance when their duration expires (default: 1 week). Coaches can override to extend, skip, or manually advance. Members can vote to bump the scheduled topic (see §6).

**Rotor Versioning:** Rotors are versioned. Admin or Coach can draft a new version of a rotor (adding/removing themes, reordering topics) without affecting the currently active version. When ready, the draft is activated and becomes the new live rotor. Previous versions are preserved for history.

**Coach Ownership:** Coaches can edit and advance the rotors for the classes they are assigned to. Admin can override everything.

**Access:** Admin ✓ (create/manage/override) | Coach ✓ (edit/advance for own classes) | Member ✓ (view if preview on) | Trial — | Guest —

### 5.2 Themes (Concurrent Categories)

Themes are positional or tactical categories that run **concurrently** within a rotor. Every week, all themes are active — each with its own currently scheduled topic.

Example categories: **Standing** (takedowns, clinch work), **Guard** (bottom game, sweeps), **Pinning** (top control, submissions from top), **In-Between Game** (transitions, scrambles, turtle).

Admin defines the theme categories per rotor. Themes are extensible — Admin can add, rename, or remove categories.

### 5.3 Topics & Topic Queues

Each theme has an ordered queue of topics. One topic is scheduled per theme per week (or configurable period). Topics auto-advance when their duration expires.

Each topic has:
- **Name**: the technique or drill (e.g., "Single Leg Takedown")
- **Duration**: how long it stays scheduled (default: 1 week, configurable per topic)
- **Description**: optional notes for the coach

### 5.4 Rotor Advancement

Topics **auto-advance by default** when their configured duration expires. Coaches can override: extend a topic, skip to the next, or manually advance early.

**Access:** Admin ✓ (override all) | Coach ✓ (for classes they own) | Member — | Trial — | Guest —

### 5.5 Rotor Versioning

Rotors are versioned to support drafting improvements without disrupting the live curriculum.

**Access:** Admin ✓ | Coach ✓ (for own classes) | Member — | Trial — | Guest —

### 5.6 Member Rotor Preview

Admin can toggle whether members can see the upcoming topic schedule for each class.

- **Preview on**: members see the full topic queues across all themes
- **Preview off**: members only see the currently scheduled topics

Hidden themes are never shown in previews regardless of the toggle.

**Access:** Admin ✓ (toggle) | Coach ✓ (view) | Member ✓ (view if enabled) | Trial — | Guest —

---

## 6. Topic Voting

Members can vote on topics within a theme category to influence the curriculum. If a voted topic wins, it **bumps the currently scheduled topic** — the bumped topic stays in place in the queue and will run on the next rotation.

### 6.1 Viewing & Voting on Topics

Members can see each theme's topic queue (if preview is enabled) and vote for a topic they'd like to see next. One vote per member per topic per rotation cycle.

Each topic in the queue shows:
- **Name** and description
- **Last covered date** (when it was last the scheduled topic)
- **Current vote count**

**Access:** Admin ✓ (view/override) | Coach ✓ (view/triage) | Member ✓ (vote) | Trial — | Guest —

### 6.2 Vote-Driven Topic Bump

When a topic accumulates enough votes (or Coach/Admin decides to honour the vote), it is **inserted before** the currently scheduled topic. The scheduled topic remains in its queue position and runs on the next rotation.

**Access:** Admin ✓ (override) | Coach ✓ (for own classes) | Member — | Trial — | Guest —

---

## 7. Technical Library & Clips

### 7.1 Clipping Tool

Paste a YouTube link, set start/end timestamps to create a single looping clip. Clips can be cross-linked from the curriculum schedule so that members can study relevant techniques offline. The schedule and clips are not hard-coupled — clips exist independently in the library but can be linked to topics.

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-7.1.1: Create a clip**
As a Member, I want to create a clip from a YouTube video so that I can isolate a specific technique sequence.

- *Given* I find a 40-minute match with a great grip-fight at 12:34
- *When* I paste the YouTube URL and set start=12:34, end=12:44
- *Then* a 10-second looping clip is created and associated with the current theme
- *And* it plays without audio by default (global mute)

**US-7.1.2: Clip loops automatically**
As a Member, I want my clip to loop continuously so that I can study the movement pattern on repeat.

- *Given* I created a 10-second clip
- *When* I play it
- *Then* it loops from start to end seamlessly without manual intervention

### 7.2 Technical Library

A searchable grid of theme cards where clips are stored. Searchable by theme, program, and tag.

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-7.2.1: Browse library by theme**
As a Member, I want to browse the technical library by theme so that I can find clips related to what we're studying.

- *Given* the library has clips across multiple themes
- *When* I filter by "Leg Lasso Series"
- *Then* I see all clips associated with that theme

**US-7.2.2: Search library**
As a Member, I want to search the library by keyword so that I can find specific techniques.

- *Given* the library has 50+ clips
- *When* I search for "armbar"
- *Then* I see clips tagged with or titled "armbar" across all themes and programs

### 7.3 Promote Clip

Coach or Admin can promote a member-submitted clip to the main library so the whole gym benefits.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-7.3.1: Promote a member's clip**
As a Coach, I want to promote a member's clip to the main library so that the whole gym can benefit from it.

- *Given* a member submitted a great clip of a K-Guard entry
- *When* I tap "Promote to Library"
- *Then* the clip appears in the main technical library, visible to all members

### 7.4 Global Mute

Persistent setting ensuring all videos start without audio.

#### User Stories

**US-7.4.1: Videos start muted**
As a Member, I want all clips to start without audio so that I can watch in the gym without disturbing others.

- *Given* I open any clip in the library
- *When* it starts playing
- *Then* audio is muted by default
- *And* I can manually unmute if needed

---

## 8. Communication

### 8.1 Notices

A unified notification system with three types:

| Type | Created by | Published by | Visible to | Where shown |
|------|-----------|-------------|------------|-------------|
| `school_wide` | Admin or Coach (draft) | Admin | All roles | Dashboard, Kiosk |
| `class_specific` | Admin or Coach (draft) | Admin | Coach (assigned to that class) | Coach's class view |
| `holiday` | System (auto from Holiday) | Auto | All roles | Dashboard, Kiosk |

Coach-created notices start in **draft** state and require Admin approval. Admin can create and publish directly.

**Access:** Admin ✓ (publish) | Coach ✓ (draft) | Member ✓ (view) | Trial ✓ (view) | Guest ✓ (view school_wide)

#### User Stories

**US-8.1.1: Coach drafts a notice**
As a Coach, I want to draft a school-wide notice so that I can communicate schedule changes.

- *Given* Open Mat is cancelled this Friday
- *When* I create a notice: "Open Mat cancelled this Friday" with type school_wide
- *Then* the notice is saved as a draft
- *And* Admin receives it for review and publishing

**US-8.1.2: Admin publishes a notice**
As an Admin, I want to publish a coach's draft notice so that all members see it.

- *Given* a coach drafted "Open Mat cancelled this Friday"
- *When* I review and click "Publish"
- *Then* the notice appears on all member dashboards and the kiosk

**US-8.1.3: Class-specific notice**
As a Coach, I want to see a class-specific notice before my class starts so that I remember to make an announcement.

- *Given* Admin set a class-specific notice: "Remind students — grading this Saturday"
- *When* I view my class for today
- *Then* the notice is displayed prominently before the attendance list

**US-8.1.4: Holiday notice auto-generated**
As a Member, I want to see holiday notices automatically so that I know when the gym is closed.

- *Given* Admin entered a holiday: "Easter Weekend — 18–21 Apr"
- *When* I view my dashboard
- *Then* I see a holiday notice: "Gym closed 18–21 Apr: Easter Weekend"

### 8.2 Email System

A full email composition and delivery system for communicating with individuals or groups. Emails are sent via **Resend** (see §8.2.7) and delivered to members' real email addresses. Sent emails are also mirrored as in-app messages on the member's dashboard with a notification badge.

**Access:** Admin ✓ (compose/send) | Coach ✓ (compose/send with Admin approval for bulk) | Member ✓ (view own received) | Trial ✓ (view own received) | Guest —

#### 8.2.1 Recipient Search & Selection

The compose screen provides powerful recipient search and selection. Recipients can be found by:

- **Name** — fuzzy search by member name
- **Gender identity** — filter by gender identity field
- **Program** — filter by program (Adults, Kids, Youth)
- **Class attended (specific session)** — select a specific past session (e.g., "Monday 6 PM Nuts & Bolts on 3 Mar") to target everyone who attended that session. Useful for found items, incidents, or follow-ups.
- **Class attended (recurring)** — select a class type (e.g., "all Nuts & Bolts attendees in the last 30 days") to target regular attendees. Useful for schedule changes.
- **Status** — filter by Active, Inactive, Archived, Trial
- **Belt** — filter by current belt level

Filters can be combined (AND logic). Results populate a selection list with:

- **Select All** — select all filtered results
- **Select None** — clear selection
- **Invert Selection** — toggle selected/unselected
- **Individual toggle** — click to add/remove individual members
- **Recipient count** — always visible showing "N recipients selected"

#### User Stories

**US-8.2.1: Search recipients by name**
As an Admin, I want to search for a member by name so that I can quickly message an individual.

- *Given* I open the compose screen
- *When* I type "Marcus" in the recipient search
- *Then* matching members appear and I can select one or more

**US-8.2.2: Filter recipients by class attended (specific)**
As an Admin, I want to email everyone who attended Monday's Gi Express so that I can ask about a found mouthguard.

- *Given* I open the compose screen
- *When* I filter by class attended → "Monday 6 PM Gi Express, 3 Mar 2026"
- *Then* all members who checked into that specific session are listed
- *And* I can select all and compose my message

**US-8.2.3: Filter recipients by class (recurring)**
As an Admin, I want to email all regular Nuts & Bolts attendees so that I can notify them of a schedule change.

- *Given* I need to inform all Nuts & Bolts regulars about a room change
- *When* I filter by class → "Nuts & Bolts" with lookback "last 30 days"
- *Then* all members who attended any Nuts & Bolts session in the last 30 days are listed

**US-8.2.4: Filter recipients by program**
As an Admin, I want to email all Kids program parents so that I can notify them of term dates.

- *Given* I need to send term information
- *When* I filter by program → "Kids"
- *Then* all members in the Kids program are listed

**US-8.2.5: Multi-select with invert**
As an Admin, I want to select all members then deselect a few so that I can email almost everyone except a couple of people.

- *Given* I filtered 40 members
- *When* I click "Select All" then individually deselect 3 members
- *Then* 37 recipients are selected and the count shows "37 recipients selected"

**US-8.2.6: Invert selection**
As an Admin, I want to invert my selection so that I can quickly switch who's included/excluded.

- *Given* I have 5 members selected out of 40
- *When* I click "Invert Selection"
- *Then* the other 35 members become selected and the original 5 are deselected

#### 8.2.2 Compose & Draft

Email composition includes:

- **Subject line** — required
- **Body** — rich text / markdown editor
- **Recipient list** — from §8.2.1
- **Header/footer template** — auto-applied from configured template (see §8.2.5)

Messages (and their recipient lists) can be saved as **drafts** at any point. Drafts appear in the email management screen and can be resumed, edited, and sent later.

#### User Stories

**US-8.2.7: Compose and send an email**
As an Admin, I want to compose an email with a subject, body, and recipient list so that I can communicate with members via email.

- *Given* I have selected 15 recipients
- *When* I write a subject and body and click "Send"
- *Then* the email is delivered to all 15 recipients via Resend
- *And* each recipient also sees the message on their in-app dashboard with a notification badge

**US-8.2.8: Save as draft**
As an Admin, I want to save my partially composed email as a draft so that I can finish it later.

- *Given* I have written a subject and selected 20 recipients but haven't finished the body
- *When* I click "Save Draft"
- *Then* the email (including recipient list) is saved with status "draft"
- *And* I can resume editing it from the email management screen

**US-8.2.9: Resume a draft**
As an Admin, I want to resume a saved draft so that I can finish composing and send it.

- *Given* I saved a draft yesterday with 20 recipients
- *When* I open it from the email management screen
- *Then* the subject, body, and recipient list are restored exactly as I left them

#### 8.2.3 Scheduled Sending

Emails can be scheduled for a specific date and time. Scheduled emails can be cancelled or rescheduled before their send time.

#### User Stories

**US-8.2.10: Schedule an email**
As an Admin, I want to schedule an email for Friday at 5 PM so that members receive it at the right time.

- *Given* I have composed an email
- *When* I set the schedule to "Friday 8 Mar 2026, 5:00 PM NZDT" and click "Schedule"
- *Then* the email is queued for delivery at that time
- *And* it appears in the email list with status "scheduled" and the scheduled time

**US-8.2.11: Cancel a scheduled email**
As an Admin, I want to cancel a scheduled email so that it doesn't go out if plans change.

- *Given* I scheduled an email for Friday
- *When* I click "Cancel" on the scheduled email before Friday
- *Then* the email is cancelled and moves to "cancelled" status
- *And* no email is delivered

**US-8.2.12: Reschedule an email**
As an Admin, I want to reschedule an email to a different time so that I can adjust timing without recomposing.

- *Given* I scheduled an email for Friday 5 PM
- *When* I change the schedule to Saturday 9 AM
- *Then* the email is updated and will be delivered at the new time

#### 8.2.4 Email History & Search

All sent emails are stored and searchable. Each email record includes the subject, body, send date, sender, and full recipient list.

#### User Stories

**US-8.2.13: Search email history**
As an Admin, I want to search past emails by keyword so that I can find a specific communication.

- *Given* I sent 50 emails over the last 6 months
- *When* I search for "schedule change"
- *Then* I see all emails whose subject or body contains "schedule change", sorted by most recent

**US-8.2.14: View email with recipients**
As an Admin, I want to view a sent email and see who received it so that I have a complete record.

- *Given* I sent a "Grading Day" email to 35 members last week
- *When* I open that email from the history
- *Then* I see the full email content, send date, and a list of all 35 recipients

**US-8.2.15: Member views received emails**
As a Member, I want to see emails sent to me on my dashboard so that I don't miss communications even if I missed the email.

- *Given* Admin sent me an email about schedule changes
- *When* I log into my dashboard
- *Then* I see the message with a notification badge and can read the full content

#### 8.2.5 Header & Footer Templates

Admin can configure a default header and footer template that is automatically applied to all outgoing emails. The template supports:

- **Header**: logo, gym name, tagline, or custom HTML/markdown
- **Footer**: contact info, social links, unsubscribe link (required for compliance), address

Templates are versioned — changes only apply to emails sent after the update.

#### User Stories

**US-8.2.16: Configure email template**
As an Admin, I want to set a header and footer template so that all emails have consistent branding.

- *Given* I navigate to email template settings
- *When* I set the header to include the Workshop logo and tagline, and the footer to include the gym address, phone, and social links
- *Then* all future outgoing emails automatically include this header and footer

**US-8.2.17: Preview email with template**
As an Admin, I want to preview how my email will look with the header/footer applied so that I can verify the layout before sending.

- *Given* I am composing an email
- *When* I click "Preview"
- *Then* I see the email as the recipient will see it, with header and footer applied

**US-8.2.18: Update template without affecting past emails**
As an Admin, I want template changes to only affect future emails so that past email records remain accurate.

- *Given* I update the footer to include a new phone number
- *When* I view a previously sent email
- *Then* it still shows the old footer as it was sent

#### 8.2.6 Account Activation Email

When a new member is registered, an activation email is sent to their email address. The member must click the activation link to verify their email and set their password. Until activated, the account cannot log in.

#### User Stories

**US-8.2.19: Registration sends activation email**
As an Admin, I want new member registrations to trigger an activation email so that the member can verify their email and set up their account.

- *Given* I register a new member with email "marcus@email.com"
- *When* the registration is saved
- *Then* an activation email is sent to "marcus@email.com" with a secure, time-limited activation link
- *And* the account is in "pending_activation" status until the link is clicked

**US-8.2.20: Member activates account**
As a new Member, I want to click the activation link to set my password so that I can log in.

- *Given* I received an activation email
- *When* I click the link and set my password
- *Then* my account is activated, my email is verified, and I can log in

**US-8.2.21: Resend activation email**
As an Admin, I want to resend the activation email so that a member who missed it can still activate.

- *Given* a member's account is still pending activation
- *When* I click "Resend Activation"
- *Then* a new activation email is sent with a fresh link
- *And* the previous link is invalidated

**US-8.2.22: Activation link expires**
As an Admin, I want activation links to expire after 72 hours so that stale links can't be used.

- *Given* a member received an activation email 4 days ago
- *When* they click the link
- *Then* they see "Link expired — contact your gym to resend"

#### 8.2.7 Email Provider: Resend

All outgoing email is delivered via **[Resend](https://resend.com)** — a developer-first email API with an official Go SDK.

**Why Resend:**

| Criteria | Resend |
|----------|--------|
| **Go SDK** | Official: `github.com/resend/resend-go/v2` |
| **Batch sending** | Up to 100 emails per API call |
| **Scheduled sending** | ISO 8601 datetime or natural language; cancel/reschedule supported |
| **Idempotency** | Built-in idempotency keys to prevent duplicate sends |
| **Compliance** | GDPR and SOC 2 compliant |
| **Deliverability** | DKIM, SPF, DMARC support; dedicated IP option |
| **Pricing** | Free: 3,000/mo (100/day). Pro ($20/mo): 50,000/mo. More than sufficient for a BJJ gym with ~80–150 members |
| **Logging** | Full delivery event tracking (sent, delivered, bounced, opened) |

**Integration pattern:**

- Emails are composed in Workshop and sent via the Resend API
- Each sent email is stored locally in the `email` table with subject, body, sender, status, scheduled_at, sent_at, and resend_message_id
- Recipients are stored in `email_recipient` linking email_id to member_id
- Resend delivery events (delivered, bounced, opened) can be received via webhook for status tracking
- The Resend API key is stored as an environment variable (`WORKSHOP_RESEND_KEY`), never hardcoded
- Default sender: `Workshop Jiu Jitsu <noreply@workshopjiujitsu.co.nz>` (configurable via `WORKSHOP_RESEND_FROM` env var)
- Sending domain `workshopjiujitsu.co.nz` must be verified in the Resend dashboard (DNS records: DKIM, SPF, DMARC)
- Gym contact email: `info@workshopjiujitsu.co.nz` (for reply-to and footer)

**Batch limitations to note:**
- Batch API does not support `scheduled_at` — scheduled emails must be sent individually
- Batch API does not support attachments (not needed for our use case)
- For sends > 100 recipients, multiple batch calls are chained

### 8.3 Coach Observations

Private per-member notes written by Coach or Admin. Used for technique feedback, grading observations, and behavioural notes. **Not visible to the member.**

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-8.3.1: Add observation**
As a Coach, I want to add a private observation to a member's profile so that I can track their development.

- *Given* I notice a member tends to muscle techniques
- *When* I add an observation: "Tends to muscle techniques — focus on flow drilling"
- *Then* the note is saved on the member's profile, visible only to coaches and admin

**US-8.3.2: View observations from readiness list**
As a Coach, I want to see a member's observations from the grading readiness list so that I have context when proposing promotions.

- *Given* I am reviewing a member on the readiness list
- *When* I click to view their observations
- *Then* I see all past notes from all coaches, sorted by date

---

## 9. Member Management

### 9.1 Digital Onboarding (Waiver)

QR-driven sign-off acknowledging risks and responsibilities. The waiver is **not** a liability transfer — it is a reminder of inherent risks and member responsibilities (hygiene, conduct, injury reporting). Valid for 1 year.

**Access:** All roles

#### User Stories

**US-9.1.1: Guest completes waiver**
As a Guest, I want to scan a QR code and complete the waiver so that I can get on the mats in under 60 seconds.

- *Given* I scan the QR code at the front desk
- *When* I enter my name, email, and sign the waiver
- *Then* a lightweight Guest account is created and I can check in

**US-9.1.2: Waiver expires after 1 year**
As an Admin, I want waivers to expire after 1 year so that members re-acknowledge the risks periodically.

- *Given* a member signed their waiver 13 months ago
- *When* they check in
- *Then* they are prompted to re-sign the waiver before proceeding

### 9.2 Red Flag (Injury Toggle)

Member-managed toggle selecting an injured body part. Appears as a warning icon next to their name on the coach's attendance list. Active for 7 days.

**Access:** Admin — | Coach — (view) | Member ✓ | Trial ✓ | Guest —

#### User Stories

**US-9.2.1: Flag an injury**
As a Member, I want to flag that my knee is sore so that the coach knows to give me space without me making a scene.

- *Given* I have a sore knee
- *When* I toggle the "knee" red flag on my profile
- *Then* a red flag icon appears next to my name on the coach's attendance view
- *And* the flag auto-expires after 7 days

**US-9.2.2: Coach sees red flags**
As a Coach, I want to see red flag icons on the attendance list so that I can acknowledge injuries before class.

- *Given* 2 members have active red flags
- *When* I view today's attendance
- *Then* those members have a red flag icon with the body part indicated

### 9.3 Member List

All members split by program, with manual "Fee" and "Frequency" fields, injury indicators, waiver status, member status, and sizing information.

**Member Profile Sizes:** Each member profile includes optional sizing fields to assist with belt promotions and merchandising:
- **Belt size**
- **Gi size**
- **Rash top size**
- **T-shirt size**

**Access:** Admin ✓ (full) | Coach ✓ (view) | Member — | Trial — | Guest —

#### User Stories

**US-9.3.1: View member list**
As an Admin, I want to see all members organised by program so that I can manage the club.

- *Given* the club has 80 active members across 5 programs
- *When* I open the member list
- *Then* I see members grouped by program with their belt, status, fee, and waiver status

**US-9.3.2: Convert trial to member**
As an Admin, I want to convert a Trial to a full Member so that they get full access after speaking to the coach.

- *Given* a Trial has been training for 3 weeks
- *When* I change their role from Trial to Member
- *Then* they gain full Member access (training log, goals, theme requests, etc.)

### 9.4 Inactive Member Radar

List of members who haven't checked in for a configurable number of days. Feeds into the archive workflow.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-9.4.1: View inactive members**
As an Admin, I want to see which members haven't trained recently so that I can reach out before they leave.

- *Given* 8 members haven't checked in for 30+ days
- *When* I open the inactive member radar
- *Then* I see a list of 8 members with their last check-in date and days since
- *And* I can message, follow up, or archive directly from this view

**US-9.4.2: Configure inactivity threshold**
As an Admin, I want to set the number of days before a member is flagged as inactive so that I can tune it for my gym.

- *Given* the default threshold is 30 days
- *When* I change it to 21 days
- *Then* members who haven't checked in for 21+ days appear on the radar

### 9.5 Archive / Restore Members

Admin can archive members who haven't trained in a while. Archived members are hidden from all active views but data is preserved. Can be restored at any time.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-9.5.1: Archive inactive members**
As an Admin, I want to archive members who haven't trained in 3+ months so that the active list stays clean.

- *Given* 5 members haven't checked in for 90+ days
- *When* I select them and click "Archive"
- *Then* their status changes to Archived and they are hidden from kiosk search, member lists, and all active views

**US-9.5.2: Restore archived member**
As an Admin, I want to restore a previously archived member who just walked back in.

- *Given* a member was archived 6 months ago
- *When* I find them in the archived list and click "Restore"
- *Then* their status changes to Active and they reappear in all active views with their full history intact

### 9.6 Coach Management

Admin can add coaches, assign them to classes, and manage their access.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-9.6.1: Add a new coach**
As an Admin, I want to add a new coach so that they can manage classes and content.

- *Given* a new instructor has joined the gym
- *When* I create a Coach account with their details and assign them to Monday 6 PM and Wednesday 6 PM
- *Then* they can log in, launch kiosk mode, view attendance, and manage rotors for their assigned classes

### 9.7 Schedule & Holiday Management

Admin defines the recurring weekly schedule and holiday overrides.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-9.7.1: Configure schedule**
As an Admin, I want to define the weekly class schedule so that sessions resolve correctly.

- *Given* I am setting up the schedule
- *When* I add an entry: Monday 6:00 PM, Nuts & Bolts, 60 min, Coach: James
- *Then* the session appears on Monday's schedule and is available for check-in

**US-9.7.2: Add holiday**
As an Admin, I want to enter a holiday closure so that classes don't appear during that period.

- *Given* the gym is closed for Easter weekend (18–21 April)
- *When* I create a holiday entry with those dates and label "Easter Weekend"
- *Then* no sessions resolve for 18–21 April
- *And* a holiday notice auto-generates on the dashboard

**US-9.7.3: Confirm NZ school term**
As an Admin, I want to confirm NZ school term dates so that Kids/Youth classes activate.

- *Given* the system suggests Term 1: 5 Feb – 11 Apr
- *When* I review and confirm the dates
- *Then* Kids and Youth classes begin resolving for check-in from 5 Feb

---

## 10. Calendar & Goals

### 10.1 Club Calendar

A shared calendar showing club events and competitions. Each layer is **toggleable**.

**Calendar Layers:**

| Layer | Content | Default | Managed by |
|-------|---------|---------|------------|
| **Club Events** | Seminars, social events, gym closures, grading days | On | Admin |
| **Competitions** | Local and international events with registration status | On | Admin |
| **Program Rotor layers** | Auto-generated per program with a rotor | Off | System |
| **My Goals** | Personal training goals with progress | Off | Member (own) |

**Access:** Admin ✓ (manage) | Coach ✓ (view) | Member ✓ (view) | Trial ✓ (view) | Guest —

#### User Stories

**US-10.1.1: View club calendar**
As a Member, I want to see the club calendar so that I know about upcoming events and competitions.

- *Given* there is a grading day Saturday, a seminar next month, and Grappling Industries in 3 weeks
- *When* I open the club calendar
- *Then* I see all three events on the calendar

**US-10.1.2: Add club event**
As an Admin, I want to add a club event so that everyone knows about it.

- *Given* we are having an End of Year BBQ
- *When* I create an event: "End of Year BBQ — Saturday 14 Dec"
- *Then* it appears on everyone's calendar

**US-10.1.3: Add competition**
As an Admin, I want to add a competition with a registration link so that members can sign up.

- *Given* Grappling Industries Wellington is on 15 March
- *When* I create a competition event with the registration URL
- *Then* members see it on the calendar with "Interested" and "Registered" toggles

**US-10.1.4: See who's going to a competition**
As a Member, I want to see which teammates are going to a competition so that we can coordinate.

- *Given* 5 members have marked "Registered" for Grappling Industries
- *When* I view the competition event
- *Then* I see the list of registered teammates

### 10.2 Program Rotor Calendar View

Each program's rotor schedule is rendered as a toggleable layer on the calendar. Draws from §5 rotor data.

**Access:** Admin ✓ | Coach ✓ | Member ✓ (toggleable) | Trial — | Guest —

#### User Stories

**US-10.2.1: Toggle program rotor on calendar**
As a Member, I want to toggle on a program's rotor view so that I can see what themes are coming up.

- *Given* I am interested in Gi Express
- *When* I toggle on the "Gi Express Rotor" layer
- *Then* I see "Leg Lasso" this week and "Closed Guard" starting next week

**US-10.2.2: Compare program rotors**
As a Member, I want to see multiple program rotors at once so that I can see how they align.

- *Given* I train both NoGi Express and Feet-to-Floor
- *When* I toggle on both rotor layers
- *Then* I see both programs' themes overlaid on the calendar

**US-10.2.3: Coach views full curriculum**
As a Coach, I want to see all program rotors enabled at once so that I can plan across the week.

- *Given* I teach 3 programs
- *When* I enable all rotor layers on the calendar
- *Then* I see the complete curriculum map for the week

### 10.3 Personal Goals

Members can overlay personal training goals on the calendar. Goals have a target metric, a time period, and track progress.

Goals are created with:
- **Description**: free text (e.g., "50 rear naked strangles")
- **Target**: a numeric target (e.g., 50)
- **Unit**: what's being counted (submissions, hours, sessions)
- **Period**: start and end date
- **Progress**: manual increment/decrement, or auto-tracked where possible (hours from attendance, sessions from check-ins)

Goals appear as coloured bars spanning their period, with a progress indicator.

**Access:** Admin — | Coach — | Member ✓ | Trial — | Guest —

#### User Stories

**US-10.3.1: Create a personal goal**
As a Member, I want to add a personal goal so that I can track my targets on the calendar.

- *Given* I want to work on rear naked chokes
- *When* I create a goal: "50 rear naked strangles during April", target=50, unit=submissions, period=1 Apr – 30 Apr
- *Then* a coloured bar appears on my calendar spanning April with a progress indicator at 0/50

**US-10.3.2: Log progress manually**
As a Member, I want to log progress toward my goal after each session so that I stay motivated.

- *Given* I got 3 rear naked chokes in today's session
- *When* I increment my goal progress by 3
- *Then* the progress updates to 3/50 on the calendar bar

**US-10.3.3: Auto-tracked goal**
As a Member, I want my "hours" goals to auto-track from attendance so that I don't have to log manually.

- *Given* I set a goal: "20 hours of training in February", unit=hours
- *When* I check into sessions during February
- *Then* my flight time hours for February are automatically counted toward the 20-hour target

---

## 11. Advanced Study (Laboratory)

### 11.1 Technical Tagging

Every clip is tagged with **Action** (e.g., Sweep) and **Connection** (e.g., K-Guard → SLX) metadata. Uses the same Connection/Action taxonomy as theme requests (§6).

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-11.1.1: Tag a clip**
As a Member, I want to tag my clip with an Action and Connection so that it's discoverable in the library.

- *Given* I created a clip of a K-Guard entry to SLX
- *When* I tag it: Action = "Entry", Connection = "K-Guard → SLX"
- *Then* the clip is searchable by those tags

### 11.2 4-Up Mode

A 2×2 grid allowing four clips to loop simultaneously with unified playback controls. Used for comparative technique study.

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-11.2.1: Compare four clips**
As a Member, I want to watch four athletes perform the same K-Guard entry at once so that I can see universal mechanics.

- *Given* the library has multiple clips tagged "K-Guard Entry"
- *When* I open 4-Up mode and load 4 clips
- *Then* all four play simultaneously in a 2×2 grid with unified play/pause controls

### 11.3 Predictive Search

Search bar suggesting clips based on technical relationships (tags).

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-11.3.1: Predictive clip search**
As a Member, I want to search for "Connection: SLX to Back" and have the system suggest relevant clips.

- *Given* the library has clips tagged with various Connection paths
- *When* I search "SLX to Back"
- *Then* the system suggests clips with that Connection tag and populates a study grid

### 11.4 Research Journal

Private text notes below 4-Up mode for personal research observations about technique patterns.

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-11.4.1: Write research notes**
As a Member, I want to write private research notes while watching clips so that I can capture technique insights.

- *Given* I am in 4-Up mode watching 4 K-Guard entries
- *When* I type notes: "All four athletes post the same-side hand before inverting. The hip angle is consistent at ~45°."
- *Then* the note is saved as a private research journal entry, linked to the current clips

---

## 12. Competition & Performance

### 12.1 Advice Repository

Read-only section for coach-curated strategy guides and rule-set breakdowns.

**Access:** Admin ✓ | Coach ✓ (curate) | Member ✓ (read) | Trial — | Guest —

#### User Stories

**US-12.1.1: Read competition advice**
As a Member, I want to read the coach's advice on ADCC rules so that I don't give away points.

- *Given* the coach has written an ADCC rule-set breakdown
- *When* I open the Advice Repository
- *Then* I can read the guide

### 12.2 Scouting Pipeline

Members submit rival clips; Admin vets and publishes as official "Advice."

**Access:** Admin ✓ (vet/publish) | Coach ✓ | Member ✓ (submit) | Trial — | Guest —

#### User Stories

**US-12.3.1: Submit scouting clip**
As a Member, I want to submit a clip of a competitor I'll face so that the coach can provide strategy advice.

- *Given* I found video of my next opponent
- *When* I submit the clip via the scouting pipeline
- *Then* Admin reviews and can publish it as official advice

---

## 13. Business Operations

### 13.1 Xero/Bank Reconciliation

Auto-updates "Last Payment" and "Status" (Green/Amber/Red) by matching bank references.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-13.1.1: Auto-reconcile payments**
As an Admin, I want payments to auto-reconcile from bank data so that I can see who's paid without manual checking.

- *Given* a member's bank reference matches their account
- *When* a payment comes through
- *Then* their payment status updates to Green and "Last Payment" is updated

### 13.2 Program ROI Dashboard

Filter earnings by program (Adults vs. Kids) to track overhead coverage.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-13.2.1: Break-even progress**
As an Admin, I want to see a break-even progress bar so that I know when the month's rent is covered.

- *Given* the monthly overhead is $5,440
- *When* I view the ROI dashboard
- *Then* I see a progress bar showing how much of this month's overhead has been covered by revenue

### 13.3 Digital Wallet

Member-side view of reconciled receipts for tax purposes.

**Access:** Admin — | Coach — | Member ✓ | Trial — | Guest —

#### User Stories

**US-13.3.1: Download receipts**
As a Member, I want to download payment receipts so that I have records for tax purposes.

- *Given* I have made 6 payments this year
- *When* I open my digital wallet
- *Then* I can see and download receipts for each payment

### 13.4 Engagement ROI Formula

Track the health of the high-agency model:

```
ROI = (α × Attendance + β × Study) / DaysSinceLastCheckIn
```

- **Attendance** (A): Physical presence (flight time)
- **Study** (S): Clips tagged / Research Journal entries
- **DaysSinceLastCheckIn** (D): Recency factor
- **α, β**: Custom weights for physical vs. technical engagement

---

## 14. Data Privacy & Compliance

Workshop operates as both **Controller** and **Processor** under GDPR / NZ Privacy Act 2020. Full guidance in `PRIVACY.md`.

### 14.1 Audit Logging

Every security-relevant and data-access event is logged to an immutable, append-only audit log. See `PRIVACY.md §1.3` for the full event list.

**Access:** Admin ✓ (view) | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-14.1.1: View audit trail**
As an Admin, I want to view the audit trail so that I can investigate security events and demonstrate compliance.

- *Given* a member's profile was updated yesterday
- *When* I search the audit log for that member
- *Then* I see who viewed/updated the profile, when, and what changed

**US-14.1.2: Audit DevMode impersonation**
As an Admin, I want all DevMode impersonation events logged so that there is accountability for actions taken while impersonating.

- *Given* I impersonate a coach role
- *When* I perform actions as coach
- *Then* the audit log records both my real admin identity and the impersonated role

### 14.2 Consent Management

Granular, versioned consent records. No single "I agree to everything" checkbox.

**Access:** Admin ✓ (view all) | Coach — | Member ✓ (own, manage) | Trial ✓ (own) | Guest ✓ (waiver only)

| Consent Type | Required | Default |
|-------------|----------|---------|
| Terms of Service | Yes | Must accept |
| Liability Waiver | Yes | Must sign |
| Marketing Emails | No | **Unchecked** (opt-in) |
| Photo/Video Usage | No | **Unchecked** (opt-in) |
| Injury Data Collection | Yes | Must accept |

#### User Stories

**US-14.2.1: Granular consent at registration**
As a new Member, I want to give consent separately for each purpose so that I control how my data is used.

- *Given* I am registering
- *When* I reach the consent step
- *Then* I see separate checkboxes for ToS (required), waiver (required), marketing (optional, unchecked), and photo/video (optional, unchecked)

**US-14.2.2: Revoke marketing consent**
As a Member, I want to revoke my marketing consent at any time so that I stop receiving promotional emails.

- *Given* I previously opted into marketing
- *When* I toggle marketing consent off in my profile
- *Then* my consent record is updated with a `revoked_at` timestamp and takes effect immediately

**US-14.2.3: Re-consent on waiver update**
As a Member, I want to be prompted to re-sign the waiver when it changes so that my consent is always current.

- *Given* Admin updated the waiver from v1.0 to v2.0
- *When* I next log in or check in
- *Then* I am prompted to review and sign the new waiver before proceeding

### 14.3 Right to be Forgotten (Data Deletion)

Members can request deletion of their personal data. Admin processes the request.

**Access:** Admin ✓ (execute) | Coach — | Member ✓ (request) | Trial — | Guest —

#### User Stories

**US-14.3.1: Request data deletion**
As a Member, I want to request deletion of my personal data so that my right to be forgotten is respected.

- *Given* I no longer train at Workshop
- *When* I submit a data deletion request from my profile
- *Then* Admin receives a notification and has 30 days to process it

**US-14.3.2: Process deletion request**
As an Admin, I want to process a deletion request so that we comply with privacy law.

- *Given* a member has requested data deletion
- *When* I approve the request
- *Then* PII (name, email, phone, sizes) is anonymised, medical/injury data is hard deleted, attendance records are anonymised, payment records are retained for 7 years (IRD), and the deletion is logged in the audit trail

**US-14.3.3: Grace period**
As an Admin, I want a 30-day grace period before hard deletion so that accidental requests can be reversed.

- *Given* a deletion request was approved
- *When* 30 days have not yet passed
- *Then* the member is marked as "pending deletion" and data is still recoverable
- *And* after 30 days, anonymisation/deletion executes automatically

### 14.4 Data Export (Portability)

Members can export their own data at any time.

**Access:** Admin — | Coach — | Member ✓ | Trial — | Guest —

#### User Stories

**US-14.4.1: Export my data**
As a Member, I want to export all my data so that I can take it with me if I leave.

- *Given* I have been training for 2 years
- *When* I click "Export My Data" in my profile
- *Then* I receive a JSON or CSV file containing my profile, attendance history, training log, belt progression, consent records, and messages
- *And* the export is logged in the audit trail

### 14.5 Data Classification

All data in the system is classified by sensitivity level.

| Classification | Examples | Access | Retention |
|---------------|----------|--------|-----------|
| **Public** | Class schedule, program names | Anyone | Indefinite |
| **Internal** | Attendance counts, rotor themes | Coach, Admin | Indefinite |
| **Confidential** | Member name, email, belt, sizes | Member (own), Coach, Admin | Until deletion + 30 days |
| **Restricted** | Injuries, medical notes, observations | Coach, Admin only | Until deletion (hard delete) |
| **Financial** | Payment records, fee amounts | Admin only | 7 years (IRD) |

---

## Appendix A: Data Model

| Concept | Section | Storage | Description |
|---------|---------|---------|-------------|
| `Account` | §1 | accounts | User identity with role and password hash |
| `Member` | §1 | members | Profile: name, email, gender_identity, program_id, fee, frequency, status, belt_size, gi_size, rash_top_size, tshirt_size |
| `Program` | §1.3 | programs | Audience group: Adults, Kids, Youth. Determines class visibility, term structure, and comms targeting |
| `Class` | §1.3 | classes | Named session type within a program (Gi Express, Nuts & Bolts, etc.). Has duration and optional mat_hours_weight (default 1.0) |
| `Schedule` | §9.7 | schedules | Recurring weekly entry: day, time, class_id, coach_id, duration |
| `Term` | §1.3 | terms | NZ school term date ranges with manual confirmation |
| `Holiday` | §9.7 | holidays | Date ranges overriding schedule; auto-generates Notice |
| `Waiver` | §9.1 | waivers | Risk acknowledgement: member_id, version, content_hash, signed_at, ip_address. Re-prompt on version change |
| `Injury` | §9.2 | injuries | Red Flag body-part toggle, active 7 days |
| `Attendance` | §3.1 | attendance | Check-in record: member_id + class_id + date + time. Supports multi-session and un-check-in (soft delete). Mat hours = duration × class weight |
| `Notice` | §8.1 | notices | Unified notification: type (school_wide / class_specific / holiday), status (draft / published) |
| `Email` | §8.2 | emails | Composed email: subject, body_html, body_text, sender_id, status (draft/scheduled/sending/sent/cancelled/failed), scheduled_at, sent_at, resend_message_id, template_header_snapshot, template_footer_snapshot |
| `EmailRecipient` | §8.2 | email_recipients | Join table: email_id, member_id, delivery_status (pending/delivered/bounced/opened), resend_recipient_id |
| `EmailTemplate` | §8.2.5 | email_templates | Header/footer template: type (header/footer), content_html, version, created_by, created_at. Versioned — only latest applies to new sends |
| `ActivationToken` | §8.2.6 | activation_tokens | Account activation: account_id, token (secure random), expires_at, used_at. 72-hour expiry. One active token per account |
| `GradingRecord` | §4.6 | grading_records | Promotion history: belt, stripe, date, proposed_by, approved_by, method (standard/override). Ceremony handled outside system |
| `GradingConfig` | §4.1 | grading_config | Per-belt thresholds: mat hours (adults) or attendance % (kids), stripe count, grading mode toggle |
| `GradingProposal` | §4.6 | grading_proposals | Coach-proposed promotion: member, target belt, notes, status (pending/approved/rejected) |
| `EstimatedHours` | §3.4 | estimated_hours | Bulk-estimated mat hours: date range, weekly hours, source (estimate/self_estimate), status, overlap mode, note |
| `Goal` | §10.3 | goals | Member target: description, target, unit (submissions/hours/sessions), period, progress |
| `Milestone` | §3.3 | milestones | Admin-configured achievement (e.g., "100 classes") |
| `CoachObservation` | §8.3 | coach_observations | Private per-member notes from Coach or Admin |
| `BeltConfig` | §4.4 | belt_config | Belt/stripe icon config: belt_name, colour (hex or split pair), stripe_count, sort_order, age_range |
| `Rotor` | §5.1 | rotors | Versioned curriculum for a class: class_id, version, status (draft/active/archived), preview_enabled, created_by |
| `Theme` | §5.2 | themes | Concurrent category within a rotor: rotor_id, name (Standing/Guard/Pinning/etc.), hidden, sort_order |
| `Topic` | §5.3 | topics | Technique in a theme's queue: theme_id, name, description, duration (default 1 week), sort_order, last_covered_date |
| `TopicVote` | §6.1 | topic_votes | Member vote on a topic: topic_id, member_id, rotation_cycle. One per member per topic per cycle |
| `Clip` | §7.1 | clips | YouTube timestamp loop. Can be cross-linked to topics (not hard-coupled) |
| `ClipTopicLink` | §7.1 | clip_topic_links | Optional link from a clip to a topic for offline study |
| `Tag` | §11.1 | tags | Action/Connection metadata for clips (independent taxonomy from rotor themes) |
| `ResearchEntry` | §11.4 | research_entries | Private research journal notes from 4-Up mode |
| `CalendarEvent` | §10.1 | calendar_events | Club event or competition: title, type, dates, registration_url |
| `Advice` | §12.1 | advice | Coach-curated strategy guides |
| `Payment` | §13.1 | payments | Reconciled bank transactions. Retained 7 years (IRD). Never store raw card numbers — token only |
| `AuditLog` | §14.1 | audit_logs | Immutable, append-only event log: actor_id, actor_role, action, resource_type, resource_id, metadata (JSON), timestamp (UTC). No updates or deletes |
| `ConsentRecord` | §14.2 | consent_records | Granular consent: member_id, consent_type (terms/waiver/marketing/photo_video/injury_data), granted, granted_at, revoked_at, version, ip_address |
| `DeletionRequest` | §14.3 | deletion_requests | Member data deletion request: member_id, requested_at, approved_at, executed_at, status (pending/approved/executed/cancelled), approved_by |

---

## Appendix B: Implementation Priority

1. **§2–§3** — Kiosk, check-in (auto-select, multi-session, un-check-in), attendance, training log, mat hours (with weighting), estimated hours, historical attendance
2. **§4** — Belt progression (stripe inference, term-based grading with per-term reset, belt/stripe icons, proposals, admin overrides)
3. **§9** — Member management (waiver, red flags, member list with sizes, inactive radar, archive, schedule/holiday management, program assignment)
4. **§8** — Communication (notices, email system via Resend with recipient search/filtering, drafts, scheduling, templates, email history; account activation emails; coach observations) — *account activation (§8.2.6) should be woven into §9.1 registration flow*
5. **§5–§6** — Curriculum engine (versioned rotors with concurrent themes, topic queues, auto-advance, coach ownership, topic voting)
6. **§7** — Technical library & clips (cross-linked to topics, not hard-coupled)
7. **§10** — Calendar (events, competitions, rotor views, personal goals)
8. **§13** — Business operations (Xero, ROI dashboard, digital wallet)
9. **§11–§12** — Advanced study & competition tools
10. **§14** — Data privacy & compliance (audit log table, consent management, deletion/anonymisation, data export) — *cross-cutting: audit logging and consent checks should be woven into §2–§9 as they are built*

---

## Appendix C: Resolved Design Decisions

The following decisions were resolved during PRD review.

| # | Question | Decision | Notes |
|---|----------|----------|-------|
| 1 | Program hierarchy | **Three-tier: Program (Adults/Kids/Youth) → Class → Session.** Members belong to one program. Classes show by default to members in that program. Kids/Youth age up via Admin reassignment. | §1.3 updated |
| 2 | "Session" term overload | **"Session" = time-slot only.** "Attendance count" for kids grading. Extra Session feature removed. | §1.4, §12 updated |
| 3 | Mat hours terminology | **"Mat hours" is canonical.** "Flight time" is the member-facing display name. Data model uses `mat_hours`. | §1.4 updated, refs throughout |
| 4 | Student vs Member | **Standardise on "Member" everywhere.** | §1.1 |
| 5 | Rotor model / voting scope | **Concurrent theme categories** (Standing, Guard, Pinning, In-Between Game) with topic queues. **Members vote on topics**, not themes. Successful vote bumps the scheduled topic; bumped topic stays in queue for next rotation. | §5, §6 rewritten |
| 6 | TrainingGoal vs PersonalGoal | **Merged into single "Goal" concept.** | §10.3, Appendix A |
| 7 | Kids term grading | **Per-term with reset.** Must hit threshold each term. Admin has ultimate discretion. | §4.3 updated |
| 8 | Rotor cardinality | **1:N with versioning.** Multiple rotor versions per class for drafting improvements. One active at a time. | §5.5 |
| 9 | Connection/Action taxonomy sharing | **Separate data, same terminology.** Clip tags (§11) and rotor themes (§5) are independent. Cross-links from schedule to clips for offline study, but not hard-coupled. | §7.1, Appendix A |
| 10 | Mat time weighting | **Equal by default (1.0×), Admin can configure per-class multipliers.** | §3.1 |
| 11 | Rotor advancement | **Auto-advance by default** when topic duration expires. Coaches can extend, skip, or manually advance. | §5.4 |
| 12 | Grading vs Promotion terminology | **"Belt Progression" = the system.** "Promotion" = the result. Ceremony handled outside the system. Added belt size, gi size, rash top size, t-shirt size to member profiles. | §4, §9.3 |
| 13 | Who can advance rotors | **Coaches edit/advance for classes they own.** Admin can override everything. | §5.4 |
| 14 | Displaced (bumped) topic | **Voted topic inserted before scheduled topic.** Bumped topic stays in its queue position for next rotation cycle. | §6.2 |
| 15 | Email provider | **Resend** (`resend.com`). Official Go SDK, batch + scheduled sending, GDPR/SOC 2 compliant, free tier sufficient for small gyms, Pro ($20/mo) for growth. API key via env var, never hardcoded. | §8.2.7 |
| 16 | Registration requires activation | **Yes.** Registration sends activation email; account is `pending_activation` until link clicked. 72-hour token expiry. Admin can resend. | §8.2.6 |
| 17 | Email vs in-app messaging | **Both.** Emails are delivered via Resend AND mirrored as in-app dashboard messages. Email is the primary channel; in-app is the fallback. Old `Message` concept replaced by `Email` + `EmailRecipient`. | §8.2 |
| 18 | Gender identity on member profile | **Added** as optional field for recipient filtering. Not displayed publicly; used only for targeted communications. | §8.2.1, Appendix A |
| 19 | List view pattern | **Server-side pagination, column sorting, search/filters, and configurable row count (10/20/50/100/200) on all list views.** URL query string encodes full state (bookmarkable). Default 20 rows. Single-column sort. Debounced search. Row count persisted in localStorage per-view. | §1.5 |
