**Epic:** S9 Member Management | **PRD:** Section 4.4 Belt & Stripe Icons

### User Story
As an Admin or Coach, I want to see a member's belt icon in the member list and on their profile so that I can quickly assess rank at a glance.

### Acceptance Criteria
- *Given* I am an Admin or Coach viewing /members
- *When* the member list renders
- *Then* each member row shows a belt icon representing their current belt (and stripes where applicable)
- *And* members without a configured belt show a neutral placeholder (no broken UI)

- *Given* I am viewing an individual member profile page
- *When* the profile renders
- *Then* the belt icon is shown next to the member's key identity info (name/email/etc)
- *And* it matches the belt shown elsewhere in the app

### Invariants
- Belt icon rendering must not break list layout on mobile
- Icon rendering must be consistent across pages (same source of truth for belt/stripe data)

### Pre-conditions
- Member belt/stripe data exists (or is absent and handled gracefully)
- Belt icon assets / CSS exist or are added per PRD Section 4.4

### Post-conditions
- Admin/Coach can visually identify belt rank in list and profile without extra clicks

### Test Plan

**Unit tests**
- [ ] Rendering helper outputs expected CSS class / SVG for a belt + stripe combination
- [ ] Missing belt data produces placeholder output

**Browser tests**
- [ ] Admin visits /members and sees belt icons in the list
- [ ] Admin opens a member profile and sees the belt icon next to member info

### Implementation Notes
- Likely affects templates for /members list row + member profile header
- Prefer reusing any existing belt icon rendering used in attendance/training log
