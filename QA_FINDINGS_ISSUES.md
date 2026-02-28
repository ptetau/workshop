# QA Findings - Issues to Create

## P1 Issues (High Priority)

1. **Register form button colour** - Blue to orange
   - File: form_register_member.html:60-74
   - Fix: Change .btn-primary to `background: var(--orange)`

2. **Register form dead Cancel link** - /index.html to /members
   - File: form_register_member.html:27
   - Fix: Change href to `/members`

3. **Injury form dead Cancel link** - /index.html to /dashboard
   - File: form_report_injury.html:42
   - Fix: Change href to `/dashboard`

4. **Injury form UUID input** - Replace with autocomplete
   - File: form_report_injury.html:8-10
   - Fix: Replace with name-search autocomplete

5. **Waiver form button colour** - Blue to orange
   - File: form_sign_waiver.html:65-79
   - Fix: Change to `var(--orange)`

6. **Check-in form button colour** - Blue to orange
   - File: form_check_in_member.html:44,71-75
   - Fix: Change to orange theme

7. **Kiosk resetKiosk() bug** - Redundant classList.remove
   - File: kiosk.html:279-280
   - Fix: Remove the `remove('hidden')` line

8. **Delete confirmation missing** - No confirm dialog
   - Files: admin_terms.html:67, admin_holidays.html:67
   - Fix: Add `if (!confirm(...)) return;`

## P2 Issues (Medium Priority)

1. **Login h1 "Sign In" → "Log In"** - QA_02_AUTH #1
2. **Dashboard button font-size** - Standardise to 0.85rem - QA_02_AUTH #6
3. **Member quick-links vs nav** - Add Themes to member nav - QA_02_AUTH #9
4. **Injury form button colour** - Yellow to orange - QA_03 #5
5. **Waiver form Cancel link** - / to /dashboard - QA_03 #7
6. **Check-in form link colour** - Blue to orange - QA_03 #8
7. **Emoji in h1 headings** - Remove all emoji prefixes - QA_03 #10
8. **Member profile link colour** - Blue to orange - QA_03 #11
9. **Local style blocks** - Remove from legacy forms - QA_03 #13
10. **Belt CSS duplication** - Extract to shared location - QA_03 #14
11. **Attendance empty-state button** - Blue to orange - QA_04 #2
12. **Attendance links colour** - Blue to orange - QA_04 #3
13. **Themes h1** - "Theme Carousel" → "Themes" - QA_06 #1
14. **Curriculum h1** - "Curriculum Management" → "Curriculum" - QA_06 #5
