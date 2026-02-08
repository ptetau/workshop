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
| **Member** | Active paying student. Can check in, flag injuries, view schedule, track flight time, request themes, set personal goals, and access study tools. |
| **Trial** | Prospective student. Can check in, sign waiver, and view schedule. No hard visit limit — Admin manually converts to Member when ready. |
| **Guest** | Drop-in visitor. A lightweight account (name + email) is created during the waiver flow. If they return, they are recognised and prompted to convert to Trial or Member. |

### 1.2 Member Statuses

| Status | Description |
|--------|-------------|
| **Active** | Currently training. Appears in kiosk search, member lists, and all active views. |
| **Inactive** | Has stopped checking in (flagged after a configurable number of days). Still visible in member lists with an "inactive" indicator. |
| **Archived** | Manually archived by Admin. Hidden from all active views and kiosk search. All data preserved. Can be restored to Active at any time. |

### 1.3 Programs & Schedule

Workshop runs multiple programs, each of which can have its own independent curriculum rotor (see §5). Programs are extensible by Admin.

| Program | Description | Duration | Audience |
|---------|-------------|----------|----------|
| **Nuts & Bolts** | Beginner skills development (Gi) | 60 min | Adults — all levels welcome |
| **Gi Express** | All-levels skills development and sparring (Gi) | 60 min | Adults — all levels |
| **NoGi Express** | All-levels skills development and sparring (NoGi) | 60 min | Adults — all levels |
| **Feet-to-Floor** | Skills for taking the match to the ground | 60 min | Adults — all levels |
| **NoGi Long Games** | Specific sparring and competition prep (NoGi) | 60 min | Adults — experience required |
| **BYO Task Based Games** | Practice game design and specific skills (Gi or NoGi) | 60 min | Adults — experience recommended |
| **Open Mat** | Gi or NoGi sparring | 60–180 min | Adults — experience required |
| **Workshop Kids** | Kids program | 45 min | Ages 6–11, runs during NZ school terms |
| **Workshop Youth** | Youth program | 45–60 min | Ages 12–16, runs during NZ school terms |

**Default Weekly Timetable (from workshopjiujitsu.co.nz):**

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
| **Flight time** | Total accumulated mat hours for a member — the canonical metric for adult grading. Includes recorded hours (from check-ins) and estimated hours (admin/self-submitted). |
| **Session** | A specific class time-slot on a given day (e.g., "Monday 6 PM Nuts & Bolts"). Resolved on-the-fly from Schedule + Terms − Holidays. |
| **Program** | A named class type that appears on the timetable (Gi Express, Kids, etc.). Each program can have its own rotor. |
| **Rotor** | A cycling curriculum of themes assigned to a program. Coaches advance the rotor manually. |
| **Theme** | A technical block within a rotor with a configurable duration (days or weeks). Contains topics. |
| **Topic** | A specific technique or drill within a theme, with a frequency weight. |
| **Promotion** | The act of advancing a member to a new belt. Proposed by Coach, approved by Admin. |
| **Stripe** | A progress marker within a belt level (0–4). For adults, inferred automatically from flight time. |

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

Each check-in creates an attendance record linking a member to a resolved session (program + date + time). Flight time is calculated from the session's configured duration.

**Access:** Admin ✓ (view all) | Coach ✓ (view all) | Member ✓ (own) | Trial ✓ (own) | Guest —

#### User Stories

**US-3.1.1: Attendance creates flight time**
As a Member, I want my check-in to automatically add the session's duration to my flight time so that my training hours are tracked.

- *Given* I check into "6:00 PM Nuts & Bolts" (60 min)
- *When* the attendance record is created
- *Then* 1 hour is added to my recorded flight time

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

A member-facing projection of their attendance data: classes attended, flight time (split into recorded and estimated), streaks, belt/stripe icon, and grading progress.

**Access:** Admin — | Coach — | Member ✓ | Trial ✓ | Guest —

#### User Stories

**US-3.3.1: View training log**
As a Member, I want to see my training log so that I know how much I've trained and how close I am to my next grade.

- *Given* I have attended 47 classes this year
- *When* I open my training log
- *Then* I see: 47 classes, 120 flight time hours (108 recorded + 12 estimated), current 3-week streak, Blue belt with 2 inferred stripes, and a progress bar toward Purple eligibility

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
- *Then* any member who reaches 200 flight time hours will receive the badge

### 3.4 Estimated Training Hours

For members with incomplete records, Admin or Coach can bulk-add estimated flight time by selecting a date range and entering estimated weekly hours. Overlaps with existing recorded training are handled explicitly.

**Access:** Admin ✓ | Coach ✓ | Member — (see §3.5 for self-estimates)

#### User Stories

**US-3.4.1: Bulk-add estimated hours**
As a Coach, I want to add estimated training hours for a member who hasn't been checking in so that their flight time reflects reality.

- *Given* a member trained Jan–Mar without checking in
- *When* I select the member, set the date range to Jan 1 – Mar 31, and enter "3 hours/week"
- *Then* the system calculates 13 weeks × 3 hours = 39 estimated hours
- *And* adds them to the member's flight time, tagged as `source: estimate`

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
As a Member who trained elsewhere while travelling, I want to submit an estimated training period so that my flight time stays accurate.

- *Given* I trained 3×/week for 6 weeks at a partner gym in São Paulo
- *When* I submit an estimate: date range, "3 hours/week", note: "Trained at Checkmat SP while travelling"
- *Then* the estimate appears in Admin's review queue with status "pending"
- *And* it is **not** added to my flight time until approved

**US-3.5.2: Admin reviews self-estimate**
As an Admin, I want to review member-submitted estimates so that I can verify their accuracy.

- *Given* a member submitted an estimate of 18 hours
- *When* I open the review queue and view the submission
- *Then* I see the date range, weekly hours, total, and the member's note
- *And* I can approve as-is, adjust the hours (e.g., reduce to 12), or reject with a reason

**US-3.5.3: Approved estimate tagged**
As an Admin, I want approved self-estimates to be clearly marked so that they are distinguishable from recorded hours.

- *Given* I approve a self-estimate
- *When* the hours are added to the member's flight time
- *Then* they are tagged as `source: self_estimate` and appear separately in the training log

---

## 4. Grading & Belt Progression

### 4.1 Belt Progression System

Tracks each member's current belt and progress toward their next grade. IBJJF-aligned.

**Adults (18+ years):** White → Blue → Purple → Brown → Black. 4 stripes per belt before promotion eligibility. Black belt has degrees per IBJJF.

| Belt | Min. Time at Belt | Stripes | Default Flight Time Threshold |
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
- *Then* a note is displayed: *"Grading requires minimum flight time. Exceptions may apply for active competitors."*

### 4.2 Stripe Inference (Adults)

Stripes are automatically inferred by pro-rating accumulated flight time across the configured threshold. No manual stripe awards needed — the system calculates stripe count from hours.

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
As a Member, I want my stripes to be calculated automatically from my flight time so that I always see my current progress.

- *Given* I am a Blue belt with a 150h threshold and I have 85 flight time hours
- *When* I view my profile or training log
- *Then* my belt icon shows 2 stripes (floor(85 / 37.5) = 2)
- *And* no manual stripe award was needed

**US-4.2.2: Stripe updates as hours accumulate**
As a Member, I want my stripe count to update automatically when I train more so that my progress is always current.

- *Given* I have 74 hours (Stripe 1) and I check into a 1-hour session
- *When* my flight time reaches 75 hours
- *Then* my stripe count updates to Stripe 2 across all views (training log, profile, attendance)

### 4.3 Term-Based Grading (Kids/Youth)

For term-based programs, Admin can toggle the grading metric between sessions attended and hours.

| Mode | Metric | Example |
|------|--------|---------|
| **Sessions** (default for kids) | Sessions attended in the current term | "24 of 30 sessions this term" |
| **Hours** | Accumulated flight time (same as adults) | "45 hours since last promotion" |

**Sessions mode**: eligibility is based on **% of term attendance** — the system counts total available sessions in the current NZ school term, divides by attendance, and compares to admin-configured thresholds.

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
- *Then* their eligibility is based on accumulated flight time instead of term attendance %
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

- *Given* several members are within 80% of their flight time threshold
- *When* I open the grading readiness list
- *Then* I see each member's name, belt/stripe icon, flight time progress (or term attendance %), and any grading notes
- *And* the list is sorted by proximity to eligibility

**US-4.5.2: Add grading notes**
As a Coach, I want to add notes to a member's grading readiness entry so that I can track areas for improvement.

- *Given* I am viewing a member's readiness entry
- *When* I add a note: "Needs to work on guard retention before promotion"
- *Then* the note is saved and visible to Admin and other coaches

**US-4.5.3: Propose promotion from readiness list**
As a Coach, I want to propose a promotion directly from the readiness list so that I don't have to navigate elsewhere.

- *Given* Sarah has 155 hours at Blue belt (threshold: 150h)
- *When* I click "Propose Promotion" and confirm "Blue → Purple"
- *Then* a grading proposal is created with status "pending"
- *And* Admin receives a notification to review the proposal

### 4.6 Grading Proposals & Promotions

Coaches propose promotions; Admin approves to make them official. The workflow prevents unilateral promotions.

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

**US-4.7.1: Add flight time credit**
As an Admin, I want to add flight time credit to a competitor's record so that they aren't sandbagging.

- *Given* a member is competing at Blue belt but dominating their division
- *When* I add 20 hours of flight time credit to their record
- *Then* their flight time increases, pushing them closer to Purple eligibility
- *And* the credit is logged with my Admin ID and a reason

**US-4.7.2: Force immediate promotion**
As an Admin, I want to force a promotion bypassing thresholds so that I can handle exceptional cases.

- *Given* a member is clearly above their belt level but hasn't met the hour threshold
- *When* I force a promotion from Blue to Purple
- *Then* their belt is updated immediately
- *And* the grading record shows method: "override" with my notes

**US-4.7.3: Adjust individual thresholds**
As an Admin, I want to adjust grading thresholds for a specific member so that competitors can be fast-tracked.

- *Given* a member is actively competing and training at a higher level
- *When* I reduce their flight time threshold from 150h to 120h for Blue → Purple
- *Then* their grading progress recalculates using the new threshold
- *And* the override is logged

---

## 5. Curriculum Rotor System

### 5.1 Program Rotors

Each program can have its own independent **rotor** — a cycling curriculum of themes. Admin creates rotors; coaches control advancement.

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

**Access:** Admin ✓ (create/manage) | Coach ✓ (control advancement) | Member ✓ (view if preview on) | Trial — | Guest —

#### User Stories

**US-5.1.1: Create rotor for a program**
As an Admin, I want to create a rotor for a program so that I can define its curriculum cycle.

- *Given* the "NoGi Long Games" program has no rotor
- *When* I create a new rotor with 6 themes, each 2 weeks long
- *Then* the rotor is saved and attached to "NoGi Long Games"
- *And* coaches can begin managing advancement

**US-5.1.2: View current theme**
As a Member, I want to see the current theme for a program so that I can prepare my mind before class.

- *Given* the Gi Express rotor is on Theme 1
- *When* I view the Gi Express program page
- *Then* I see "Leg Lasso Series — Week 2 of 2"
- *And* the topics for this theme are listed below

**US-5.1.3: Rotor cycles back**
As a Coach, I want the rotor to wrap back to the first theme after the last one so that the curriculum repeats.

- *Given* the rotor has 4 themes and I am on Theme 4
- *When* I advance the rotor
- *Then* the active theme becomes Theme 1
- *And* the cycle continues

### 5.2 Themes

A theme is a technical block within a rotor. Duration is configurable per theme as either **D days** or **W weeks**.

- **Short themes** (1–3 days): fun/surprise content, focused workshops
- **Standard themes** (1–4 weeks): deep technical blocks

#### User Stories

**US-5.2.1: Create a theme**
As an Admin, I want to create a theme within a rotor so that I can define the curriculum content.

- *Given* I am editing the Gi Express rotor
- *When* I create a new theme: name "Closed Guard Attacks", duration "2 weeks", visible
- *Then* the theme is added to the rotor in the specified position
- *And* I can add topics to it

**US-5.2.2: Theme with day duration**
As a Coach, I want to create a short theme lasting only 1 day so that I can schedule a focused workshop.

- *Given* I am editing the Kids rotor
- *When* I create a theme: "Takedown Day", duration "1 day", visible
- *Then* the theme occupies a single day in the rotor cycle

### 5.3 Topics & Frequency

Each theme contains topics (specific techniques or drills). Each topic has a **frequency weight F** controlling how often it appears within the theme's rotation.

- **F=1** (default): covered once per cycle through the theme's topics
- **F=2**: high priority — covered twice as often (appears twice in the rotation)

#### User Stories

**US-5.3.1: Add topics to a theme**
As a Coach, I want to add topics to a theme so that the daily class content is defined.

- *Given* I am editing the "Leg Lasso Series" theme
- *When* I add three topics: "Entry from DLR" (F=1), "Sweep to Mount" (F=2), "Lasso → Omoplata" (F=1)
- *Then* the theme has 3 topics, with "Sweep to Mount" appearing twice as often in the daily rotation

**US-5.3.2: Set topic frequency**
As a Coach, I want to set a topic's frequency to 2 so that the most important technique gets drilled more.

- *Given* "Leg Lasso Entry from DLR" has frequency F=1
- *When* I change it to F=2
- *Then* this topic appears twice as often in the rotation compared to F=1 topics

**US-5.3.3: Member views topics**
As a Member, I want to see the topics for the current theme so that I know what we'll be covering in class.

- *Given* the active theme is "Leg Lasso Series" with 3 topics
- *When* I view the program page (and preview is enabled)
- *Then* I see the list of topics for the current theme

### 5.4 Rotor Advancement

Coaches control when the rotor advances to the next theme via a manual "advance" action. The theme's configured duration is a guideline shown in the UI, not enforced automatically.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-5.4.1: Advance rotor**
As a Coach, I want to advance the rotor to the next theme so that the curriculum progresses.

- *Given* the Feet-to-Floor rotor is on "Wrestling Shots" (2 weeks, and 2 weeks have passed)
- *When* I tap "Advance Rotor"
- *Then* the active theme changes to "Clinch Takedowns"
- *And* the calendar view updates to show the new theme

**US-5.4.2: Extend a theme**
As a Coach, I want to keep a theme running longer than its configured duration if the class needs more time.

- *Given* "Wrestling Shots" is configured for 2 weeks but I want to run it for 3
- *When* I simply don't advance the rotor for an extra week
- *Then* the theme remains active until I manually advance it

### 5.5 Hidden / Surprise Themes

Themes can be marked as **hidden** by Admin or Coach. Hidden themes do not appear in member rotor previews. They are revealed only when they become the active theme.

**Access:** Admin ✓ (create) | Coach ✓ (create) | Member — | Trial — | Guest —

#### User Stories

**US-5.5.1: Create hidden theme**
As a Coach, I want to create a hidden theme so that it's a surprise for the class.

- *Given* I am editing the Kids rotor
- *When* I create a theme: "Dodgeball Day", duration "1 day", marked as hidden
- *Then* the theme does not appear in the kids' rotor preview
- *And* it is only revealed when I advance the rotor to it

**US-5.5.2: Surprise reveal**
As a Member, I want to be surprised by a hidden theme so that training stays fun.

- *Given* the rotor preview shows "Half Guard" next week
- *When* Monday arrives and the coach has advanced to a hidden theme instead
- *Then* I see "★ Sumo Games ★" as the active theme — it wasn't on the preview

### 5.6 Member Rotor Preview

Admin can toggle whether members can see the upcoming rotor schedule for each program.

- **Preview on**: members see the full sequence of upcoming themes
- **Preview off**: members only see the current active theme

**Access:** Admin ✓ (toggle) | Coach ✓ (view) | Member ✓ (view if enabled) | Trial — | Guest —

#### User Stories

**US-5.6.1: Toggle preview off**
As an Admin, I want to disable rotor preview for the Kids program so that the schedule is always a surprise.

- *Given* the Kids rotor preview is currently enabled
- *When* I toggle it off
- *Then* kids can only see the current theme, not upcoming ones

**US-5.6.2: View rotor preview**
As a Member, I want to see the upcoming themes for NoGi Express so that I can plan my training.

- *Given* the NoGi Express rotor has preview enabled
- *When* I view the NoGi Express program page
- *Then* I see the next 4 upcoming themes in order

---

## 6. Theme Requests & Voting

### 6.1 Requestable Item Menu

Admin configures a menu of items that members can request. Items are categorised as:

| Category | Description | Example |
|----------|-------------|---------|
| **Connection** | A positional relationship or guard/position | "K-Guard → SLX", "Closed Guard", "Half Guard — Deep Half" |
| **Action** | A transition, entry, attack, defense, or escape | "Leg Lasso Entry", "Arm Drag to Back Take", "Guillotine Defense" |

Each item tracks: **last covered date**, **request count**, and **vote count**.

**Access:** Admin ✓ (configure) | Coach ✓ (view) | Member ✓ (view) | Trial — | Guest —

#### User Stories

**US-6.1.1: Configure request menu**
As an Admin, I want to add items to the request menu so that members have a curated list of topics they can ask for.

- *Given* I navigate to the request menu configuration
- *When* I add "Heel Hook Defense" categorised as an Action
- *Then* the item appears in the member-facing request menu
- *And* its "last covered" date is empty until it runs for the first time

**US-6.1.2: Browse request menu**
As a Member, I want to browse the request menu so that I can see what's available to request.

- *Given* the menu has 20 items across Connections and Actions
- *When* I open the request menu and filter by "Connection"
- *Then* I see only Connection items, each showing when it was last covered

### 6.2 Member Requests & Voting

Members can request items (one per member per item) and vote on other members' requests (one vote per member per request).

**Access:** Admin ✓ (triage) | Coach ✓ (triage) | Member ✓ | Trial — | Guest —

#### User Stories

**US-6.2.1: Request a theme**
As a Member, I want to request a theme from the menu so that the coaches know what I want to work on.

- *Given* "De La Riva Guard" was last covered 6 weeks ago
- *When* I tap "Request" on that item
- *Then* my request is recorded (one request per member per item)
- *And* the item's request count increases by 1

**US-6.2.2: Vote on a request**
As a Member, I want to vote on another member's request so that popular topics rise to the top.

- *Given* another member requested "Arm Drag Series" and it has 7 votes
- *When* I vote on it
- *Then* it now has 8 votes
- *And* I cannot vote on this request again (one vote per member per request)

**US-6.2.3: View vote rankings**
As a Member, I want to see the current vote rankings so that I know what's popular.

- *Given* several items have been requested and voted on
- *When* I view the request board
- *Then* I see items sorted by vote count, with the most popular at the top

### 6.3 Vote-Driven Scheduling

When Admin or Coach chooses to run a voted theme, it is **brought forward** in the rotor. The displaced theme is inserted immediately after the voted theme, preserving the rest of the rotor order.

**Access:** Admin ✓ | Coach ✓ | Member — | Trial — | Guest —

#### User Stories

**US-6.3.1: Bring forward a voted theme**
As a Coach, I want to bring forward a popular request so that the class covers what members want.

- *Given* "De La Riva Guard" has 12 votes and the rotor's next theme is "Half Guard Sweeps"
- *When* I choose to bring forward "De La Riva Guard"
- *Then* Admin creates a theme for "De La Riva Guard" and inserts it as the next active theme
- *And* "Half Guard Sweeps" is pushed to the slot immediately after

**US-6.3.2: Request marked as scheduled**
As a Member, I want to see that my request has been scheduled so that I know it's coming.

- *Given* I requested "De La Riva Guard" and it was brought forward
- *When* I view the request board
- *Then* "De La Riva Guard" shows status "Scheduled"
- *And* its "last covered" date will update when the theme runs

**US-6.3.3: Admin triages requests**
As an Admin, I want to view requests sorted by vote count so that I can prioritise what the gym wants.

- *Given* 15 members voted for "De La Riva Guard" and 3 voted for "Rubber Guard"
- *When* I view the triage queue
- *Then* "De La Riva Guard" is at the top with 15 votes
- *And* I can choose to bring it forward or dismiss it

---

## 7. Technical Library & Clips

### 7.1 Clipping Tool

Paste a YouTube link, set start/end timestamps to create a single looping clip. Clips are associated with a theme.

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

All members split by program, with manual "Fee" and "Frequency" fields, injury indicators, waiver status, and member status.

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

### 12.2 Extra Session RSVP

Minimalist booking system for high-intensity or "Comp-Only" mat times.

**Access:** Admin ✓ | Coach ✓ | Member ✓ | Trial — | Guest —

#### User Stories

**US-12.2.1: RSVP for extra session**
As a Member, I want to RSVP for a competition prep session so that the coach knows how many are coming.

- *Given* there is a "Comp Prep — Saturday 2 PM" extra session
- *When* I tap "RSVP"
- *Then* I am registered and the coach sees the headcount

### 12.3 Scouting Pipeline

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
| `Member` | §1 | members | Profile: name, email, program, fee, frequency, status |
| `Waiver` | §9.1 | waivers | Risk acknowledgement, valid 1 year |
| `Injury` | §9.2 | injuries | Red Flag body-part toggle, active 7 days |
| `Attendance` | §3.1 | attendance | Check-in record: member + session + date + time. Supports multi-session and un-check-in (soft delete) |
| `Program` | §1.3 | programs | Named class type (Gi Express, Kids, etc.). Determines term structure, billing, and rotor assignment |
| `Schedule` | §9.7 | schedules | Recurring weekly entry: day, time, program, coach, duration |
| `Term` | §1.3 | terms | NZ school term date ranges with manual confirmation |
| `Holiday` | §9.7 | holidays | Date ranges overriding schedule; auto-generates Notice |
| `Notice` | §8.1 | notices | Unified notification: type (school_wide / class_specific / holiday), status (draft / published) |
| `Message` | §8.2 | messages | Direct in-app message from Admin to member |
| `GradingRecord` | §4.6 | grading_records | Promotion history: belt, stripe, date, proposed_by, approved_by, method |
| `GradingConfig` | §4.1 | grading_config | Per-belt thresholds: flight time or attendance %, stripe count, grading mode toggle |
| `GradingProposal` | §4.6 | grading_proposals | Coach-proposed promotion: member, target belt, notes, status |
| `EstimatedHours` | §3.4 | estimated_hours | Bulk-estimated flight time: date range, weekly hours, source, status, overlap mode, note |
| `Goal` | §10.3 | goals | Member target: description, target, unit, period, progress. Covers both recurring cadence and bounded targets |
| `Milestone` | §3.3 | milestones | Admin-configured achievement (e.g., "100 classes") |
| `CoachObservation` | §8.3 | coach_observations | Private per-member notes from Coach or Admin |
| `BeltConfig` | §4.4 | belt_config | Belt/stripe icon config: belt_name, colour, stripe_count, sort_order, age_range |
| `Rotor` | §5.1 | rotors | Curriculum cycle: program_id, name, current_theme_index, preview_enabled, created_by |
| `Theme` | §5.2 | themes | Technical block: rotor_id, name, duration_value, duration_unit, sort_order, hidden, created_by |
| `Topic` | §5.3 | topics | Technique within a theme: theme_id, name, frequency, sort_order |
| `ThemeRequestItem` | §6.1 | theme_request_items | Requestable item: name, category (connection/action), last_covered_date |
| `ThemeRequest` | §6.2 | theme_requests | Member request: item_id, member_id, status (open/scheduled/closed) |
| `ThemeVote` | §6.2 | theme_votes | Vote on a request: request_id, member_id. One per member per request |
| `Clip` | §7.1 | clips | YouTube timestamp loop associated with a theme |
| `Tag` | §11.1 | tags | Action/Connection metadata for clips |
| `ResearchEntry` | §11.4 | research_entries | Private research journal notes from 4-Up mode |
| `CalendarEvent` | §10.1 | calendar_events | Club event or competition: title, type, dates, registration_url |
| `Advice` | §12.1 | advice | Coach-curated strategy guides |
| `ExtraSession` | §12.2 | extra_sessions | Extra mat-time bookings |
| `Payment` | §13.1 | payments | Reconciled bank transactions |

---

## Appendix B: Implementation Priority

1. **§2–§3** — Kiosk, check-in (auto-select, multi-session, un-check-in), attendance, training log, estimated hours, historical attendance
2. **§4** — Grading (stripe inference, term-based toggle, belt/stripe icons, proposals, overrides)
3. **§9** — Member management (waiver, red flags, member list, inactive radar, archive, schedule/holiday management)
4. **§8** — Communication (notices, messaging, coach observations)
5. **§5–§6** — Curriculum engine (rotors, themes, topics, requests, voting, hidden themes)
6. **§7** — Technical library & clips
7. **§10** — Calendar (events, competitions, rotor views, personal goals)
8. **§13** — Business operations (Xero, ROI dashboard, digital wallet)
9. **§11–§12** — Advanced study & competition tools

---

## Appendix C: Open Questions

The following terminology and design decisions need resolution. Each is numbered with suggested answers lettered.

**Q1: Does all mat time count equally toward flight time?**
- (a) All equal — 1 hour = 1 hour regardless of session type
- (b) Weighted by session type — Admin configures multipliers
- (c) Capped at scheduled duration — no more, no less

**Q2: Should the rotor support 1:N (multiple rotors per program)?**
- (a) 1:1 — one rotor per program, keep it simple
- (b) 1:N — allow parallel rotors (e.g., technique + sparring focus)
- (c) 1:1 now, extensible to 1:N later

**Q3: Connection/Action taxonomy — shared between theme requests (§6) and clip tags (§11)?**
- (a) Same taxonomy, shared data — tags and requestable items draw from one list
- (b) Same terminology, separate data — independent lists
- (c) Different concepts — curriculum categories vs technique-level metadata

**Q4: Kids term grading — per-term or cumulative?**
- (a) Per-term — must hit threshold each term, resets each term
- (b) Cumulative since last promotion — terms are scheduling only
- (c) Per-term with carry-over — need N qualifying terms for eligibility
