# QA Report: Messaging

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 7

## Summary
Messages and inbox are simple, functional screens. Both use JS-driven rendering.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Messages h1 "My Messages" but inbox h1 "My Inbox" — two separate screens for similar concepts | member_messages.html:3, member_inbox.html:3 | Keep separate: Messages = coach/system notifications, Inbox = club emails. But add subtitle to clarify |
| 2 | P1 | Messages page has no subtitle explaining what messages are — confusing for new users | member_messages.html:3 | Add subtitle: "Notifications from coaches and the system" |
| 3 | P1 | Inbox page has no subtitle — users may wonder why there are two message screens | member_inbox.html:3 | Add subtitle: "Emails sent by the club" |
| 4 | P2 | Messages "Mark as Read" button is inline orange — consistent with brand | member_messages.html:29 | No action |
| 5 | P3 | Inbox toggle to show email body uses `[Show]` text toggle — could use a chevron icon for polish | member_inbox.html:33-34 | Low priority |
| 6 | P2 | Nav links: member nav has "Messages" (links to /messages). No explicit "Inbox" nav link — inbox accessed from messages page? | layout.html:178 | Verify inbox is discoverable; consider combining or adding nav link |

## Recommendations
1. **Add subtitles** to both Messages and Inbox pages to differentiate them
2. **Verify inbox discoverability** — ensure members can find /inbox from /messages or add to nav
