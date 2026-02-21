**Epic:** S3 Attendance and Training Log | **PRD:** Section 3.3 Training Log

### User Story
As a Member, I want graphs of my training volume so that I can understand my consistency and compare performance over time.

### Acceptance Criteria
- *Given* I am a Member viewing my attendance/training page
- *When* the page loads
- *Then* I see a graph of my training volume over time for the selected range

- *Given* the graph is shown
- *When* I switch the range to "Current month"
- *Then* the graph updates to show training volume over the current month

- *Given* the graph is shown
- *When* I switch the range to "Current year"
- *Then* the graph updates to show training volume over the current year

- *Given* I am viewing "Current month"
- *When* I enable comparison mode
- *Then* I can compare against "Last month" and "Same month last year"

- *Given* the graph is shown
- *When* comparison or context is enabled
- *Then* an additional line is shown for the average attendance of **all members** for the same time buckets

### Invariants
- Members can only view their own detailed attendance volume
- Aggregate averages must not reveal individual attendance records

### Pre-conditions
- Attendance data exists for the member (or returns an empty series)

### Post-conditions
- Member can visually track progress and compare periods

### Test Plan

**Unit tests**
- [ ] Aggregation: month range bucketing is correct
- [ ] Aggregation: year range bucketing is correct
- [ ] Comparison logic: last month and same month last year ranges are computed correctly
- [ ] All-member average query returns expected results

**Browser tests**
- [ ] Member switches month/year and graph updates
- [ ] Member enables comparison and sees both lines (self + all-members average)

### Implementation Notes
- Likely new API endpoint for graph data (member scoped) plus an aggregate query for all-member averages
- Consider caching/efficient aggregation for year view
