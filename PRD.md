# Workshop Jiu Jitsu — Product Requirements Document

A **Modular Evolution** approach where each layer is a self-contained, functional system adding specific value while maintaining the "Minimalist/Technical" philosophy.

---

## Roles

All features are gated by role. Every account belongs to exactly one role.

| Role | Description |
|------|-------------|
| **Admin** | Gym owner/operator. Full system access including schedule, billing, grading, and user management. |
| **Coach** | Instructor. Can manage check-ins, view attendance, manage kiosk mode, propose gradings, and curate content. |
| **Member** | Active paying student. Can check in, flag injuries, view schedule, track training hours, and access study tools. |
| **Trial** | Prospective student. Can check in, sign waiver, and view schedule. No hard visit limit — Admin manually converts to Member when ready. After each check-in, the system displays a prompt: *"Enjoying Workshop? Talk to your coach about signing up!"* |
| **Guest** | Drop-in visitor. A lightweight account (name + email) is created during the waiver flow. If they return, they are recognised and prompted to convert to Trial or Member. |

### Member Statuses

Every Member/Trial record has one of three statuses:

| Status | Description |
|--------|-------------|
| **Active** | Currently training. Appears in kiosk search, member lists, and all active views. |
| **Inactive** | Has stopped checking in (flagged after a configurable number of days). Still visible in member lists with an "inactive" indicator. Candidate for follow-up or archive. |
| **Archived** | Manually archived by Admin. Hidden from all active views and kiosk search. All data preserved. Can be restored to Active at any time. |

---

## Role Priorities

What matters most to each role — these priorities drive the default dashboard and navigation for each experience.

### Member

| Priority | What they need | Why |
|:--------:|----------------|-----|
| 1 | **Class Timetable** | Know when to show up — today's classes, this week's schedule, holiday closures |
| 2 | **Check-In** | Fast kiosk check-in with auto-selected session, multi-session support, and un-check-in capability |
| 3 | **Training Log** | Personal training hours (flight time), classes attended, streaks, grading progress with belt/stripe icons, and milestone achievements |
| 4 | **Club Notices** | Announcements from admin/coaches: schedule changes, events, gym news, holiday closures |
| 5 | **Theme Requests & Voting** | Request techniques to cover (connections, actions), vote on others' requests, see when topics were last covered |
| 6 | **Club Calendar & Personal Goals** | View club events, competitions, and program rotor schedules; overlay personal training goals (e.g., "50 RNCs in April") |
| 7 | **Estimated Training Time** | Submit estimates for unrecorded training periods (flagged for admin review) so flight time stays accurate |

### Coach

| Priority | What they need | Why |
|:--------:|----------------|-----|
| 1 | **Class Timetable** | See their classes for today/this week, know which program and time slot |
| 2 | **Rotor & Themes** | Current rotor position for each program, this week's topics, ability to control rotor advancement |
| 3 | **Injury Flags** | Red flags on today's attendance — know who needs modifications before class starts |
| 4 | **Class Notices** | Pre-set announcements to make at the beginning or end of class (set by admin or self) |
| 5 | **Grading Readiness List** | Members approaching grading eligibility with stripe progress, belt icons, and areas to work on |
| 6 | **Coach Observations** | Private per-student notes: technique feedback, grading observations, behavioural notes |
| 7 | **Estimated Training Hours** | Bulk-add estimated hours for members with poor records by specifying a date range and weekly hours |
| 8 | **Theme Request Triage** | Review member-voted theme requests and choose to bring them forward in the rotor |

### Admin

| Priority | What they need | Why |
|:--------:|----------------|-----|
| 1 | **Rotor & Program Configuration** | Create and manage rotors per program, define themes and topics with frequency weights, control rotor advancement, toggle student preview |
| 2 | **Content Configuration** | Full control over schedules, themes, notices, holidays, club calendar events — everything the gym runs on |
| 3 | **Grading & Estimated Hours** | Review estimated-hours submissions (from coaches and students), approve/adjust, manage stripe inference and term-based grading |
| 4 | **Notices & Coach Reminders** | Set school-wide notices (visible to all) and class-specific reminders (visible before/after class) |
| 5 | **Direct Messaging** | Message individual students — follow-ups, payment reminders, personal check-ins |
| 6 | **Inactive Member Radar** | See who hasn't checked in for a while — proactive retention before they ghost |
| 7 | **Club Calendar Management** | Manage club events, competitions, and toggle program rotor views on the shared calendar |
| 8 | **Coach Management** | Add new coaches, assign them to classes, manage their access |
| 9 | **Archive / Restore Members** | Archive members who haven't trained in a while (keeps data, removes from active lists). Restore when they return. |

---

## Feature Access by Role

| Feature | Admin | Coach | Member | Trial | Guest |
|---------|:-----:|:-----:|:------:|:-----:|:-----:|
| **Layer 1a: Core Operations** | | | | | |
| Kiosk Mode (launch/exit) | ✓ | ✓ | — | — | — |
| Kiosk Check-In (auto-select session) | — | — | ✓ | ✓ | ✓ |
| Multi-Session Check-In / Un-Check-In | — | — | ✓ | ✓ | — |
| Program Schedule Management | ✓ | — | — | — | — |
| Holiday Schedule Management | ✓ | — | — | — | — |
| View Schedule / Dashboard | ✓ | ✓ | ✓ | ✓ | ✓ |
| Digital Onboarding (Waiver) | ✓ | ✓ | ✓ | ✓ | ✓ |
| Red Flag (Injury Toggle) | — | — | ✓ | ✓ | — |
| View Attendance + Red Flags | ✓ | ✓ | — | — | — |
| Historical Attendance (other days) | ✓ | ✓ | — | — | — |
| Member List + Management | ✓ | View | — | — | — |
| Role Management | ✓ | — | — | — | — |
| **Layer 1b: Engagement** | | | | | |
| Training Log (own history) | — | — | ✓ | ✓ | — |
| Training Goals | — | — | ✓ | — | — |
| Milestones & Achievements | ✓ (configure) | View | View own | — | — |
| Belt & Stripe Icons | ✓ (configure) | View | View own | View own | — |
| Notices (school-wide) | Publish | Draft | View | View | View |
| Notices (class-specific) | Publish | Draft | — | — | — |
| Notices (holiday — auto) | ✓ (configure) | — | View | View | View |
| Direct Messaging | ✓ | — | View own | — | — |
| Grading System (stripe inference) | ✓ (approve/override) | Propose | View own | View own | — |
| Term-Based Grading (kids toggle) | ✓ (configure) | View | View own | — | — |
| Grading Readiness List | ✓ | ✓ | — | — | — |
| Estimated Training Hours (bulk add) | ✓ | ✓ | Submit own | — | — |
| Coach Observations (private) | ✓ | ✓ | — | — | — |
| Inactive Member Radar | ✓ | ✓ | — | — | — |
| Archive / Restore Members | ✓ | — | — | — | — |
| Coach Management | ✓ | — | — | — | — |
| **Layer 2: Spine** | | | | | |
| Program Rotors (create/manage) | ✓ | Control | View (if preview on) | — | — |
| Rotor Themes & Topics | ✓ (create) | ✓ (create) | View (if preview on) | — | — |
| Theme Requests & Voting | ✓ (triage) | ✓ (triage) | ✓ | — | — |
| Hidden / Surprise Themes | ✓ (create) | ✓ (create) | — | — | — |
| Clipping Tool | ✓ | ✓ | ✓ | — | — |
| Technical Library | ✓ | ✓ | ✓ | — | — |
| Promote Clip | ✓ | ✓ | — | — | — |
| **Layer 3: Laboratory** | | | | | |
| Technical Tagging | ✓ | ✓ | ✓ | — | — |
| 4-Up Mode | ✓ | ✓ | ✓ | — | — |
| Predictive Search | ✓ | ✓ | ✓ | — | — |
| Research Journal | ✓ | ✓ | ✓ | — | — |
| **Layer 4: War Room & Calendar** | | | | | |
| Club Calendar (events, competitions) | ✓ (manage) | ✓ | ✓ | ✓ | — |
| Program Rotor Calendar View | ✓ | ✓ | ✓ (toggleable) | — | — |
| Personal Goal Calendar Layer | — | — | ✓ | — | — |
| Advice Repository | ✓ | ✓ | ✓ | — | — |
| Extra Session RSVP | ✓ | ✓ | ✓ | — | — |
| Scouting Pipeline | ✓ | ✓ | Submit | — | — |
| **Layer 5: Optimization** | | | | | |
| Xero/Bank Reconciliation | ✓ | — | — | — | — |
| Program ROI Dashboard | ✓ | — | — | — | — |
| Digital Wallet (own receipts) | — | — | ✓ | — | — |

---

## Architecture Summary

| Layer | Priority | Primary UI | Key Tech |
|-------|----------|------------|----------|
| **1a. Core Operations** | Safety/Ops | Tablet (Kiosk + Check-in) | Database/Auth/Roles, Session auto-select |
| **1b. Engagement** | Retention/Culture | Dashboard (Member + Coach) | Grading/Stripe inference, Belt icons, Estimated hours |
| **2. Spine** | Technical Alignment | Mobile (Library + Rotor) | Rotor engine, Theme requests/voting, YouTube API |
| **3. Laboratory** | Research/Logic | Desktop/Tablet (2x2) | YouTube API (v2) / Tagging |
| **4. War Room & Calendar** | Performance/Planning | Mobile (Calendar + Hub) | Calendar layers, Personal goals, Competition |
| **5. Optimization** | Sustainability | Admin (Dashboard) | Xero API |

---

## Layer 1a: Core Operations

**Goal:** Replace paper trails, ensure facility safety, and provide the daily operating backbone — kiosk, check-in, scheduling, waivers, and injury tracking.

### Features

| Feature | Description |
|---------|-------------|
| **Kiosk Mode** | A locked check-in screen for the front-desk tablet. Launched by a Coach or Admin from their own logged-in account. The kiosk is **tied to the account that launched it** — only that person's password can exit it. If a different Coach/Admin needs to exit, they must first authenticate with their own credentials. Prevents members from navigating away. |
| **Check-In by Name Search** | The **only** way to check in is by typing your name. The system performs a fuzzy search and presents a shortlist of matching Active members as the user types. The member selects their name from the suggestions to proceed to class selection. No member ID, email, or barcode is ever required. Only Active members appear in results (Inactive and Archived are hidden). This applies to both the kiosk and the standard check-in form. |
| **Session Auto-Select** | After selecting their name, the system auto-selects the session closest to the current time. Multiple sessions run each day (see Programs below). The member can confirm the auto-selected session, change to a different session, or select multiple sessions to check into at once. |
| **Multi-Session Check-In** | Members can check into more than one session in a single visit (e.g., "6 PM Nuts & Bolts" + "7:15 PM Feet-to-Floor"). Each session is recorded as a separate attendance entry. |
| **Un-Check-In** | If a member checked into the wrong session by mistake, they can undo their check-in from the same kiosk screen. Un-check-in is only available for sessions on the current day. |
| **Historical Attendance View** | Admin and Coach can navigate to view attendance for other days using a date picker or prev/next day navigation. A prominent "Back to Today" button allows quick return to the current day's sessions. |
| **Program Schedule Management** | Admin defines the recurring weekly class schedule for each program. Programs have two tiers: a top-level **Program** (Adults, Kids — determines term dates and billing) and **class types** within it (Fundamentals, No-Gi, Competition, etc.). Some programs (e.g., Kids) run aligned to NZ school terms, but term start/end dates must be manually confirmed by Admin each term — the system suggests NZ term dates but does not auto-activate. |
| **Holiday Schedule** | Admin can enter holiday dates that override the default schedule (e.g., Christmas break, public holidays, gym closures). Holidays auto-generate a notice on the dashboard so members know when the gym is closed or running a modified timetable. |
| **Digital Onboarding (Waiver)** | QR-driven sign-off acknowledging risks and responsibilities as a member or guest at Workshop. The waiver is **not** a liability transfer — it is a reminder of the inherent risks of martial arts training and the member's responsibilities (hygiene, conduct, injury reporting). Valid for 1 year. For Guests, completing the waiver flow creates a lightweight account (name + email). |
| **The "Red Flag"** | Member-managed toggle selecting an injured body part — appears as a warning icon next to their name on the Coach's attendance list. Active for 7 days. |
| **Member List** | All members split by Program and class type, with manual "Fee" and "Frequency" fields, injury indicators, waiver status, and current member status (active/inactive/archived). |

### Kiosk Mode Detail

The kiosk is the primary interface for day-to-day gym operations:

1. **Coach/Admin opens Kiosk Mode** — authenticates with their own password; tablet locks into kiosk view tied to that account
2. **Member walks up** — types their name into the search bar. The system performs a **fuzzy search** and presents a **shortlist of matching Active members** as the user types (autocomplete). No member ID, email, or barcode is required. Only Active members appear; Inactive and Archived are hidden.
3. **Member selects their name from the shortlist** — system resolves today's sessions for their program on-the-fly (Schedule + Terms − Holidays) and **auto-selects the session closest to the current time**
4. **Member confirms or adjusts** — the auto-selected session is highlighted. The member can confirm, switch to a different session, or **select multiple sessions** (e.g., "6 PM Nuts & Bolts" + "7:15 PM Feet-to-Floor"). Each selected session creates a separate attendance record.
5. **Un-Check-In** — if the member tapped the wrong session, they can undo their check-in for any session on the current day from the same kiosk screen
6. **Trial member sees prompt** — after check-in, a message displays: *"Enjoying Workshop? Talk to your coach about signing up!"*
7. **Coach/Admin exits Kiosk Mode** — the person who launched the kiosk enters their password to exit. A different Coach/Admin must authenticate with their own credentials first.

For **Guests**: a "Guest Check-In" button on the kiosk launches the waiver flow (creating a lightweight account with name + email), then records attendance. If the guest is recognised from a previous visit, their existing record is used and they are prompted to convert to Trial or Member.

### Historical Attendance

Admin and Coach can view attendance for days other than today:

- **Date navigation**: a date picker or prev/next arrows allow browsing to any past date
- **Back to Today**: a prominent button always visible to return to the current day's live view
- **Read-only past**: past attendance cannot be modified from this view (only Admin can adjust via member profiles)
- **Same layout**: historical view uses the same attendance layout (session list, check-in counts, injury flags) as the live view

### Programs & Schedule Detail

**Programs:**

Workshop runs multiple programs, each of which can have its own independent rotor (see Layer 2). Programs are extensible by Admin.

| Program | Description | Duration | Audience |
|---------|-------------|----------|----------|
| **Nuts & Bolts** | Beginner skills development (Gi) | 60 min | Adults — all levels welcome |
| **Gi Express** | All-levels skills development and sparring (Gi) | 60 min | Adults — all levels |
| **NoGi Express** | All-levels skills development and sparring (NoGi) | 60 min | Adults — all levels |
| **Feet-to-Floor** | Skills for taking the match to the ground | 60 min | Adults — all levels |
| **NoGi Long Games** | Specific sparring and competition prep (NoGi) | 60 min | Adults — experience required, prior arrangement with coaches |
| **BYO Task Based Games** | Practice game design and specific skills (Gi or NoGi) | 60 min | Adults — experience recommended |
| **Open Mat** | Gi or NoGi sparring | 60–180 min | Adults — experience required |
| **Workshop Kids** | Kids program | 45 min | Ages 6–11, runs during NZ school terms |
| **Workshop Youth** | Youth program | 45–60 min | Ages 12–16, runs during NZ school terms |

**Default Weekly Timetable (extracted from workshopjiujitsu.co.nz):**

| Day | Time | Session | Duration |
|-----|------|---------|----------|
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

**NZ School Terms (default suggestions):**

Kids and Youth programs are term-aligned. The system suggests NZ school term dates but Admin must manually confirm each term before the schedule activates.

| Term | Typical Dates |
|------|--------------|
| Term 1 | Early February – mid-April |
| Term 2 | Late April/early May – early July |
| Term 3 | Mid/late July – late September |
| Term 4 | Mid-October – mid-December |

**Holidays:** Admin creates holiday entries (date range, label, optional message). During a holiday, affected classes are excluded from the on-the-fly class resolution and a holiday notice is auto-generated on the dashboard (e.g., *"Gym closed 23 Dec – 7 Jan: Summer Break"*).

**Session Resolution:** Class sessions are **not** pre-generated. When the kiosk or dashboard loads "today's classes," the system resolves `Schedule + Terms − Holidays` on-the-fly to determine which sessions are available. The session closest to the current time is auto-highlighted for check-in. No rows are stored until a member checks in.

### User Stories (Layer 1a)

| Role | Story |
|------|-------|
| Guest | Scan a QR code, enter my name and email, sign the waiver acknowledging risks and responsibilities, and get on the mats in under 60 seconds |
| Guest | Return for a second visit and have the kiosk recognise me — prompted to talk to a coach about becoming a Trial or Member |
| Member | Walk up to the kiosk at 6:05 PM, type my name, and "6 PM Nuts & Bolts" is already auto-selected — I confirm and I'm checked in in 3 seconds |
| Member | I'm staying for back-to-back sessions, so I select both "6 PM Nuts & Bolts" and "7:15 PM Feet-to-Floor" at check-in |
| Member | I accidentally checked into the wrong session — I tap "un-check-in" and select the correct one |
| Member | Flag that my knee is sore so the coach knows to give me space without me making a scene |
| Trial | Check in at the kiosk just like a member, and see a friendly prompt after: "Enjoying Workshop? Talk to your coach about signing up!" |
| Coach | See a red flag icon next to a student's name on today's attendance so I can acknowledge their injury |
| Coach | Launch kiosk mode from my account and know only my password (or another coach logging in) can exit it |
| Coach | Browse to last Wednesday's attendance to check who came to NoGi Long Games, then hit "Back to Today" to return |
| Admin | Set up the Kids term schedule by confirming the NZ Term 1 dates, and classes auto-populate on the calendar |
| Admin | Enter a holiday closure for Easter weekend and know the dashboard will inform members via auto-generated notice |
| Admin | Convert a Trial to a full Member after they've spoken to the coach and are ready to commit |

---

## Layer 1b: Engagement

**Goal:** Retain members, support grading decisions, and keep the club informed — training logs, grading, notices, messaging, and coach tools.

### Features

| Feature | Description |
|---------|-------------|
| **Training Log** | A projection over the member's Attendance data showing: classes attended, total mat hours ("flight time") split into recorded and estimated hours, streaks, date of last check-in, belt/stripe icon, and grading progress. Members can set personal training goals (e.g., "3× per week"). Admin can configure milestones (e.g., "100 classes", "1 year streak") that display as achievements. |
| **Grading System** | Tracks each member's current belt and progress toward their next grade with visual belt/stripe icons. **Adults:** grading eligibility is based on accumulated "flight time" (total mat hours); stripes are inferred by pro-rating total hours across the belt level. **Kids:** a toggle allows using sessions attended instead of hours, with % of term attendance as the eligibility criterion. Admin can configure thresholds per belt level. Admin can override by adding time/classes or forcing a promotion. Coaches **propose** promotions (creating a pending record); Admin **approves** to make them official. Members can view their own training hours, grading history, and are informed that grading requires minimum flight times — *unless they are competing and sandbagging their division, at Admin's discretion.* |
| **Belt & Stripe Icons** | Visual CSS/SVG icons for all belt levels. **Adults:** White (4 stripes) → Blue (4) → Purple (4) → Brown (4) → Black (degrees per IBJJF). **Kids (IBJJF):** White → Grey/White → Grey → Grey/Black → Yellow/White → Yellow → Yellow/Black → Orange/White → Orange → Orange/Black → Green/White → Green → Green/Black → Blue. Each belt displays the member's current stripe count. Icons appear on training logs, grading readiness lists, member profiles, and dashboards. |
| **Grading Readiness List** | Auto-generated list of members approaching grading eligibility thresholds. Shows belt/stripe icon, flight time progress (adults) or term attendance % (kids). Coach can view readiness, add grading notes, and propose promotions from this view. |
| **Estimated Training Hours** | For members with incomplete records, Admin or Coach can bulk-add estimated flight time by selecting a date range (start/end via calendar picker) and entering estimated weekly training hours. The system calculates total estimated hours for the period. If the period overlaps existing recorded training, Admin chooses: **replace** the recorded hours in that overlap, or **add** the estimate on top of existing training blocks. Estimated hours are delineated from recorded hours throughout the system (shown separately in training logs, flagged in grading calculations). |
| **Student Self-Estimates** | Members who have been slack with check-ins can submit their own estimated training period (date range + weekly hours + a required note describing the context, e.g., "Trained at a partner gym while travelling"). Self-estimates are flagged for Admin review. Admin can approve as-is, adjust the hours, or reject. Approved estimates count toward flight time but remain marked as "estimated." |
| **Notices** | A unified notification system with three types: **school_wide** (general club announcements), **class_specific** (reminders for a coach to deliver at start/end of a particular class), and **holiday** (auto-generated from Holiday entries). Coaches can **draft** notices; Admin **publishes** them. Admin can also publish directly. All notices appear contextually: school-wide on dashboards, class-specific on the coach's class view, holidays on the dashboard and kiosk. |
| **Direct Messaging** | Admin can send messages to individual members — payment follow-ups, personal check-ins, welcome messages. In-app only in v1 (message appears on the member's dashboard with a notification badge). Email and SMS delivery planned for later. |
| **Coach Observations** | Private per-student notes written by Coach or Admin. Used for technique feedback, grading observations, behavioural notes. Not visible to the student. Accessible from the member's profile and from the Grading Readiness List. |
| **Inactive Member Radar** | List of members who haven't checked in for a configurable number of days. Visible to Admin and Coach. Feeds into the Admin's archive workflow — Admin can follow up, message, or archive directly from this view. |
| **Archive / Restore Members** | Admin can archive members who haven't trained in a while — changes their status to Archived, hiding them from all active views and kiosk search. All data is preserved. Archived members can be restored to Active at any time. |
| **Coach Management** | Admin can add new coaches, assign them to classes on the schedule, and manage their access. |

### Grading System Detail

**Belt Progression (IBJJF-aligned):**

**Adults (18+ years):**

White → Blue → Purple → Brown → Black. Each belt has 4 stripes before promotion eligibility. Black belt has degrees (1st–6th degree, then coral belts per IBJJF).

| Belt | Min. Time at Belt | Stripes | Default Flight Time Threshold |
|------|-------------------|---------|-------------------------------|
| White → Blue | 2 years | 4 | 150 hours |
| Blue → Purple | 1.5 years | 4 | 300 hours |
| Purple → Brown | 1.5 years | 4 | 500 hours |
| Brown → Black | 1 year | 4 | 750 hours |

**Kids (4–15 years, IBJJF):**

White → Grey/White → Grey → Grey/Black → Yellow/White → Yellow → Yellow/Black → Orange/White → Orange → Orange/Black → Green/White → Green → Green/Black → Blue (at 16+).

Each kids belt has up to 4 stripes. Configurable by Admin.

**Youth (16–17 years, IBJJF):**

May hold Purple or Brown belt (up to 4 stripes each). Transitions to the adult track at 18.

**Stripe Inference (Adults):**

Stripes are automatically inferred by pro-rating accumulated flight time across the configured threshold for the current belt level. No manual stripe awards are needed — the system calculates the member's stripe count from their hours.

| Example | Threshold | Hours Trained | Inferred Stripe |
|---------|-----------|---------------|-----------------|
| Blue belt, 150h threshold | 150h | 0–37h | Stripe 0 |
| Blue belt, 150h threshold | 150h | 38–74h | Stripe 1 |
| Blue belt, 150h threshold | 150h | 75–111h | Stripe 2 |
| Blue belt, 150h threshold | 150h | 112–149h | Stripe 3 |
| Blue belt, 150h threshold | 150h | 150h+ | Stripe 4 (eligible for promotion) |

Formula: `stripe = floor(hours / (threshold / stripe_count))`

**Term-Based Grading (Kids/Youth Toggle):**

For term-based programs (Kids, Youth), Admin can toggle the grading metric between:

| Mode | Metric | Example |
|------|--------|---------|
| **Sessions** (default for kids) | Number of sessions attended in the current term | "Attended 24 of 30 sessions this term" |
| **Hours** | Accumulated mat hours (same as adults) | "45 hours since last grading" |

When in **Sessions** mode, eligibility is based on **% of term attendance**:
- The system counts total available sessions for the member's program in the current NZ school term
- The member's attendance count is divided by total available sessions
- Admin configures the minimum % required per belt level (e.g., "80% attendance this term")

**Estimated Training Hours Detail:**

Admin/Coach bulk-add flow:
1. Select the member from the member list or profile
2. Choose a date range (start/end via calendar picker)
3. Enter estimated weekly training hours (e.g., "4 hours/week")
4. System calculates: `total = weeks × weekly_hours`
5. If the period overlaps existing recorded training blocks, Admin sees a warning and chooses:
   - **Replace**: remove recorded hours in the overlap period and substitute with the estimate
   - **Add**: keep recorded hours and add estimated hours on top
6. Estimated hours are tagged as `source: estimate` with the admin's ID and a timestamp

Student self-estimate flow:
1. Member submits: date range + estimated weekly hours + **required note** (e.g., "Trained 3×/week at Checkmat São Paulo while travelling")
2. Estimate appears in Admin's review queue with status `pending`
3. Admin can: **approve** (as-is), **adjust** (modify hours), or **reject** (with optional reason)
4. Approved estimates are tagged as `source: self_estimate` and count toward flight time

**Belt & Stripe Icons:**

Visual representations rendered as CSS/SVG showing:
- Belt colour (solid or split for kids dual-colour belts like Grey/White)
- Stripe count (0–4 small markers on the belt tab)
- Black belt degree markers (per IBJJF)

Icons are displayed wherever a member's rank is shown: training logs, grading readiness, member profiles, dashboards, and attendance lists.

**Admin Overrides:** Admin can adjust a member's grading eligibility by:
- Adding flight time or class credits manually
- Forcing an immediate promotion (bypassing thresholds)
- Adjusting thresholds per individual (e.g., competitor fast-tracking)

**Sandbagging Exception:** Members who are actively competing may be promoted ahead of schedule at Admin's discretion if they are sandbagging their division (competing below their skill level).

**Member Visibility:** Members see their own training hours (split into recorded vs. estimated), current belt with stripe icon, grading history, and a progress indicator toward the next grade. A note is displayed: *"Grading requires minimum flight time. Exceptions may apply for active competitors."*

### Notice System Detail

All notifications flow through a single `Notice` concept with a `type` field:

| Type | Created by | Published by | Visible to | Where shown |
|------|-----------|-------------|------------|-------------|
| `school_wide` | Admin or Coach (draft) | Admin | All roles | Dashboard, Kiosk |
| `class_specific` | Admin or Coach (draft) | Admin | Coach (assigned to that class) | Coach's class view |
| `holiday` | System (auto from Holiday entries) | Auto | All roles | Dashboard, Kiosk |

Coach-created notices start in **draft** state and require Admin approval to publish. Admin can create and publish directly.

### User Stories (Layer 1b)

| Role | Story |
|------|-------|
| Member | Open my training log and see I've attended 47 classes this year with 120 mat hours (108 recorded + 12 estimated) and a current 3-week streak |
| Member | See my belt icon with 2 stripes on my Blue belt — the system inferred them from my 85/150 hours toward Purple eligibility |
| Member | Set a personal goal of 3 sessions per week and see my weekly progress on the dashboard |
| Member | Earn a "100 classes" milestone badge displayed on my training log |
| Member | Submit an estimate: "I trained 3×/week for 6 weeks while travelling at Checkmat SP" — it goes to Admin for review |
| Member | See a notice on my dashboard that the gym is closed next Monday for a public holiday |
| Member | See a message from Admin on my dashboard: "Hey — everything OK? We missed you at No-Gi" |
| Coach | Before class starts, see a class-specific notice: "Remind students — grading this Saturday" |
| Coach | Open the grading readiness list and see belt/stripe icons alongside flight time progress for each member |
| Coach | Propose a promotion for Sarah from Blue to Purple — Admin receives the proposal for approval |
| Coach | Add a private observation to a student's profile: "Tends to muscle techniques — focus on flow drilling" |
| Coach | A member hasn't been checking in — I select Jan–Mar, enter "3 hours/week", and the system adds 36 estimated hours to their flight time |
| Coach | Draft a school-wide notice: "Open Mat cancelled this Friday" — Admin reviews and publishes |
| Admin | See a Kids grading view showing term attendance %: "Mia — Grey belt — 26/30 sessions (87%) — eligible" |
| Admin | Toggle the Kids grading metric from sessions to hours for a specific child who trains at multiple gyms |
| Admin | Review a student's self-estimate submission, adjust hours from 18 to 12, approve with a note |
| Admin | Bulk-add estimated hours for a member who trained Jan–Jun without checking in — system warns about overlap with February recorded hours, I choose "add" to keep both |
| Admin | See the inactive member radar showing 8 members who haven't trained in 30+ days |
| Admin | Archive 5 members who haven't checked in for 3+ months, keeping the active list clean |
| Admin | Restore a previously archived member who just walked back in after 6 months away |
| Admin | Approve a coach's grading proposal — member's belt is officially updated to Purple stripe 0 |
| Admin | Override grading for a competitor: add 20 hours of flight time credit because they're sandbagging Blue belt at comps |
| Admin | Configure a new milestone: "200 mat hours" achievement |
| Admin | Send a direct message to a member who missed last week — message appears on their dashboard |

---

## Layer 2: The Researcher's Spine

**Goal:** Transition the gym from "showing moves" to structured technical study with per-program curriculum rotors, student-driven theme requests, and a curated clip library.

### Features

| Feature | Description |
|---------|-------------|
| **Program Rotors** | Each program (Gi Express, NoGi Express, Feet-to-Floor, Kids, Youth, etc.) can have its own independent **rotor** — a cycling curriculum of themes. Admin creates rotors; coaches control advancement (moving to the next theme in the cycle). |
| **Themes** | A theme is a technical block with a configurable duration of **D days** or **W weeks** (e.g., "Leg Lasso Series — 2 weeks", "Takedown Week — 5 days"). Themes can optionally be **hidden** for surprise "fun" sessions (e.g., dodgeball, sumo) that are only revealed on the day they run. |
| **Topics** | Each theme contains **M topics** (specific techniques or drills). Each topic has a **frequency F** where F=1 means normal importance and F=2 means "cover this twice as often." Topics drive the day-to-day class content within a theme. |
| **Theme Requests & Voting** | Members can request themes from an admin-configured menu of requestable items categorised as **Connections** (positional relationships, e.g., "K-Guard → SLX", "Closed Guard → Back") or **Actions** (transitions, entries, attacks, defenses, e.g., "Leg Lasso Entry", "Arm Drag to Back"). Members can vote on other members' requests. Each requestable item shows when it was **last covered**. |
| **Vote-Driven Scheduling** | When Admin chooses to run a member-voted theme, the voted theme is **brought forward** in the rotor. The theme that was automatically scheduled for that slot is bumped to the next position in the cycle. |
| **Student Rotor Preview** | Admin can toggle whether students can see the upcoming rotor schedule for each program. When preview is enabled, members see the sequence of upcoming themes. When disabled, only the current theme is visible. |
| **The Clipping Tool** | Paste YouTube link, set Start/End timestamps for a single loop. Clips are associated with a theme. |
| **Technical Library** | Grid of theme cards where clips are stored. Searchable by theme, program, and tag. |
| **Promote Clip** | Coach or Admin can promote a member-submitted clip to the main library so the whole gym benefits. |
| **Global Mute** | Persistent setting ensuring all videos start without audio. |

### Rotor System Detail

**Structure:**

```
Program (e.g., "Gi Express")
  └── Rotor
       ├── Theme 1: "Leg Lasso Series" (2 weeks, visible)
       │    ├── Topic A: "Leg Lasso Entry from DLR" (F=1)
       │    ├── Topic B: "Leg Lasso Sweep to Mount" (F=2)
       │    └── Topic C: "Leg Lasso → Omoplata" (F=1)
       ├── Theme 2: "Half Guard Sweeps" (2 weeks, visible)
       │    ├── Topic A: "Underhook Recovery" (F=1)
       │    └── Topic B: "Lucas Leite Sweep" (F=2)
       ├── Theme 3: "★ SURPRISE ★" (1 day, hidden)
       │    └── Topic A: "Sumo Games" (F=1)
       └── Theme 4: "Closed Guard Attacks" (2 weeks, visible)
            ├── Topic A: "Cross-Collar Choke" (F=1)
            ├── Topic B: "Armbar from Guard" (F=1)
            └── Topic C: "Triangle Setup" (F=1)
```

**Rotor Advancement:**
- Coaches control when the rotor advances to the next theme (manual "advance" action)
- When a voted theme is brought forward, the displaced theme moves to the next slot
- The rotor cycles — after the last theme, it wraps back to the first
- Hidden themes are only revealed to students on the day they start

**Topic Frequency:**
- Each topic within a theme has a frequency weight F (default: 1)
- F=1: normal — covered once per cycle through the theme's topics
- F=2: high priority — covered twice as often (appears twice in the rotation)
- This helps coaches emphasise foundational techniques within a theme

**Theme Duration:**
- Duration is configurable per theme as either **D days** or **W weeks**
- Short themes (1–3 days) work well for fun/surprise content or focused workshops
- Standard themes (1–4 weeks) work for deep technical blocks

**Hidden Themes:**
- Themes can be marked as **hidden** by Admin or Coach
- Hidden themes do not appear in student rotor previews
- They are revealed only when they become the active theme
- Use case: surprise "fun" themes like dodgeball, sumo, or special guest workshops

### Theme Requests & Voting Detail

**Requestable Item Menu:**

Admin configures a menu of items that members can request. Items are categorised as:

| Category | Description | Example |
|----------|-------------|---------|
| **Connection** | A positional relationship or guard/position | "K-Guard → SLX", "Closed Guard", "Half Guard — Deep Half" |
| **Action** | A transition, entry, attack, defense, or escape | "Leg Lasso Entry", "Arm Drag to Back Take", "Guillotine Defense" |

Each item in the menu tracks:
- **Last Covered**: date the item was last part of an active theme
- **Request Count**: how many unique members have requested it
- **Vote Count**: total votes from members

**Member Flow:**
1. Browse the request menu (filterable by category)
2. See when each item was last covered (e.g., "Leg Lasso — last covered 3 weeks ago")
3. Request an item (one request per member per item)
4. Vote on other members' requests (one vote per member per request)
5. See the current vote rankings

**Admin/Coach Triage:**
1. View requests sorted by vote count
2. Choose to **bring forward** a requested theme into the rotor
3. The system bumps the currently scheduled theme to the next slot
4. The request is marked as "scheduled" and the last-covered date updates when the theme runs

### User Stories (Layer 2)

| Role | Story |
|------|-------|
| Member | See the current theme for Gi Express is "Leg Lasso Series — Week 2 of 2" and prepare my mind before class |
| Member | See the rotor preview for NoGi Express showing the next 4 themes coming up |
| Member | Request "De La Riva Guard" from the request menu — it was last covered 6 weeks ago |
| Member | Vote on another member's request for "Arm Drag Series" — it now has 8 votes |
| Member | See a surprise theme appear on Monday: "★ Sumo Games ★" — it wasn't on the preview! |
| Member | Isolate a 10-second sequence from a 40-minute match so I can watch the grip-fight on repeat |
| Coach | Advance the Feet-to-Floor rotor from "Wrestling Shots" to "Clinch Takedowns" after 2 weeks |
| Coach | See that "De La Riva Guard" has 12 votes — bring it forward in the Gi Express rotor |
| Coach | Create a hidden "Dodgeball Day" theme for Kids that won't appear in their preview |
| Coach | Set Topic B within a theme to frequency 2 because it's the most important entry to drill |
| Coach | "Promote" a member's clip to the main library so the whole gym benefits |
| Admin | Create a new rotor for the "NoGi Long Games" program with 6 themes, each 2 weeks long |
| Admin | Configure the request menu: add "Heel Hook Defense" as an Action item |
| Admin | Toggle student preview off for the Kids rotor so the schedule is a surprise each week |
| Admin | Review a voted theme with 15 requests — bring it forward, bumping "Half Guard" to next slot |

---

## Layer 3: The Advanced Laboratory

**Goal:** Enable high-level pattern recognition through data and multi-video study.

### Features

| Feature | Description |
|---------|-------------|
| Technical Tagging | Every clip tagged with Action (e.g., Sweep) and Connection (e.g., K-Guard → SLX) |
| The 4-Up Mode | 2x2 grid allowing four clips to loop simultaneously with unified controls |
| Predictive Search | Search bar suggesting clips based on technical relationships (tags) |
| Research Journal | Text field below 4-Up mode for personal "Unified Observations" — private research notes about technique patterns |

### User Stories

| Role | Story |
|------|-------|
| Member | Watch four athletes perform the same K-Guard entry at once to see universal mechanics |
| Member | Search for "Connection: SLX to Back" and have the system populate a study grid |

---

## Layer 4: The War Room & Calendar

**Goal:** Provide a unified calendar showing club events, competitions, program rotor schedules, and personal training goals. Focus the gym's technical power toward specific performance goals.

### Features

| Feature | Description |
|---------|-------------|
| **Club Calendar** | A shared calendar showing club events (seminars, social events, gym closures) and competitions (local and international). Admin manages events; all authenticated users can view. Each event type is **toggleable** — users can show/hide categories to declutter their view. |
| **Program Rotor Calendar View** | Each program's rotor schedule (current and upcoming themes) is rendered as a layer on the club calendar. Users can toggle individual program views on/off (e.g., show "Gi Express" themes but hide "Kids"). Draws data from the Layer 2 rotor system. |
| **Personal Goal Calendar Layer** | Members can overlay their own training goals on the calendar. Goals have a **target metric** and a **time period** (e.g., "50 rear naked strangles during April", "20 hours of De La Riva guard in Q2", "Train 4× per week in March"). Progress is tracked against attendance and training log data where possible. |
| **Competition Calendar** | Upcoming local/international events with "Interested" and "Registered" toggles. Teammates can see who else is going. |
| **Admin Advice Repository** | Read-only section for Coach-curated strategy guides and rule-set breakdowns. |
| **Extra Session RSVP** | Minimalist booking system for high-intensity or "Comp-Only" mat times. |
| **Scouting Pipeline** | Members submit rival clips; Admin vets and publishes as official "Advice". |

### Club Calendar Detail

**Calendar Layers (all toggleable):**

| Layer | Content | Default | Managed by |
|-------|---------|---------|------------|
| **Club Events** | Seminars, social events, gym closures, grading days | On | Admin |
| **Competitions** | Local and international events with registration status | On | Admin |
| **Gi Express Rotor** | Current and upcoming themes for Gi Express | Off | System (from rotor) |
| **NoGi Express Rotor** | Current and upcoming themes for NoGi Express | Off | System (from rotor) |
| **Feet-to-Floor Rotor** | Current and upcoming themes for Feet-to-Floor | Off | System (from rotor) |
| **Kids Rotor** | Current and upcoming themes for Kids | Off | System (from rotor) |
| **Youth Rotor** | Current and upcoming themes for Youth | Off | System (from rotor) |
| *(other program rotors)* | Auto-generated for each program with a rotor | Off | System (from rotor) |
| **My Goals** | Personal training goals with progress | Off | Member (own) |

**Personal Goals:**

Members create goals with:
- **Description**: free text (e.g., "50 rear naked strangles")
- **Target**: a numeric target (e.g., 50)
- **Unit**: what's being counted (e.g., "submissions", "hours", "sessions")
- **Period**: start and end date (e.g., "1 Apr – 30 Apr")
- **Progress**: tracked manually by the member (increment/decrement) or automatically where the system can infer (e.g., hours from attendance, sessions from check-ins)

Goals appear as coloured bars on the calendar spanning their period, with a progress indicator.

### User Stories (Layer 4)

| Role | Story |
|------|-------|
| Member | Open the club calendar and see grading day on Saturday, a seminar next month, and Grappling Industries in 3 weeks |
| Member | Toggle on the "Gi Express Rotor" layer and see that "Leg Lasso" runs this week and "Closed Guard" starts next week |
| Member | Toggle on "NoGi Express" and "Feet-to-Floor" rotors simultaneously to see how the programs align |
| Member | Add a personal goal: "20 hours of De La Riva guard study in February" — it appears as a bar on my calendar |
| Member | Add a goal: "50 rear naked strangles during April" and manually log progress after each session |
| Member | See which teammates are going to Grappling Industries so we can coordinate study |
| Member | Read Coach's specific advice on ADCC rule-set so I don't give away points |
| Coach | View the calendar with all program rotors enabled to see the full curriculum across the week |
| Admin | Add a club event: "End of Year BBQ — Saturday 14 Dec" — it appears on everyone's calendar |
| Admin | Add a competition: "Grappling Industries Wellington — 15 Mar" with registration link |

---

## Layer 5: The Sustainable Facility

**Goal:** Automate business logic to protect the owner's time and the gym's liquidity.

### Features

| Feature | Description |
|---------|-------------|
| Xero/Bank Reconciliation | Auto-updates "Last Payment" and "Status" (Green/Amber/Red) by matching bank references |
| Program ROI Dashboard | Filter earnings by Adults vs. Kids to track $65,300/year overhead coverage |
| The Digital Wallet | Member-side view of reconciled receipts for tax purposes |

### User Stories

| Role | Story |
|------|-------|
| Admin | See a "Break-Even" progress bar showing when the month's rent is covered |
| Member | Download payment receipts for the last six months without emailing the owner |

### Technical ROI Formula

Track the health of the high-agency model using:

```
ROI = (α × Attendance + β × Study) / DaysSinceLastCheckIn
```

- **Attendance** (A): Physical presence (flight time)
- **Study** (S): Clips tagged / Research Journal entries
- **DaysSinceLastCheckIn** (D): Recency factor
- **α, β**: Custom weights for physical vs. technical engagement

---

## Concepts (Guidelines Mapping)

| Concept | Layer | Storage | Description |
|---------|-------|---------|-------------|
| `Account` | 1a | accounts table | User identity with role (admin/coach/member/trial/guest) and password hash |
| `Member` | 1a | members table | Profile data: name, email, program, class types, fee, frequency, status (active/inactive/archived) |
| `Waiver` | 1a | waivers table | Risk/responsibility acknowledgement, valid 1 year |
| `Injury` | 1a | injuries table | Red Flag body-part toggle, active 7 days |
| `Attendance` | 1a | attendance table | Check-in record linking member to a resolved class session (program + class type + date + time). Supports multi-session check-in and un-check-in (soft delete) |
| `Program` | 1a | programs table | Top-level grouping (Gi Express, NoGi Express, Feet-to-Floor, Kids, Youth, etc.). Determines term structure, billing rules, and rotor assignment |
| `ClassType` | 1a | class_types table | Class category within a program (Nuts & Bolts, Gi Express, NoGi Long Games, Open Mat, etc.) |
| `Schedule` | 1a | schedules table | Recurring weekly class timetable entry: day, time, class type, coach, program, duration |
| `Term` | 1a | terms table | NZ school term date ranges with manual confirmation flag. Default suggestions for 4 NZ terms per year |
| `Holiday` | 1a | holidays table | Date ranges that override the default schedule; auto-generates a holiday Notice |
| `Notice` | 1b | notices table | Unified notification: type (school_wide / class_specific / holiday), status (draft / published), content, target |
| `Message` | 1b | messages table | Direct in-app messages from Admin to individual members |
| `GradingRecord` | 1b | grading_records table | Belt promotion history: belt, stripe, date, proposed_by, approved_by, method (standard/override) |
| `GradingConfig` | 1b | grading_config table | Per-belt eligibility thresholds: flight time (adults) or attendance % / session count (kids), stripe count, grading mode toggle (hours/sessions). Configurable by Admin |
| `GradingProposal` | 1b | grading_proposals table | Coach-proposed promotion: member, target belt, notes, status (pending/approved/rejected) |
| `EstimatedHours` | 1b | estimated_hours table | Bulk-added estimated training hours for a date range. Fields: member_id, start_date, end_date, weekly_hours, total_hours, source (estimate/self_estimate), submitted_by, approved_by, status (pending/approved/rejected), note, overlap_mode (replace/add) |
| `TrainingGoal` | 1b | training_goals table | Member-set personal targets (e.g., sessions per week) |
| `Milestone` | 1b | milestones table | Admin-configured achievements (e.g., "100 classes", "1 year streak", "200 mat hours") |
| `CoachObservation` | 1b | coach_observations table | Private per-student notes from Coach or Admin |
| `BeltConfig` | 1b | belt_config table | Belt/stripe icon configuration per program. Fields: program, belt_name, belt_colour (hex or split pair), stripe_count, sort_order, age_range |
| `Rotor` | 2 | rotors table | Curriculum cycle for a program. Fields: program_id, name, current_theme_index, student_preview_enabled, created_by |
| `Theme` | 2 | themes table | Technical block within a rotor. Fields: rotor_id, name, description, duration_value, duration_unit (days/weeks), sort_order, hidden, start_date, created_by |
| `Topic` | 2 | topics table | Specific technique or drill within a theme. Fields: theme_id, name, description, frequency (default 1), sort_order |
| `ThemeRequestItem` | 2 | theme_request_items table | Admin-configured requestable item. Fields: name, category (connection/action), description, last_covered_date |
| `ThemeRequest` | 2 | theme_requests table | Member request for a theme item. Fields: item_id, member_id, status (open/scheduled/closed), created_at |
| `ThemeVote` | 2 | theme_votes table | Member vote on a theme request. Fields: request_id, member_id, created_at. One vote per member per request |
| `Clip` | 2, 3 | clips table | YouTube timestamp loop associated with a theme |
| `Tag` | 3 | tags table | Action/Connection metadata for clips |
| `ResearchEntry` | 3 | research_entries table | Personal research journal notes from 4-Up mode |
| `CalendarEvent` | 4 | calendar_events table | Club event or competition. Fields: title, description, event_type (club_event/competition), start_date, end_date, registration_url, created_by |
| `PersonalGoal` | 4 | personal_goals table | Member's calendar goal. Fields: member_id, description, target, unit (submissions/hours/sessions), start_date, end_date, progress, created_at |
| `Advice` | 4 | advice table | Coach-curated strategy guides |
| `ExtraSession` | 4 | extra_sessions table | Extra mat-time bookings (renamed from Session to avoid collision with class sessions) |
| `Payment` | 5 | payments table | Reconciled bank transactions |

---

## Implementation Priority

1. **Layer 1a** — Core operations: kiosk, session auto-select, multi-session check-in, un-check-in, schedule (with actual programs), waiver, red flags, member list, historical attendance
2. **Layer 1b** — Engagement: training log, grading (stripe inference, term-based toggle, belt/stripe icons, estimated hours, self-estimates), notices, messaging, coach observations, archive
3. **Layer 2** — Curriculum engine: per-program rotors, themes with topics and frequency, theme requests and voting, hidden themes, clip library
4. **Layer 4** — Club calendar (events, competitions, rotor views, personal goals)
5. **Layer 5** — Revenue protection (business sustainability)
6. **Layer 3** — Advanced research (power users)
