# Milestone Schedule GitHub Action

This action creates Issue Milestones in a GitHub repository, according to an
iCalendar (RFC 5545) Recurrence schedule.

## Inputs

### `recurrence`

**Required** The RFC 5545 Recurrence pattern defining a schedule of Milestone
due dates.

### `count`

**Required** The number of upcoming milestones that should be created each run.

### `format`

**Optional** A format string as specified in Golang's `time` package used to
determine each Milestone's title.
_defaults to `2006-01-02`_

## Example usage

```yaml
uses: MrGossett/github-action-milestone-schedule@v1
with:
  recurrence: 'FREQ=WEEKLY;BYDAY=FR;DTSTART=20191211T220000Z'
  format: 'Due 2006-Jan-02'
  count: 4
```
