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
| 2 | **Check-In** | Fast kiosk check-in, select their class, done |
| 3 | **Training Log** | Personal training hours (flight time), classes attended, streaks, grading progress, and milestone achievements |
| 4 | **Club Notices** | Announcements from admin/coaches: schedule changes, events, gym news, holiday closures |

### Coach

| Priority | What they need | Why |
|:--------:|----------------|-----|
| 1 | **Class Timetable** | See their classes for today/this week, know which program and time slot |
| 2 | **Class Themes** | Current technical block for each class so they can prep curriculum |
| 3 | **Injury Flags** | Red flags on today's attendance — know who needs modifications before class starts |
| 4 | **Class Notices** | Pre-set announcements to make at the beginning or end of class (set by admin or self) |
| 5 | **Grading Readiness List** | Members approaching grading eligibility. Quick access to take notes on readiness and areas to work on |
| 6 | **Coach Observations** | Private per-student notes: technique feedback, grading observations, behavioural notes |

### Admin

| Priority | What they need | Why |
|:--------:|----------------|-----|
| 1 | **Content Configuration** | Full control over schedules, themes, notices, holidays — everything the gym runs on |
| 2 | **Notices & Coach Reminders** | Set school-wide notices (visible to all) and class-specific reminders (visible before/after class) |
| 3 | **Direct Messaging** | Message individual students — follow-ups, payment reminders, personal check-ins |
| 4 | **Inactive Member Radar** | See who hasn't checked in for a while — proactive retention before they ghost |
| 5 | **Coach Management** | Add new coaches, assign them to classes, manage their access |
| 6 | **Archive / Restore Members** | Archive members who haven't trained in a while (keeps data, removes from active lists). Restore when they return. |

---

## Feature Access by Role

| Feature | Admin | Coach | Member | Trial | Guest |
|---------|:-----:|:-----:|:------:|:-----:|:-----:|
| **Layer 1a: Core Operations** | | | | | |
| Kiosk Mode (launch/exit) | ✓ | ✓ | — | — | — |
| Kiosk Check-In (name search) | — | — | ✓ | ✓ | ✓ |
| Program Schedule Management | ✓ | — | — | — | — |
| Holiday Schedule Management | ✓ | — | — | — | — |
| View Schedule / Dashboard | ✓ | ✓ | ✓ | ✓ | ✓ |
| Digital Onboarding (Waiver) | ✓ | ✓ | ✓ | ✓ | ✓ |
| Red Flag (Injury Toggle) | — | — | ✓ | ✓ | — |
| View Attendance + Red Flags | ✓ | ✓ | — | — | — |
| Member List + Management | ✓ | View | — | — | — |
| Role Management | ✓ | — | — | — | — |
| **Layer 1b: Engagement** | | | | | |
| Training Log (own history) | — | — | ✓ | ✓ | — |
| Training Goals | — | — | ✓ | — | — |
| Milestones & Achievements | ✓ (configure) | View | View own | — | — |
| Notices (school-wide) | Publish | Draft | View | View | View |
| Notices (class-specific) | Publish | Draft | — | — | — |
| Notices (holiday — auto) | ✓ (configure) | — | View | View | View |
| Direct Messaging | ✓ | — | View own | — | — |
| Grading System | ✓ (approve/override) | Propose | View own | View own | — |
| Grading Readiness List | ✓ | ✓ | — | — | — |
| Coach Observations (private) | ✓ | ✓ | — | — | — |
| Inactive Member Radar | ✓ | ✓ | — | — | — |
| Archive / Restore Members | ✓ | — | — | — | — |
| Coach Management | ✓ | — | — | — | — |
| **Layer 2: Spine** | | | | | |
| Theme Carousel | ✓ | ✓ | ✓ | — | — |
| Clipping Tool | ✓ | ✓ | ✓ | — | — |
| Technical Library | ✓ | ✓ | ✓ | — | — |
| Promote Clip | ✓ | ✓ | — | — | — |
| **Layer 3: Laboratory** | | | | | |
| Technical Tagging | ✓ | ✓ | ✓ | — | — |
| 4-Up Mode | ✓ | ✓ | ✓ | — | — |
| Predictive Search | ✓ | ✓ | ✓ | — | — |
| Research Journal | ✓ | ✓ | ✓ | — | — |
| **Layer 4: War Room** | | | | | |
| Competition Calendar | ✓ | ✓ | ✓ | — | — |
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
| **1a. Core Operations** | Safety/Ops | Tablet (Kiosk + Check-in) | Database/Auth/Roles |
| **1b. Engagement** | Retention/Culture | Dashboard (Member + Coach) | Notifications/Grading |
| **2. Spine** | Technical Alignment | Mobile (Library) | YouTube API (v1) |
| **3. Laboratory** | Research/Logic | Desktop/Tablet (2x2) | YouTube API (v2) / Tagging |
| **4. War Room** | Performance | Mobile (Hub) | Calendar / CMS |
| **5. Optimization** | Sustainability | Admin (Dashboard) | Xero API |

---

## Layer 1a: Core Operations

**Goal:** Replace paper trails, ensure facility safety, and provide the daily operating backbone — kiosk, check-in, scheduling, waivers, and injury tracking.

### Features

| Feature | Description |
|---------|-------------|
| **Kiosk Mode** | A locked check-in screen for the front-desk tablet. Launched by a Coach or Admin from their own logged-in account. The kiosk is **tied to the account that launched it** — only that person's password can exit it. If a different Coach/Admin needs to exit, they must first authenticate with their own credentials. Prevents members from navigating away. |
| **Check-In by Name Search** | The **only** way to check in is by typing your name. The system performs a fuzzy search and presents a shortlist of matching Active members as the user types. The member selects their name from the suggestions to proceed to class selection. No member ID, email, or barcode is ever required. Only Active members appear in results (Inactive and Archived are hidden). This applies to both the kiosk and the standard check-in form. |
| **Class Selection at Check-In** | After selecting their name, the member chooses which of today's scheduled classes they are checking into (e.g., "6:00 AM Adults Fundamentals", "5:30 PM Kids"). Only classes running today are shown, resolved on-the-fly from Schedule + Terms + Holidays. |
| **Program Schedule Management** | Admin defines the recurring weekly class schedule for each program. Programs have two tiers: a top-level **Program** (Adults, Kids — determines term dates and billing) and **class types** within it (Fundamentals, No-Gi, Competition, etc.). Some programs (e.g., Kids) run aligned to NZ school terms, but term start/end dates must be manually confirmed by Admin each term — the system suggests NZ term dates but does not auto-activate. |
| **Holiday Schedule** | Admin can enter holiday dates that override the default schedule (e.g., Christmas break, public holidays, gym closures). Holidays auto-generate a notice on the dashboard so members know when the gym is closed or running a modified timetable. |
| **Digital Onboarding (Waiver)** | QR-driven sign-off acknowledging risks and responsibilities as a member or guest at Workshop. The waiver is **not** a liability transfer — it is a reminder of the inherent risks of martial arts training and the member's responsibilities (hygiene, conduct, injury reporting). Valid for 1 year. For Guests, completing the waiver flow creates a lightweight account (name + email). |
| **The "Red Flag"** | Member-managed toggle selecting an injured body part — appears as a warning icon next to their name on the Coach's attendance list. Active for 7 days. |
| **Member List** | All members split by Program and class type, with manual "Fee" and "Frequency" fields, injury indicators, waiver status, and current member status (active/inactive/archived). |

### Kiosk Mode Detail

The kiosk is the primary interface for day-to-day gym operations:

1. **Coach/Admin opens Kiosk Mode** — authenticates with their own password; tablet locks into kiosk view tied to that account
2. **Member walks up** — types their name into the search bar. The system performs a **fuzzy search** and presents a **shortlist of matching Active members** as the user types (autocomplete). No member ID, email, or barcode is required. Only Active members appear; Inactive and Archived are hidden.
3. **Member selects their name from the shortlist** — system resolves today's classes for their program on-the-fly (Schedule + Terms − Holidays) and shows available classes
4. **Member selects their class** — check-in is recorded, screen resets for the next person
5. **Trial member sees prompt** — after check-in, a message displays: *"Enjoying Workshop? Talk to your coach about signing up!"*
6. **Coach/Admin exits Kiosk Mode** — the person who launched the kiosk enters their password to exit. A different Coach/Admin must authenticate with their own credentials first.

For **Guests**: a "Guest Check-In" button on the kiosk launches the waiver flow (creating a lightweight account with name + email), then records attendance. If the guest is recognised from a previous visit, their existing record is used and they are prompted to convert to Trial or Member.

### Programs & Schedule Detail

**Two-Tier Program Model:**

- **Program** (top level): Adults, Kids — extensible by Admin. Determines term structure and billing rules.
- **Class Type** (within a program): Fundamentals, No-Gi, Competition, Open Mat, etc. — defined per Schedule entry.

**Weekly Timetable:** Admin sets recurring classes with: day, time, class type, coach, program. Example: *"Monday 6:00 AM — Adults Fundamentals — Coach Pat"*.

**NZ School Terms:** Kids programs can be linked to NZ school terms. The system provides suggested term dates but Admin must manually confirm start/end dates each term before the schedule activates.

**Holidays:** Admin creates holiday entries (date range, label, optional message). During a holiday, affected classes are excluded from the on-the-fly class resolution and a holiday notice is auto-generated on the dashboard (e.g., *"Gym closed 23 Dec – 7 Jan: Summer Break"*).

**Class Session Resolution:** Class sessions are **not** pre-generated. When the kiosk or dashboard loads "today's classes," the system resolves `Schedule + Terms − Holidays` on-the-fly to determine which classes are available. No rows are stored until a member checks in.

### User Stories (Layer 1a)

| Role | Story |
|------|-------|
| Guest | Scan a QR code, enter my name and email, sign the waiver acknowledging risks and responsibilities, and get on the mats in under 60 seconds |
| Guest | Return for a second visit and have the kiosk recognise me — prompted to talk to a coach about becoming a Trial or Member |
| Member | Walk up to the kiosk, type my name, tap my name, select "6 PM No-Gi", and I'm checked in — 5 seconds |
| Member | Flag that my knee is sore so the coach knows to give me space without me making a scene |
| Trial | Check in at the kiosk just like a member, and see a friendly prompt after: "Enjoying Workshop? Talk to your coach about signing up!" |
| Coach | See a red flag icon next to a student's name on today's attendance so I can acknowledge their injury |
| Coach | Launch kiosk mode from my account and know only my password (or another coach logging in) can exit it |
| Admin | Set up the Kids term schedule by confirming the NZ Term 1 dates, and classes auto-populate on the calendar |
| Admin | Enter a holiday closure for Easter weekend and know the dashboard will inform members via auto-generated notice |
| Admin | Convert a Trial to a full Member after they've spoken to the coach and are ready to commit |

---

## Layer 1b: Engagement

**Goal:** Retain members, support grading decisions, and keep the club informed — training logs, grading, notices, messaging, and coach tools.

### Features

| Feature | Description |
|---------|-------------|
| **Training Log** | A projection over the member's Attendance data showing: classes attended, total mat hours ("flight time"), streaks, date of last check-in, and grading progress. Members can set personal training goals (e.g., "3× per week"). Admin can configure milestones (e.g., "100 classes", "1 year streak") that display as achievements. |
| **Grading System** | Tracks each member's current belt and progress toward their next grade. **Adults:** grading eligibility is based on accumulated "flight time" (total mat hours). **Kids:** grading eligibility is based on percentage of classes attended in the current term. Admin can configure thresholds per belt level. Admin can override by adding time/classes or forcing a promotion. Coaches **propose** promotions (creating a pending record); Admin **approves** to make them official. Members can view their own training hours, grading history, and are informed that grading requires minimum flight times — *unless they are competing and sandbagging their division, at Admin's discretion.* |
| **Grading Readiness List** | Auto-generated list of members approaching grading eligibility thresholds. Coach can view readiness, add grading notes, and propose promotions from this view. |
| **Notices** | A unified notification system with three types: **school_wide** (general club announcements), **class_specific** (reminders for a coach to deliver at start/end of a particular class), and **holiday** (auto-generated from Holiday entries). Coaches can **draft** notices; Admin **publishes** them. Admin can also publish directly. All notices appear contextually: school-wide on dashboards, class-specific on the coach's class view, holidays on the dashboard and kiosk. |
| **Direct Messaging** | Admin can send messages to individual members — payment follow-ups, personal check-ins, welcome messages. In-app only in v1 (message appears on the member's dashboard with a notification badge). Email and SMS delivery planned for later. |
| **Coach Observations** | Private per-student notes written by Coach or Admin. Used for technique feedback, grading observations, behavioural notes. Not visible to the student. Accessible from the member's profile and from the Grading Readiness List. |
| **Inactive Member Radar** | List of members who haven't checked in for a configurable number of days. Visible to Admin and Coach. Feeds into the Admin's archive workflow — Admin can follow up, message, or archive directly from this view. |
| **Archive / Restore Members** | Admin can archive members who haven't trained in a while — changes their status to Archived, hiding them from all active views and kiosk search. All data is preserved. Archived members can be restored to Active at any time. |
| **Coach Management** | Admin can add new coaches, assign them to classes on the schedule, and manage their access. |

### Grading System Detail

**Belt Progression:**

- **Adults:** White → Blue → Purple → Brown → Black. Configurable stripe count per belt (default: 4 stripes before next belt).
- **Kids:** Separate progression track (e.g., White → Grey → Yellow → Orange → Green → Blue). Configurable by Admin.

**Eligibility Criteria:**

| Program | Metric | Example |
|---------|--------|---------|
| Adults | Accumulated flight time (mat hours) since last grading | "150 hours at Blue belt before Purple eligibility" |
| Kids | Percentage of classes attended in current term | "80% attendance this term to be eligible" |

**Admin Overrides:** Admin can adjust a member's grading eligibility by:
- Adding flight time or class credits manually
- Forcing an immediate promotion (bypassing thresholds)
- Adjusting thresholds per individual (e.g., competitor fast-tracking)

**Sandbagging Exception:** Members who are actively competing may be promoted ahead of schedule at Admin's discretion if they are sandbagging their division (competing below their skill level).

**Member Visibility:** Members see their own training hours, current belt, grading history, and a progress indicator toward the next grade. A note is displayed: *"Grading requires minimum flight time. Exceptions may apply for active competitors."*

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
| Member | Open my training log and see I've attended 47 classes this year with 120 mat hours and a current 3-week streak |
| Member | See my grading progress: "Blue belt — 85/150 hours toward Purple eligibility" |
| Member | Set a personal goal of 3 sessions per week and see my weekly progress on the dashboard |
| Member | Earn a "100 classes" milestone badge displayed on my training log |
| Member | See a notice on my dashboard that the gym is closed next Monday for a public holiday |
| Member | See a message from Admin on my dashboard: "Hey — everything OK? We missed you at No-Gi" |
| Coach | Before class starts, see a class-specific notice: "Remind students — grading this Saturday" |
| Coach | Open the grading readiness list and see who's approaching eligibility, with their flight time or attendance % |
| Coach | Propose a promotion for Sarah from Blue to Purple — Admin receives the proposal for approval |
| Coach | Add a private observation to a student's profile: "Tends to muscle techniques — focus on flow drilling" |
| Coach | Draft a school-wide notice: "Open Mat cancelled this Friday" — Admin reviews and publishes |
| Admin | See the inactive member radar showing 8 members who haven't trained in 30+ days |
| Admin | Archive 5 members who haven't checked in for 3+ months, keeping the active list clean |
| Admin | Restore a previously archived member who just walked back in after 6 months away |
| Admin | Approve a coach's grading proposal — member's belt is officially updated to Purple stripe 0 |
| Admin | Override grading for a competitor: add 20 hours of flight time credit because they're sandbagging Blue belt at comps |
| Admin | Configure a new milestone: "200 mat hours" achievement |
| Admin | Send a direct message to a member who missed last week — message appears on their dashboard |

---

## Layer 2: The Researcher's Spine

**Goal:** Transition the gym from "showing moves" to "technical study."

### Features

| Feature | Description |
|---------|-------------|
| Theme Carousel | Displays current 4-week technical block (e.g., Week 2 of 4: Leg Lasso) |
| The Clipping Tool | Paste YouTube link, set Start/End timestamps for a single loop |
| Technical Library | Grid of "Theme Cards" where clips are stored |
| Global Mute | Persistent setting ensuring all videos start without audio |

### User Stories

| Role | Story |
|------|-------|
| Member | See exactly what technical block we are in so I can prepare my mind before class |
| Member | Isolate a 10-second sequence from a 40-minute match so I can watch the grip-fight on repeat |
| Coach | "Promote" a member's clip to the main library so the whole gym benefits |

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

## Layer 4: The War Room

**Goal:** Focus the gym's technical power toward specific performance goals.

### Features

| Feature | Description |
|---------|-------------|
| Competition Calendar | Upcoming local/international events with "Interested" and "Registered" toggles |
| Admin Advice Repository | Read-only section for Coach-curated strategy guides and rule-set breakdowns |
| Extra Session RSVP | Minimalist booking system for high-intensity or "Comp-Only" mat times |
| Scouting Pipeline | Members submit rival clips; Admin vets and publishes as official "Advice" |

### User Stories

| Role | Story |
|------|-------|
| Member | See which teammates are going to Grappling Industries so we can coordinate study |
| Member | Read Coach's specific advice on ADCC rule-set so I don't give away points |

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
| `Attendance` | 1a | attendance table | Check-in record linking member to a resolved class session (program + class type + date + time) |
| `Program` | 1a | programs table | Top-level grouping (Adults, Kids). Determines term structure and billing rules |
| `ClassType` | 1a | class_types table | Class category within a program (Fundamentals, No-Gi, Competition, Open Mat, etc.) |
| `Schedule` | 1a | schedules table | Recurring weekly class timetable entry: day, time, class type, coach, program |
| `Term` | 1a | terms table | NZ school term date ranges with manual confirmation flag |
| `Holiday` | 1a | holidays table | Date ranges that override the default schedule; auto-generates a holiday Notice |
| `Notice` | 1b | notices table | Unified notification: type (school_wide / class_specific / holiday), status (draft / published), content, target |
| `Message` | 1b | messages table | Direct in-app messages from Admin to individual members |
| `GradingRecord` | 1b | grading_records table | Belt promotion history: belt, date, proposed_by, approved_by, method (standard/override) |
| `GradingConfig` | 1b | grading_config table | Per-belt eligibility thresholds: flight time (adults) or attendance % (kids), configurable by Admin |
| `GradingProposal` | 1b | grading_proposals table | Coach-proposed promotion: member, target belt, notes, status (pending/approved/rejected) |
| `TrainingGoal` | 1b | training_goals table | Member-set personal targets (e.g., sessions per week) |
| `Milestone` | 1b | milestones table | Admin-configured achievements (e.g., "100 classes", "1 year streak", "200 mat hours") |
| `CoachObservation` | 1b | coach_observations table | Private per-student notes from Coach or Admin |
| `Theme` | 2 | themes table | 4-week technical block |
| `Clip` | 2, 3 | clips table | YouTube timestamp loop |
| `Tag` | 3 | tags table | Action/Connection metadata for clips |
| `ResearchEntry` | 3 | research_entries table | Personal research journal notes from 4-Up mode |
| `Competition` | 4 | competitions table | Event calendar entries |
| `Advice` | 4 | advice table | Coach-curated strategy guides |
| `Session` | 4 | sessions table | Extra mat-time bookings |
| `Payment` | 5 | payments table | Reconciled bank transactions |

---

## Implementation Priority

1. **Layer 1a** — Core operations: kiosk, check-in, schedule, waiver, red flags, member list
2. **Layer 1b** — Engagement: training log, grading, notices, messaging, coach observations, archive
3. **Layer 2** — Core value proposition (technical study)
4. **Layer 5** — Revenue protection (business sustainability)
5. **Layer 3** — Advanced research (power users)
6. **Layer 4** — Competition focus (niche use case)
