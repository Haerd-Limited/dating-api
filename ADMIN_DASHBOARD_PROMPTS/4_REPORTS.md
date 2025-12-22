# Admin Dashboard - Part 4: Reports Management

## Project Overview

Build a reports management interface for admins to review, resolve, and manage user safety reports. This is part 4 of the admin dashboard.

## Prerequisites

- Part 1 (Login) must be completed first
- Admin authentication via `X-Admin-Token` header is required
- All requests must include the admin token in headers

## API Base Configuration

- **Base URL**: `http://localhost:8080` (development) or your production URL
- **API Version**: `/api/v1`
- **Content-Type**: `application/json`
- **Authentication**: Include `X-Admin-Token` header in all requests

---

## API Endpoints

### 1. List Reports

**GET** `/api/v1/admin/reports`

**Query Parameters**:
- `status` (optional): Comma-separated list of statuses: `pending`, `in_review`, `resolved`, `escalated`, `dismissed`
  - Example: `?status=pending,in_review`
- `category` (optional): Comma-separated list of categories
- `reporter_id` (optional): Filter by reporter user ID
- `reported_id` (optional): Filter by reported user ID
- `limit` (optional, default: 50): Number of results per page
- `offset` (optional, default: 0): Pagination offset

**Response**:
```json
[
  {
    "id": "uuid-string",
    "reporter_user_id": "uuid-string",
    "reported_user_id": "uuid-string",
    "subject_type": "user",
    "subject_id": null,
    "category": "inappropriate_content",
    "notes": "User sent inappropriate messages",
    "status": "pending",
    "severity": "high",
    "metadata": {},
    "auto_action": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "resolved_at": null,
    "actions": []
  }
]
```

### 2. Get Single Report

**GET** `/api/v1/admin/reports/{reportID}`

**Response**: Same structure as single report object above, but with full action history:

```json
{
  "id": "uuid-string",
  "reporter_user_id": "uuid-string",
  "reported_user_id": "uuid-string",
  "subject_type": "user",
  "subject_id": null,
  "category": "inappropriate_content",
  "notes": "User sent inappropriate messages",
  "status": "in_review",
  "severity": "high",
  "metadata": {
    "message_id": "123",
    "conversation_id": "456"
  },
  "auto_action": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z",
  "resolved_at": null,
  "actions": [
    {
      "id": "action-uuid",
      "action_type": "warn_user",
      "action_data": {
        "warning_message": "User warned about inappropriate behavior"
      },
      "notes": "First warning issued",
      "reviewer_id": "admin-uuid",
      "created_at": "2024-01-02T10:00:00Z"
    }
  ]
}
```

### 3. Resolve Report

**POST** `/api/v1/admin/reports/{reportID}/resolve`

**Request Body**:
```json
{
  "action_type": "warn_user",
  "action_data": {
    "warning_message": "User warned about inappropriate behavior",
    "user_id": "reported-user-id"
  },
  "notes": "First warning issued. User has been notified.",
  "new_status": "resolved",
  "auto_action": null,
  "resolved_at": "2024-01-02T10:00:00Z"
}
```

**Status Values**:
- `pending`: Initial state
- `in_review`: Report is being reviewed
- `resolved`: Report has been resolved
- `escalated`: Report escalated to higher authority
- `dismissed`: Report dismissed as invalid

**Response**:
```json
{
  "status": "resolved"
}
```

### 4. Get Report Categories (Public Endpoint)

**GET** `/api/v1/lookup/report-categories`

**Note**: This is a public endpoint, no authentication required.

**Response**:
```json
{
  "report_categories": [
    {
      "id": 1,
      "key": "inappropriate_content",
      "label": "Inappropriate Content",
      "sort_order": 1
    },
    {
      "id": 2,
      "key": "harassment",
      "label": "Harassment",
      "sort_order": 2
    }
  ]
}
```

---

## Subject Types

- **user**: Report about a user
- **message**: Report about a specific message
- **profile**: Report about a user's profile

---

## UI Requirements

### Main Page: Reports Queue

**Layout**:
- **Sidebar Navigation** (left side):
  - Logo/Brand name
  - "Reports" (active)
  - Navigation to other sections (Video Verifications, Analytics)
  - Logout button (bottom)
  
- **Main Content Area**:
  - **Page Header**: 
    - Title: "Safety Reports"
    - Subtitle showing counts: "X pending, Y in review"
    - "New Reports" badge if there are unread pending reports
  
  - **Filters Section** (horizontal bar, collapsible):
    - **Status Filter** (multi-select dropdown):
      - Options: All, Pending, In Review, Resolved, Escalated, Dismissed
      - Show selected count badge
      - Clear all button
    
    - **Category Filter** (dropdown):
      - Load categories from `/api/v1/lookup/report-categories`
      - Multi-select or single select
      - "All Categories" option
    
    - **Reporter ID Search**:
      - Input field with search icon
      - Placeholder: "Search by reporter ID"
      - Clear button
    
    - **Reported User ID Search**:
      - Input field with search icon
      - Placeholder: "Search by reported user ID"
      - Clear button
    
    - **Date Range Picker** (optional):
      - From/To date inputs
      - Quick filters: Today, Last 7 Days, Last 30 Days
    
    - **"Clear All Filters"** button
    - **"Apply Filters"** button (if not auto-applying)
  
  - **Reports Table**:
    - **Columns**:
      - **ID** (truncated, e.g., "abc123...", clickable)
      - **Reporter** (user ID, truncated, clickable to view user)
      - **Reported** (user ID, truncated, clickable to view user)
      - **Category** (badge with category label)
      - **Subject Type** (badge: User, Message, Profile)
      - **Status** (badge with color coding):
        - pending = Yellow/Amber (#F59E0B)
        - in_review = Blue (#3B82F6)
        - resolved = Green (#10B981)
        - escalated = Red (#EF4444)
        - dismissed = Gray (#6B7280)
      - **Severity** (badge, if available):
        - high = Red
        - medium = Orange
        - low = Yellow
      - **Created At** (formatted: "2 days ago" or "Jan 15, 2024")
      - **Actions** (View Details button)
    
    - **Table Features**:
      - Sortable columns (optional)
      - Row hover effects
      - Click row to view details
      - Highlight unread/new reports
      - Bulk selection (optional)
    
    - **Pagination** (bottom):
      - Previous/Next buttons
      - Page numbers
      - "Showing X-Y of Z results"
      - Items per page selector
    
    - **Empty State**:
      - Message: "No reports found"
      - Illustration or icon
      - "Clear filters" suggestion if filters are active
    
    - **Loading State**:
      - Skeleton loaders for table rows
      - Loading spinner

### Detail/Review Modal/Page

When clicking "View Details" on a report:

**Layout** (Full Page or Large Modal):
- **Header Section**:
  - Report ID (with copy button)
  - Status badge (editable)
  - Back button to list
  - Close button (if modal)
  - Priority indicator (if severity is high)

- **Main Content** (Two-Column or Stacked Layout):

  **Left Column - Report Details**:
  - **Report Information Card**:
    - **Reporter**:
      - User ID (clickable, opens user profile)
      - "View Reporter Profile" button
    - **Reported User**:
      - User ID (clickable, opens user profile)
      - "View Reported User Profile" button
    - **Subject Type**: Badge (User, Message, Profile)
    - **Subject ID**: Display if available (clickable if it's a message)
    - **Category**: Badge with label
    - **Severity**: Badge (if available)
  
  - **Report Content Card**:
    - **Notes**: Full text display
    - **Metadata**: Expandable section showing JSON metadata
    - **Auto Action**: Display if available
  
  - **Timeline Card**:
    - **Created At**: Date/time
    - **Updated At**: Date/time
    - **Resolved At**: Date/time (if resolved)
    - **Action History**: 
      - Timeline view of all actions
      - Each action shows:
        - Action type
        - Reviewer ID
        - Notes
        - Timestamp
        - Action data (expandable)

  **Right Column - Resolution Actions**:
  - **Current Status**: Large badge display
  - **Resolution Form**:
    - **Status Dropdown**:
      - Options: pending, in_review, resolved, escalated, dismissed
      - Required field
      - Show confirmation if changing from resolved to another status
    
    - **Action Type** (required):
      - Input field or dropdown
      - Examples: "warn_user", "suspend_user", "ban_user", "no_action"
      - Help text explaining each action type
    
    - **Action Data** (optional):
      - JSON editor or structured form
      - Fields depend on action type
      - Example for "warn_user":
        - warning_message (textarea)
        - user_id (auto-filled)
    
    - **Notes** (optional):
      - Large textarea
      - Placeholder: "Internal notes about this resolution"
      - Character counter (optional)
    
    - **Resolved At** (optional):
      - Date/time picker
      - Default: Current time
      - Only shown if status is "resolved"
    
    - **Submit Button**:
      - Primary action button
      - Shows loading state
      - Confirmation dialog before submitting
      - Success message after submission

- **Quick Actions** (optional):
  - "Mark as In Review" button
  - "Escalate" button
  - "Dismiss" button
  - Quick action buttons for common resolutions

### Design Requirements

1. **Color Scheme**:
   - Status badges: 
     - Pending: Yellow/Amber (#F59E0B)
     - In Review: Blue (#3B82F6)
     - Resolved: Green (#10B981)
     - Escalated: Red (#EF4444)
     - Dismissed: Gray (#6B7280)
   - Severity badges:
     - High: Red (#EF4444)
     - Medium: Orange (#F97316)
     - Low: Yellow (#F59E0B)
   - Primary actions: Blue (#3B82F6)
   - Danger actions: Red (#EF4444)

2. **Responsive Design**:
   - Desktop: Full table with all columns
   - Tablet: Condensed table, hide less important columns
   - Mobile: Card view instead of table, stack details

3. **Loading States**:
   - Skeleton loaders for table rows
   - Spinner for detail view
   - Button loading states during API calls
   - Loading overlay for filters

4. **Error Handling**:
   - Toast notifications for errors
   - Inline error messages for forms
   - Retry buttons for failed API calls
   - Validation errors in form fields

5. **User Experience**:
   - Auto-refresh queue every 30-60 seconds (optional toggle)
   - Keyboard shortcuts (Esc to close modal)
   - Confirmation dialogs for status changes
   - Success feedback after resolution
   - Undo action (optional, if supported by API)
   - Bulk actions (select multiple, resolve together)

### Report Categories Display

- Load categories from lookup endpoint on page load
- Cache categories to avoid repeated API calls
- Display category labels (not keys) in UI
- Group similar categories if needed

### Action History Timeline

- Visual timeline showing all actions taken
- Most recent actions at top
- Each action as a card with:
  - Icon based on action type
  - Action type label
  - Reviewer information
  - Timestamp (relative: "2 hours ago")
  - Notes (if available)
  - Expandable action data

---

## Implementation Notes

1. **State Management**:
   - Use React Query/SWR for API state management
   - Cache report categories
   - Optimistic updates for status changes
   - Invalidate cache after resolution

2. **Performance**:
   - Virtual scrolling for large report lists
   - Debounce search inputs
   - Paginate results
   - Lazy load detail view

3. **Filtering**:
   - Build filter object from UI state
   - Update URL query params (optional, for shareable links)
   - Persist filter preferences in localStorage
   - Clear filters button resets to defaults

4. **Form Validation**:
   - Validate status is selected
   - Validate action type is provided
   - Validate action data matches action type
   - Show helpful error messages

5. **Security**:
   - Never expose sensitive user data unnecessarily
   - Validate admin token on every request
   - Sanitize user input in notes/action data

---

## Example Component Structure

```
/admin
  /reports
    - ReportsList.tsx
    - ReportDetail.tsx (Full Page or Modal)
    - ReportFilters.tsx
    - ReportTable.tsx
    - ResolutionForm.tsx
    - ActionTimeline.tsx
    - StatusBadge.tsx
  /components
    - Sidebar.tsx
    - LoadingSpinner.tsx
    - ConfirmationDialog.tsx
```

---

## Testing Checklist

- [ ] List reports with all filter combinations
- [ ] Pagination works correctly
- [ ] Report detail view loads correctly
- [ ] Action history displays correctly
- [ ] Resolve report with all status options
- [ ] Form validation works
- [ ] Status updates reflect immediately
- [ ] Error states handled gracefully
- [ ] Responsive on mobile/tablet
- [ ] Confirmation dialogs appear
- [ ] Success messages show after actions
- [ ] Categories load correctly
- [ ] Search by reporter/reported ID works

---

## Technical Stack Recommendations

- **Framework**: React with Next.js or Vue with Nuxt
- **UI Library**: Tailwind CSS + shadcn/ui or similar
- **State Management**: React Query/TanStack Query or SWR
- **HTTP Client**: Axios
- **Date Handling**: date-fns or dayjs
- **JSON Editor**: react-json-view or similar (for metadata/action data)

---

## Key Features Summary

1. **Reports List**: Filterable, sortable, paginated table
2. **Report Details**: Full report information with action history
3. **Resolution System**: Form to resolve reports with actions
4. **Action History**: Timeline of all actions taken
5. **Filtering**: Comprehensive filter options
6. **Status Management**: Real-time status updates
7. **User Context**: Links to view reporter/reported user profiles

---

## Important Notes

1. **Subject Types**: Handle different subject types appropriately:
   - If subject_type is "message", show message preview if available
   - If subject_type is "profile", show profile link
   - If subject_type is "user", show user profile link

2. **Action Types**: Common action types might include:
   - `warn_user`: Issue warning to reported user
   - `suspend_user`: Temporarily suspend user
   - `ban_user`: Permanently ban user
   - `no_action`: Dismiss report without action
   - `escalate`: Escalate to higher authority

3. **Metadata**: Metadata field contains additional context:
   - May include message IDs, conversation IDs
   - May include screenshots or evidence URLs
   - Display in expandable/collapsible section

4. **Priority Handling**: 
   - High severity reports should be highlighted
   - Consider auto-refresh more frequently for pending reports
   - Show notification badge for new high-priority reports

---

**Note**: Reports management is critical for platform safety. Ensure the UI makes it easy to review reports thoroughly and take appropriate actions quickly.

