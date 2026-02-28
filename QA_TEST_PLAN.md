# UI Regression Test Plan

**Issue:** US-15.1.1  
**Goal:** Comprehensive UI regression test plan (Markdown)  
**Date:** 2026-03-01

---

## Test Environment

- **URL:** http://localhost:8080 (or https://app.workshopjiujitsu.co.nz)
- **Roles to test:** Admin, Coach, Member, Trial
- **Browser:** Chrome/Firefox latest
- **Viewport:** Desktop (1920x1080), Tablet (768px), Mobile (375px)

---

## Test Matrix

| Area | Page | URL | Admin | Coach | Member | Trial |
|------|------|-----|-------|-------|--------|-------|
| **1. Layout & Navigation** | | | | | | |
| | Login | /login | ✓ | ✓ | ✓ | ✓ |
| | Dashboard | /dashboard | ✓ | ✓ | ✓ | ✓ |
| | Layout/Nav | All pages | ✓ | ✓ | ✓ | ✓ |
| **2. Authentication** | | | | | | |
| | Change Password | /change-password | ✓ | ✓ | ✓ | ✓ |
| | Activate | /activate | - | - | - | ✓ |
| **3. Member Management** | | | | | | |
| | Member List | /members | ✓ | ✓ | - | - |
| | Member Profile | /members/profile | ✓ | ✓ | ✓ | - |
| | Register Member | /members/register | ✓ | - | - | - |
| | Check-in Form | /checkin/form | ✓ | ✓ | - | - |
| | Injury Form | /injuries/form | ✓ | ✓ | ✓ | ✓ |
| | Waiver Form | /waivers/form | ✓ | ✓ | ✓ | ✓ |
| **4. Attendance & Training** | | | | | | |
| | Today's Attendance | /attendance | ✓ | ✓ | - | - |
| | Training Log | /training-log | ✓ | ✓ | ✓ | - |
| | Kiosk | /kiosk | ✓ | ✓ | - | - |
| **5. Calendar & Goals** | | | | | | |
| | Calendar | /calendar | ✓ | ✓ | ✓ | ✓* |
| **6. Content & Library** | | | | | | |
| | Themes | /themes | ✓ | ✓ | ✓ | ✓ |
| | Technical Library | /library | ✓ | ✓ | ✓ | - |
| | Curriculum | /curriculum | ✓ | ✓ | ✓ | - |
| **7. Messaging** | | | | | | |
| | Messages | /messages | ✓ | ✓ | ✓ | - |
| | Inbox | /inbox | ✓ | ✓ | ✓ | - |
| **8. Admin: Email** | | | | | | |
| | Email List | /admin/emails | ✓ | - | - | - |
| | Compose Email | /admin/emails/compose | ✓ | - | - | - |
| | Email Template | /admin/emails/template | ✓ | - | - | - |
| **9. Admin: Training** | | | | | | |
| | Grading | /admin/grading | ✓ | - | - | - |
| | Milestones | /admin/milestones | ✓ | - | - | - |
| | Self Estimates | /admin/self-estimates | ✓ | - | - | - |
| | Schedules | /admin/schedules | ✓ | - | - | - |
| | Class Types | /admin/class-types | ✓ | - | - | - |
| **10. Admin: Settings** | | | | | | |
| | Accounts | /admin/accounts | ✓ | - | - | - |
| | Features | /admin/features | ✓ | - | - | - |
| | Terms | /admin/terms | ✓ | - | - | - |
| | Holidays | /admin/holidays | ✓ | - | - | - |
| | Inactive Members | /admin/inactive | ✓ | - | - | - |
| | Notices | /admin/notices | ✓ | - | - | - |
| | Performance | /admin/perf | ✓ | - | - | - |
| **11. Privacy** | | | | | | |
| | Delete Request | /privacy/delete | ✓ | ✓ | ✓ | ✓ |
| | Consent | /privacy/consent | ✓ | ✓ | ✓ | ✓ |

*Feature flagged

---

## Test Scenarios (Per Page)

### Login Page (/login)
- [ ] Page loads with "Log In" heading
- [ ] Email/password fields present with labels
- [ ] CSRF token present
- [ ] Submit button works
- [ ] Error messages display for invalid credentials
- [ ] "Forgot password" link (if applicable)
- [ ] Mobile responsive layout

### Dashboard (/dashboard)
- [ ] Shows role-appropriate content
- [ ] Quick links work
- [ ] Today's classes display
- [ ] Recent notices show
- [ ] Mobile layout works

### Member List (/members)
- [ ] Table displays with columns: Name, Email, Program, Status, Injury
- [ ] Search works
- [ ] Filters (Program, Status) work
- [ ] Sorting works on each column
- [ ] Pagination works
- [ ] Export CSV button (admin/coach only)
- [ ] Import CSV button (admin only)
- [ ] Register Member button (admin only)
- [ ] Belt icons display correctly
- [ ] Clicking name navigates to profile

### Member Profile (/members/profile)
- [ ] Profile info displays
- [ ] Belt/stripes show correctly
- [ ] Attendance history shows
- [ ] Edit button (admin only)
- [ ] Archive/Restore buttons (admin only)

### Calendar (/calendar)
- [ ] Calendar grid displays
- [ ] Events show on correct dates
- [ ] Add event button (admin only)
- [ ] Event details modal works
- [ ] Personal goals section works
- [ ] Mobile view works

### Technical Library (/library)
- [ ] Clip grid displays
- [ ] Theme filter works
- [ ] Search works
- [ ] Promoted filter works
- [ ] Video player opens on clip click
- [ ] Mute/Unmute works
- [ ] Loop works
- [ ] Add clip form works (admin/coach/member)
- [ ] Promote button works (admin/coach)

### Curriculum (/curriculum)
- [ ] Rotor displays
- [ ] Theme carousel works
- [ ] Topic voting works (member+)
- [ ] Admin controls work

---

## Cross-Cutting Checks

### Terminology
- [ ] Page h1 matches nav link text
- [ ] "Log In" / "Log Out" consistent (not "Sign In" / "Logout")
- [ ] Button labels are verb-first
- [ ] Date formats consistent (DD Mon YYYY)

### Visual Consistency
- [ ] Primary color = orange (#F9B232)
- [ ] Secondary = dark (#1A1B1F)
- [ ] Danger = red
- [ ] Button styles consistent
- [ ] Table styles consistent

### Accessibility
- [ ] All form inputs have labels
- [ ] Hover states on interactive elements
- [ ] Focus indicators visible
- [ ] Mobile responsive doesn't break

### Empty States
- [ ] No data shows helpful message
- [ ] Not just blank/empty

---

## Findings Template

**Title:** [short, specific]
- **Role:** Admin/Coach/Member/Trial
- **Page:** /path
- **Steps to reproduce:**
  1. …
- **Expected:**
- **Actual:**
- **Severity:** blocker / high / medium / low
- **Screenshot/clip:** (attach)
- **Suggested fix:**
- **Notes / related issues:**

---

## Exit Criteria

- [ ] All nav-accessible pages verified for all roles
- [ ] No blocker/high-severity findings remain open
- [ ] Smoke automation exists and runs in CI (see US-15.2.1)
