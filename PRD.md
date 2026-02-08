# Workshop Jiu Jitsu — Product Requirements Document

Organised by **feature area**. Each feature includes a description, role access, detailed user stories, and acceptance criteria.

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

#### User Stories

**US-2.1.1: Launch kiosk mode**
As a Coach, I want to launch kiosk mode from my logged-in account so that the tablet is locked into the check-in screen for members.

- *Given* I am logged in as Coach or Admin
- *When* I tap "Launch Kiosk Mode"
- *Then* the tablet locks into the kiosk view, tied to my account
- *And* only my password (or another Coach/Admin authenticating) can exit it

**US-2.1.2: Exit kiosk mode**
As a Coach, I want to exit kiosk mode with my password so that I can return to the full app.

- *Given* the kiosk is running, launched by my account
- *When* I tap "Exit Kiosk" and enter my password
- *Then* the kiosk closes and I am returned to the normal app

**US-2.1.3: Different coach exits kiosk**
As a Coach, I want to exit a kiosk launched by another coach so that I can take over if needed.

- *Given* the kiosk was launched by Coach A
- *When* Coach B taps "Exit Kiosk" and authenticates with their own credentials
- *Then* the kiosk closes and Coach B is logged in

### 2.2 Check-In by Name Search

The **only** way to check in is by typing your name. Fuzzy search presents a shortlist of matching Active members as the user types. No member ID, email, or barcode is ever required. Inactive and Archived members are hidden from results.

**Access:** Admin — | Coach — | Member ✓ | Trial ✓ | Guest ✓ (via waiver flow)

#### User Stories

**US-2.2.1: Member check-in**
As a Member, I want to type my name at the kiosk and see myself in the suggestions so that I can check in quickly.

- *Given* the kiosk is active and I am an Active member
- *When* I type the first few characters of my name
- *Then* a shortlist of matching Active members appears as I type
- *And* I can tap my name to proceed to session selection

**US-2.2.2: Inactive members hidden**
As a Coach, I want Inactive and Archived members hidden from kiosk search so that only current members can check in.

- *Given* a member has been marked Inactive or Archived
- *When* anyone searches for that name at the kiosk
- *Then* the member does not appear in search results

**US-2.2.3: Guest check-in**
As a Guest, I want to check in at the kiosk even though I don't have an account so that I can get on the mats.

- *Given* I tap "Guest Check-In" on the kiosk
- *When* I enter my name and email and complete the waiver
- *Then* a lightweight Guest account is created and my attendance is recorded
- *And* the whole process takes under 60 seconds

**US-2.2.4: Returning guest recognised**
As a returning Guest, I want the kiosk to recognise me so that I don't have to fill out the waiver again.

- *Given* I visited previously and my Guest account exists
- *When* I type my name at the kiosk
- *Then* my existing record is found and I am prompted to convert to Trial or Member

### 2.3 Session Auto-Select

After selecting their name, the system auto-selects the session closest to the current time. The member can confirm, change, or select multiple sessions.

**Access:** All check-in users (Member, Trial, Guest)

#### User Stories

**US-2.3.1: Auto-select closest session**
As a Member, I want the system to auto-select the closest session when I check in so that check-in is as fast as possible.

- *Given* it is 6:05 PM and today's sessions include "6:00 PM Nuts & Bolts" and "7:15 PM Feet-to-Floor"
- *When* I select my name at the kiosk
- *Then* "6:00 PM Nuts & Bolts" is pre-selected (highlighted)
- *And* I can confirm with a single tap

**US-2.3.2: Change auto-selected session**
As a Member, I want to change the auto-selected session so that I can check into the correct class if I arrived early.

- *Given* the system auto-selected "6:00 PM Nuts & Bolts"
- *When* I tap "7:15 PM Feet-to-Floor" instead
- *Then* my selection changes to the 7:15 session
- *And* I can confirm check-in for that session

### 2.4 Multi-Session Check-In

Members can check into more than one session in a single visit. Each session is recorded as a separate attendance entry.

**Access:** Member ✓ | Trial ✓ | Guest —

#### User Stories

**US-2.4.1: Check into multiple sessions**
As a Member staying for back-to-back classes, I want to select multiple sessions at check-in so that I don't have to check in twice.

- *Given* today has "6:00 PM Nuts & Bolts" and "7:15 PM Feet-to-Floor"
- *When* I select both sessions and confirm
- *Then* two separate attendance records are created (one per session)
- *And* both sessions show me as checked in

### 2.5 Un-Check-In

If a member checked into the wrong session by mistake, they can undo their check-in from the kiosk. Only available for the current day.

**Access:** Member ✓ | Trial ✓ | Guest —

#### User Stories

**US-2.5.1: Undo accidental check-in**
As a Member, I want to un-check-in from a session I selected by mistake so that my attendance record is accurate.

- *Given* I just checked into "6:00 PM Nuts & Bolts" by mistake
- *When* I search my name again and tap "Un-Check-In" next to that session
- *Then* my attendance record for that session is removed (soft delete)
- *And* I can select the correct session instead

**US-2.5.2: Un-check-in limited to today**
As a Member, I should only be able to un-check-in from today's sessions so that historical records remain intact.

- *Given* I checked into a session yesterday
- *When* I try to un-check-in today
- *Then* yesterday's sessions are not available for un-check-in

---

## 3. Attendance & Training Log

### 3.1 Attendance Records

Each check-in creates an attendance record linking a member to a resolved session (class + date + time). Mat hours are calculated from the session's configured duration. **All session types count equally by default** (1 hour = 1 mat hour). Admin can optionally configure a weighting multiplier per class (e.g., Open Mat = 0.5× because it's unstructured, Competition Prep = 1.5×).

**Access:** Admin ✓ (view all) | Coach ✓ (view all) | Member ✓ (own) | Trial ✓ (own) | Guest —

#### User Stories

**US-3.1.1: Attendance creates mat hours**
As a Member, I want my check-in to automatically add the session's duration to my mat hours so that my training is tracked.

- *Given* I check into "6:00 PM Nuts & Bolts" (60 min)
- *When* the attendance record is created
- *Then* 1 hour is added to my recorded mat hours

**US-3.1.2: Coach views today's attendance**
As a Coach, I want to see who has checked in for today's sessions so that I know who is on the mat.

- *Given* it is Tuesday and 5 members have checked into "6:30 PM Gi Express"
- *When* I view the attendance screen
- *Then* I see a list of 5 members with their names, belt icons, and any red flags

### 3.2 Historical Attendance View

Admin and Coach can navigate to view attendance for other days. A prominent "Back to Today" button allows quick return.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-3.2.1: Browse past attendance**
As a Coach, I want to view attendance for a past date so that I can check who came to a specific class.

- *Given* I am on the attendance screen showing today
- *When* I use the date picker or prev/next arrows to navigate to last Wednesday
- *Then* I see Wednesday's sessions with their check-in lists and injury flags
- *And* the data is read-only (I cannot modify past attendance)

**US-3.2.2: Back to Today**
As a Coach, I want a prominent "Back to Today" button so that I can quickly return to the live view after browsing history.

- *Given* I am viewing last Wednesday's attendance
- *When* I tap "Back to Today"
- *Then* I am immediately returned to today's live attendance view

### 3.3 Training Log

A member-facing projection of their attendance data: classes attended, mat hours displayed as "flight time" (split into recorded and estimated), streaks, belt/stripe icon, and belt progression progress.

**Access:** Admin — | Coach — | Member ✓ | Trial ✓ | Guest —

#### User Stories

**US-3.3.1: View training log**
As a Member, I want to see my training log so that I know how much I've trained and how close I am to my next grade.

- *Given* I have attended 47 classes this year
- *When* I open my training log
- *Then* I see: 47 classes, 120 mat hours shown as "flight time" (108 recorded + 12 estimated), current 3-week streak, Blue belt with 2 inferred stripes, and a progress bar toward Purple eligibility

**US-3.3.2: Milestone achievement**
As a Member, I want to earn milestone badges so that I feel rewarded for consistent training.

- *Given* Admin has configured a "100 classes" milestone
- *When* I check into my 100th class
- *Then* a "100 Classes" badge appears on my training log
- *And* I see a congratulatory notification on my dashboard

**US-3.3.3: Configure milestones**
As an Admin, I want to configure milestone thresholds so that the club can celebrate member achievements.

- *Given* I navigate to milestone configuration
- *When* I create a new milestone: "200 mat hours"
- *Then* any member who reaches 200 mat hours will receive the badge

### 3.4 Estimated Training Hours

For members with incomplete records, Admin or Coach can bulk-add estimated mat hours by selecting a date range and entering estimated weekly hours. Overlaps with existing recorded training are handled explicitly.

**Access:** Admin ✓ | Coach ✓ | Member — (see §3.5 for self-estimates)

#### User Stories

**US-3.4.1: Bulk-add estimated hours**
As a Coach, I want to add estimated training hours for a member who hasn't been checking in so that their mat hours reflect reality.

- *Given* a member trained Jan–Mar without checking in
- *When* I select the member, set the date range to Jan 1 – Mar 31, and enter "3 hours/week"
- *Then* the system calculates 13 weeks × 3 hours = 39 estimated hours
- *And* adds them to the member's mat hours, tagged as `source: estimate`

**US-3.4.2: Overlap warning — replace**
As an Admin, I want to be warned about overlapping training data so that I can choose how to handle it.

- *Given* I am adding estimated hours for Feb–Apr, but the member has recorded check-ins during February
- *When* I submit the estimate
- *Then* the system warns me about the overlap
- *And* I can choose "Replace" to remove February's recorded hours and substitute the estimate

**US-3.4.3: Overlap warning — add**
As an Admin, I want to add estimated hours on top of existing records when the member was training extra.

- *Given* the same overlap scenario
- *When* I choose "Add" instead of "Replace"
- *Then* the estimated hours are added in addition to the recorded hours for February
- *And* both recorded and estimated hours appear separately in the member's training log

### 3.5 Member Self-Estimates

Members can submit their own estimated training periods. Self-estimates require a note and are flagged for Admin review.

**Access:** Admin ✓ (review) | Coach — | Member ✓ (submit) | Trial — | Guest —

#### User Stories

**US-3.5.1: Submit self-estimate**
As a Member who trained elsewhere while travelling, I want to submit an estimated training period so that my mat hours stay accurate.

- *Given* I trained 3×/week for 6 weeks at a partner gym in São Paulo
- *When* I submit an estimate: date range, "3 hours/week", note: "Trained at Checkmat SP while travelling"
- *Then* the estimate appears in Admin's review queue with status "pending"
- *And* it is **not** added to my mat hours until approved

**US-3.5.2: Admin reviews self-estimate**
As an Admin, I want to review member-submitted estimates so that I can verify their accuracy.

- *Given* a member submitted an estimate of 18 hours
- *When* I open the review queue and view the submission
- *Then* I see the date range, weekly hours, total, and the member's note
- *And* I can approve as-is, adjust the hours (e.g., reduce to 12), or reject with a reason

**US-3.5.3: Approved estimate tagged**
As an Admin, I want approved self-estimates to be clearly marked so that they are distinguishable from recorded hours.

- *Given* I approve a self-estimate
- *When* the hours are added to the member's mat hours
- *Then* they are tagged as `source: self_estimate` and appear separately in the training log

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

#### User Stories

**US-4.1.1: View grading progress (adult)**
As a Member, I want to see my grading progress so that I know how close I am to my next belt.

- *Given* I am a Blue belt with 85 hours since my last promotion
- *When* I view my training log
- *Then* I see "Blue belt — 85/150 hours toward Purple eligibility" with a progress bar
- *And* my belt icon shows 2 inferred stripes

**US-4.1.2: View grading progress (kids)**
As a parent viewing my child's profile, I want to see their grading progress in term-based format.

- *Given* my child is a Grey belt in the Kids program, Sessions mode
- *When* I view their training log
- *Then* I see "Grey belt — 26/30 sessions this term (87%) — eligible for promotion"

**US-4.1.3: Member sees grading note**
As a Member, I want to understand the grading criteria so that I know what's expected.

- *Given* I view my grading progress
- *When* I look at the progress section
- *Then* a note is displayed: *"Belt progression requires minimum mat hours. Exceptions may apply for active competitors at Admin's discretion."*

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

#### User Stories

**US-4.2.1: Stripes inferred automatically**
As a Member, I want my stripes to be calculated automatically from my mat hours so that I always see my current progress.

- *Given* I am a Blue belt with a 150h threshold and I have 85 mat hours
- *When* I view my profile or training log
- *Then* my belt icon shows 2 stripes (floor(85 / 37.5) = 2)
- *And* no manual stripe award was needed

**US-4.2.2: Stripe updates as hours accumulate**
As a Member, I want my stripe count to update automatically when I train more so that my progress is always current.

- *Given* I have 74 hours (Stripe 1) and I check into a 1-hour session
- *When* my mat hours reach 75
- *Then* my stripe count updates to Stripe 2 across all views (training log, profile, attendance)

### 4.3 Term-Based Grading (Kids/Youth)

For term-based programs, Admin can toggle the grading metric between sessions attended and hours.

| Mode | Metric | Example |
|------|--------|---------|
| **Sessions** (default for kids) | Sessions attended in the current term | "24 of 30 sessions this term" |
| **Hours** | Accumulated mat hours (same as adults) | "45 hours since last promotion" |

**Sessions mode**: eligibility is based on **% of term attendance** — the system counts total available sessions in the current NZ school term, divides by attendance, and compares to admin-configured thresholds. Attendance resets each term. Admin has ultimate discretion over all kids/youth promotions.

**Access:** Admin ✓ (configure) | Coach ✓ (view) | Member ✓ (view own) | Trial — | Guest —

#### User Stories

**US-4.3.1: Kids grading by term attendance**
As an Admin, I want to see kids' grading eligibility by term attendance percentage so that I can identify who is ready for promotion.

- *Given* the Kids program is in Sessions mode with an 80% threshold
- *When* I view the kids grading readiness list during Term 2
- *Then* I see each child's attendance count, total available sessions, and percentage
- *And* children at or above 80% are marked as eligible

**US-4.3.2: Toggle grading metric**
As an Admin, I want to toggle a specific child's grading metric from sessions to hours so that I can handle edge cases.

- *Given* a child trains at Workshop and another gym
- *When* I toggle their grading metric to Hours mode
- *Then* their eligibility is based on accumulated mat hours instead of term attendance %
- *And* estimated hours from the other gym can be added to their record

**US-4.3.3: Term attendance counts reset**
As an Admin, I want term attendance counts to reset each term so that eligibility reflects current engagement.

- *Given* a child had 95% attendance in Term 1
- *When* Term 2 begins
- *Then* their term attendance starts fresh at 0/0
- *And* Term 1 results are preserved in their grading history

### 4.4 Belt & Stripe Icons

Visual CSS/SVG icons for all belt levels, displayed wherever a member's rank is shown.

- **Belt colour**: solid (adults) or split (kids dual-colour belts like Grey/White)
- **Stripe count**: 0–4 small markers on the belt tab
- **Black belt degrees**: per IBJJF

**Displayed on:** training logs, grading readiness lists, member profiles, dashboards, attendance lists.

**Access:** Admin ✓ (configure) | Coach ✓ (view) | Member ✓ (view own) | Trial ✓ (view own) | Guest —

#### User Stories

**US-4.4.1: Belt icon on training log**
As a Member, I want to see my belt icon with stripes on my training log so that I have a visual representation of my rank.

- *Given* I am a Blue belt with 2 inferred stripes
- *When* I open my training log
- *Then* I see a Blue belt icon with 2 stripe markers

**US-4.4.2: Belt icons on attendance**
As a Coach, I want to see belt icons next to members on the attendance list so that I know everyone's level at a glance.

- *Given* 8 members have checked in for today's Gi Express
- *When* I view the attendance list
- *Then* each member's name is accompanied by their belt/stripe icon

**US-4.4.3: Kids dual-colour belt icon**
As a parent, I want to see my child's dual-colour belt icon so that I can tell which belt they're on.

- *Given* my child is a Yellow/Black belt with 3 stripes
- *When* I view their profile
- *Then* the belt icon shows a split Yellow/Black colour with 3 stripe markers

### 4.5 Grading Readiness List

Auto-generated list of members approaching promotion eligibility. Shows belt icon, progress metric, and coach notes.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-4.5.1: View readiness list**
As a Coach, I want to see which members are approaching promotion eligibility so that I can prepare for the next grading day.

- *Given* several members are within 80% of their mat hours threshold
- *When* I open the grading readiness list
- *Then* I see each member's name, belt/stripe icon, mat hours progress (or term attendance %), and any grading notes
- *And* the list is sorted by proximity to eligibility

**US-4.5.2: Add grading notes**
As a Coach, I want to add notes to a member's grading readiness entry so that I can track areas for improvement.

- *Given* I am viewing a member's readiness entry
- *When* I add a note: "Needs to work on guard retention before promotion"
- *Then* the note is saved and visible to Admin and other coaches

**US-4.5.3: Propose promotion from readiness list**
As a Coach, I want to propose a promotion directly from the readiness list so that I don't have to navigate elsewhere.

- *Given* Sarah has 155 mat hours at Blue belt (threshold: 150h)
- *When* I click "Propose Promotion" and confirm "Blue → Purple"
- *Then* a promotion proposal is created with status "pending"
- *And* Admin receives a notification to review the proposal

### 4.6 Grading Proposals & Promotions

Coaches propose promotions; Admin approves to make them official. The workflow prevents unilateral promotions. The promotion ceremony (belt presentation) is handled outside the system — the system only tracks the record.

**Access:** Admin ✓ (approve/reject) | Coach ✓ (propose) | Member — | Trial — | Guest —

#### User Stories

**US-4.6.1: Coach proposes promotion**
As a Coach, I want to propose a promotion for a member so that the process is formalised and requires Admin sign-off.

- *Given* I believe a member is ready for promotion
- *When* I create a grading proposal with the target belt and optional notes
- *Then* the proposal is saved with status "pending"
- *And* Admin can see it in their grading review queue

**US-4.6.2: Admin approves promotion**
As an Admin, I want to approve a coach's promotion proposal so that the member's belt is officially updated.

- *Given* a pending grading proposal for Sarah: Blue → Purple
- *When* I review the proposal and click "Approve"
- *Then* Sarah's belt is updated to Purple stripe 0
- *And* the promotion is recorded in her grading history with date, proposed_by, and approved_by

**US-4.6.3: Admin rejects proposal**
As an Admin, I want to reject a promotion proposal with a reason so that the coach understands why.

- *Given* a pending proposal I disagree with
- *When* I click "Reject" and enter: "Needs 3 more months of consistent training"
- *Then* the proposal is marked as rejected
- *And* the coach can see the rejection reason

### 4.7 Admin Overrides

Admin can bypass normal grading thresholds for special circumstances.

**Access:** Admin ✓ | Coach — | Member — | Trial — | Guest —

#### User Stories

**US-4.7.1: Add mat hours credit**
As an Admin, I want to add mat hours credit to a competitor's record so that they aren't sandbagging.

- *Given* a member is competing at Blue belt but dominating their division
- *When* I add 20 hours of mat hours credit to their record
- *Then* their mat hours increase, pushing them closer to Purple eligibility
- *And* the credit is logged with my Admin ID and a reason

**US-4.7.2: Force immediate promotion**
As an Admin, I want to force a promotion bypassing thresholds so that I can handle exceptional cases.

- *Given* a member is clearly above their belt level but hasn't met the hour threshold
- *When* I force a promotion from Blue to Purple
- *Then* their belt is updated immediately
- *And* the grading record shows method: "override" with my notes

**US-4.7.3: Adjust individual thresholds**
As an Admin, I want to adjust belt progression thresholds for a specific member so that competitors can be fast-tracked.

- *Given* a member is actively competing and training at a higher level
- *When* I reduce their mat hours threshold from 150h to 120h for Blue → Purple
- *Then* their belt progression recalculates using the new threshold
- *And* the override is logged

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

#### User Stories

**US-5.1.1: Create rotor for a class**
As an Admin, I want to create a rotor for a class so that I can define its curriculum structure.

- *Given* the "Gi Express" class has no rotor
- *When* I create a new rotor with 4 theme categories: Standing, Guard, Pinning, In-Between Game
- *Then* the rotor is saved as v1 and attached to "Gi Express"
- *And* I can add topics to each theme category

**US-5.1.2: View this week's curriculum**
As a Member, I want to see what topics are scheduled this week so that I can prepare before class.

- *Given* the Gi Express rotor has 4 themes, each with a currently scheduled topic
- *When* I view the Gi Express class page
- *Then* I see: Standing: "Single Leg Takedown", Guard: "Closed Guard Attacks", Pinning: "Side Control Submissions", In-Between: "Turtle Escapes"

**US-5.1.3: Topic queues cycle**
As a Coach, I want topic queues to wrap around so that the curriculum repeats.

- *Given* the "Standing" theme has 6 topics and I am on the last one
- *When* the topic auto-advances
- *Then* the queue cycles back to the first topic

### 5.2 Themes (Concurrent Categories)

Themes are positional or tactical categories that run **concurrently** within a rotor. Every week, all themes are active — each with its own currently scheduled topic.

Example categories: **Standing** (takedowns, clinch work), **Guard** (bottom game, sweeps), **Pinning** (top control, submissions from top), **In-Between Game** (transitions, scrambles, turtle).

Admin defines the theme categories per rotor. Themes are extensible — Admin can add, rename, or remove categories.

#### User Stories

**US-5.2.1: Add a theme category**
As an Admin, I want to add a new theme category to a rotor so that the curriculum covers a new area.

- *Given* the Gi Express rotor has 4 themes
- *When* I add a 5th theme: "Submission Defense"
- *Then* the rotor now has 5 concurrent themes and I can populate its topic queue

**US-5.2.2: Hidden / surprise theme**
As a Coach, I want to mark a theme category as hidden so that members don't see it in previews.

- *Given* I add a special "Fun Day" theme to the Kids rotor
- *When* I mark it as hidden
- *Then* it does not appear in member previews and is only revealed when active

### 5.3 Topics & Topic Queues

Each theme has an ordered queue of topics. One topic is scheduled per theme per week (or configurable period). Topics auto-advance when their duration expires.

Each topic has:
- **Name**: the technique or drill (e.g., "Single Leg Takedown")
- **Duration**: how long it stays scheduled (default: 1 week, configurable per topic)
- **Description**: optional notes for the coach

#### User Stories

**US-5.3.1: Add topics to a theme**
As a Coach, I want to populate a theme's topic queue so that the weekly curriculum is defined.

- *Given* I am editing the "Standing" theme in the Gi Express rotor
- *When* I add 6 topics: "Single Leg", "Double Leg", "Arm Drag to Back", "Collar Drag", "Snap Down", "Clinch Takedowns"
- *Then* the topic queue is populated and the first topic is scheduled for this week

**US-5.3.2: Reorder topic queue**
As a Coach, I want to reorder the topic queue so that I can control the curriculum sequence.

- *Given* the Standing theme has 6 topics
- *When* I drag "Arm Drag to Back" to position 2
- *Then* it will be the second topic scheduled in the rotation

**US-5.3.3: Member views scheduled topics**
As a Member, I want to see the topics scheduled for each theme this week so that I know what class will cover.

- *Given* Gi Express has 4 themes, each with a currently scheduled topic
- *When* I view the class page (and preview is enabled)
- *Then* I see all 4 themes with their current topic

**US-5.3.4: Custom topic duration**
As a Coach, I want to set a topic's duration to 2 weeks so that complex techniques get more time.

- *Given* "Leg Lasso Series" is a complex topic
- *When* I set its duration to 2 weeks
- *Then* it stays as the scheduled topic for 2 weeks before auto-advancing

### 5.4 Rotor Advancement

Topics **auto-advance by default** when their configured duration expires. Coaches can override: extend a topic, skip to the next, or manually advance early.

**Access:** Admin ✓ (override all) | Coach ✓ (for classes they own) | Member — | Trial — | Guest —

#### User Stories

**US-5.4.1: Auto-advance**
As a Coach, I want topics to auto-advance when their duration expires so that the curriculum progresses without manual intervention.

- *Given* "Single Leg Takedown" in the Standing theme has a 1-week duration
- *When* the week ends
- *Then* the next topic in the queue ("Double Leg") automatically becomes the scheduled topic

**US-5.4.2: Extend a topic**
As a Coach, I want to extend a topic beyond its configured duration so that the class can spend more time on it.

- *Given* "Single Leg" was configured for 1 week but the class needs more time
- *When* I click "Extend" and add 1 more week
- *Then* the topic stays scheduled for a total of 2 weeks before auto-advancing

**US-5.4.3: Skip a topic**
As a Coach, I want to skip the next topic in the queue so that I can jump to something more relevant.

- *Given* "Double Leg" is next in the Standing queue but I want to skip to "Arm Drag"
- *When* I skip "Double Leg"
- *Then* "Arm Drag to Back" becomes the next scheduled topic
- *And* "Double Leg" moves to its normal position in the next rotation cycle

**US-5.4.4: Coach can only advance own classes**
As a Coach, I should only be able to advance rotors for classes I'm assigned to.

- *Given* I am assigned to Monday 6 PM Nuts & Bolts and Wednesday 6 PM Nuts & Bolts
- *When* I try to advance the rotor for Thursday NoGi Express (not my class)
- *Then* the system denies the action
- *And* only Admin can override

### 5.5 Rotor Versioning

Rotors are versioned to support drafting improvements without disrupting the live curriculum.

**Access:** Admin ✓ | Coach ✓ (for own classes) | Member — | Trial — | Guest —

#### User Stories

**US-5.5.1: Draft a new rotor version**
As a Coach, I want to create a draft version of a rotor so that I can plan curriculum changes without affecting the live schedule.

- *Given* the Gi Express rotor is on v2 (active)
- *When* I create a new draft (v3) and add a "Submission Defense" theme with 4 topics
- *Then* v2 remains active and members see no change
- *And* v3 is saved as a draft that I can continue editing

**US-5.5.2: Activate a draft rotor**
As an Admin, I want to activate a draft rotor version so that the new curriculum goes live.

- *Given* v3 of the Gi Express rotor is ready
- *When* I activate v3
- *Then* v3 becomes the active rotor and members see the updated themes/topics
- *And* v2 is preserved in version history

**US-5.5.3: View rotor history**
As an Admin, I want to see previous rotor versions so that I can review curriculum evolution.

- *Given* the Gi Express rotor has versions v1, v2, v3
- *When* I view rotor history
- *Then* I see all three versions with their themes, topics, and activation dates

### 5.6 Member Rotor Preview

Admin can toggle whether members can see the upcoming topic schedule for each class.

- **Preview on**: members see the full topic queues across all themes
- **Preview off**: members only see the currently scheduled topics

Hidden themes are never shown in previews regardless of the toggle.

**Access:** Admin ✓ (toggle) | Coach ✓ (view) | Member ✓ (view if enabled) | Trial — | Guest —

#### User Stories

**US-5.6.1: Toggle preview off**
As an Admin, I want to disable rotor preview for Kids classes so that the schedule is always a surprise.

- *Given* the Workshop Kids rotor preview is enabled
- *When* I toggle it off
- *Then* kids can only see the current topics, not upcoming ones

**US-5.6.2: View upcoming topics**
As a Member, I want to see what topics are coming up so that I can plan my training.

- *Given* the Gi Express rotor has preview enabled
- *When* I view the Gi Express class page
- *Then* I see each theme's current topic and the next 3–4 upcoming topics in each queue

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

#### User Stories

**US-6.1.1: Vote for a topic**
As a Member, I want to vote for an upcoming topic so that the class covers what I want to work on.

- *Given* the "Guard" theme's topic queue shows: [1. Closed Guard Attacks ←scheduled] → [2. DLR Sweeps] → [3. Leg Lasso Series] → [4. Half Guard]
- *When* I vote for "Leg Lasso Series" (currently position 3)
- *Then* my vote is recorded and the topic's vote count increases by 1
- *And* I cannot vote for "Leg Lasso Series" again until the next rotation cycle

**US-6.1.2: View vote rankings**
As a Member, I want to see which topics have the most votes so that I know what's popular.

- *Given* several members have voted on topics in the "Guard" theme
- *When* I view the topic queue
- *Then* I see each topic's vote count alongside its queue position

**US-6.1.3: See last covered date**
As a Member, I want to see when a topic was last covered so that I can vote for things we haven't done recently.

- *Given* "DLR Sweeps" was last covered 6 weeks ago and "Leg Lasso" 2 weeks ago
- *When* I view the topic queue
- *Then* I see the last-covered dates and can make an informed vote

### 6.2 Vote-Driven Topic Bump

When a topic accumulates enough votes (or Coach/Admin decides to honour the vote), it is **inserted before** the currently scheduled topic. The scheduled topic remains in its queue position and runs on the next rotation.

**Access:** Admin ✓ (override) | Coach ✓ (for own classes) | Member — | Trial — | Guest —

#### User Stories

**US-6.2.1: Bump the scheduled topic**
As a Coach, I want to bring forward a voted topic so that the class covers what members requested.

- *Given* "Leg Lasso Series" has 12 votes in the Guard theme, and "DLR Sweeps" is the next scheduled topic
- *When* I choose to bring forward "Leg Lasso Series"
- *Then* "Leg Lasso Series" is inserted as the current scheduled topic
- *And* "DLR Sweeps" stays in its queue position and will run on the next rotation

**US-6.2.2: Bumped topic not lost**
As a Coach, I want the bumped topic to remain in the queue so that it still gets covered.

- *Given* "DLR Sweeps" was bumped by "Leg Lasso Series"
- *When* the rotor completes the current cycle and returns to this position
- *Then* "DLR Sweeps" runs as originally planned

**US-6.2.3: Vote counts reset after bump**
As a Member, I want vote counts to reset after a topic is brought forward so that voting starts fresh.

- *Given* "Leg Lasso Series" was brought forward with 12 votes
- *When* it finishes its scheduled run
- *Then* all vote counts for that theme's topics reset to 0
- *And* members can vote again for the next cycle

**US-6.2.4: Admin overrides vote**
As an Admin, I want to override the vote and bring forward any topic regardless of vote count.

- *Given* "Rubber Guard" has only 2 votes but I want to schedule it
- *When* I override and bring it forward
- *Then* it becomes the scheduled topic, same bump mechanics apply

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

### 8.2 Direct Messaging

Admin can send messages to individual members. In-app only in v1 (message appears on the member's dashboard with a notification badge). Email and SMS delivery planned for later.

**Access:** Admin ✓ (send) | Coach — | Member ✓ (view own) | Trial — | Guest —

#### User Stories

**US-8.2.1: Send a direct message**
As an Admin, I want to send a message to a member who missed training so that I can check in on them.

- *Given* a member hasn't trained in 2 weeks
- *When* I send a message: "Hey — everything OK? We missed you at No-Gi"
- *Then* the message appears on their dashboard with a notification badge

**US-8.2.2: Member views message**
As a Member, I want to see messages from Admin on my dashboard so that I don't miss important communications.

- *Given* Admin sent me a message
- *When* I log into my dashboard
- *Then* I see a notification badge and can read the message

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

## Appendix A: Data Model

| Concept | Section | Storage | Description |
|---------|---------|---------|-------------|
| `Account` | §1 | accounts | User identity with role and password hash |
| `Member` | §1 | members | Profile: name, email, program_id, fee, frequency, status, belt_size, gi_size, rash_top_size, tshirt_size |
| `Program` | §1.3 | programs | Audience group: Adults, Kids, Youth. Determines class visibility, term structure, and comms targeting |
| `Class` | §1.3 | classes | Named session type within a program (Gi Express, Nuts & Bolts, etc.). Has duration and optional mat_hours_weight (default 1.0) |
| `Schedule` | §9.7 | schedules | Recurring weekly entry: day, time, class_id, coach_id, duration |
| `Term` | §1.3 | terms | NZ school term date ranges with manual confirmation |
| `Holiday` | §9.7 | holidays | Date ranges overriding schedule; auto-generates Notice |
| `Waiver` | §9.1 | waivers | Risk acknowledgement, valid 1 year |
| `Injury` | §9.2 | injuries | Red Flag body-part toggle, active 7 days |
| `Attendance` | §3.1 | attendance | Check-in record: member_id + class_id + date + time. Supports multi-session and un-check-in (soft delete). Mat hours = duration × class weight |
| `Notice` | §8.1 | notices | Unified notification: type (school_wide / class_specific / holiday), status (draft / published) |
| `Message` | §8.2 | messages | Direct in-app message from Admin to member |
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
| `Payment` | §13.1 | payments | Reconciled bank transactions |

---

## Appendix B: Implementation Priority

1. **§2–§3** — Kiosk, check-in (auto-select, multi-session, un-check-in), attendance, training log, mat hours (with weighting), estimated hours, historical attendance
2. **§4** — Belt progression (stripe inference, term-based grading with per-term reset, belt/stripe icons, proposals, admin overrides)
3. **§9** — Member management (waiver, red flags, member list with sizes, inactive radar, archive, schedule/holiday management, program assignment)
4. **§8** — Communication (notices with program targeting, messaging, coach observations)
5. **§5–§6** — Curriculum engine (versioned rotors with concurrent themes, topic queues, auto-advance, coach ownership, topic voting)
6. **§7** — Technical library & clips (cross-linked to topics, not hard-coupled)
7. **§10** — Calendar (events, competitions, rotor views, personal goals)
8. **§13** — Business operations (Xero, ROI dashboard, digital wallet)
9. **§11–§12** — Advanced study & competition tools

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
