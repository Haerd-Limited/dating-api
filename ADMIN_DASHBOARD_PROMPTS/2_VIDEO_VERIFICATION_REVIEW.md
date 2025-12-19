# Admin Dashboard - Part 2: Video Verification Review

## Project Overview

Build a video verification review interface for admins to review user video verification submissions and compare them with user profile photos. This is part 2 of the admin dashboard.

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

### 1. List Video Verification Attempts

**GET** `/api/v1/admin/verification/video-attempts`

**Query Parameters**:
- `status` (optional): Comma-separated list of statuses: `pending`, `needs_review`, `passed`, `failed`
  - Example: `?status=pending,needs_review`
- `user_id` (optional): Filter by specific user ID
- `limit` (optional, default: 50): Number of results per page
- `offset` (optional, default: 0): Pagination offset

**Response**:
```json
[
  {
    "id": "uuid-string",
    "user_id": "uuid-string",
    "verification_code": "1234",
    "video_s3_key": "path/to/video.mp4",
    "status": "needs_review",
    "rejection_reason": null,
    "photos": [
      {
        "url": "https://s3.amazonaws.com/bucket/path/to/photo.jpg",
        "is_primary": true,
        "position": 1
      },
      {
        "url": "https://s3.amazonaws.com/bucket/path/to/photo2.jpg",
        "is_primary": false,
        "position": 2
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### 2. Get Single Video Attempt

**GET** `/api/v1/admin/verification/video-attempts/{attemptID}`

**Response**: Same structure as single item in list above

### 3. Get Presigned Video URL

**GET** `/api/v1/admin/verification/video-attempts/{attemptID}/video-url`

**Response**:
```json
{
  "video_url": "https://s3.amazonaws.com/bucket/path/to/video.mp4?presigned-params",
  "expires_in": 3600
}
```

**Note**: URL expires in 1 hour (3600 seconds)

### 4. Approve Video Attempt

**POST** `/api/v1/admin/verification/video-attempts/{attemptID}/approve`

**Request Body**:
```json
{
  "notes": "Looks good, user matches photos" // optional
}
```

**Response**:
```json
{
  "status": "passed",
  "message": "Video verification approved"
}
```

### 5. Reject Video Attempt

**POST** `/api/v1/admin/verification/video-attempts/{attemptID}/reject`

**Request Body**:
```json
{
  "rejection_reason": "User does not match profile photos", // required
  "notes": "Additional notes for internal use" // optional
}
```

**Response**:
```json
{
  "status": "failed",
  "message": "Video verification rejected"
}
```

---

## Status Values

- **pending**: Video not yet submitted
- **needs_review**: Video submitted, awaiting admin review
- **passed**: Approved by admin
- **failed**: Rejected by admin

---

## UI Requirements

### Main Page: Video Verification Queue

**Layout**:
- **Sidebar Navigation** (left side):
  - Logo/Brand name
  - "Video Verifications" (active/highlighted)
  - Navigation to other sections (Analytics, Reports)
  - Logout button (bottom)
  
- **Main Content Area**:
  - **Page Header**: 
    - Title: "Video Verification Review"
    - Subtitle showing count: "X submissions pending review"
  
  - **Filters Section** (horizontal bar):
    - Status filter dropdown: All, Pending, Needs Review, Passed, Failed
    - User ID search input (with search icon)
    - Date range picker (optional)
    - "Clear Filters" button
    - "Refresh" button
  
  - **Submissions Table**:
    - Columns:
      - ID (truncated, e.g., "abc123...", clickable to view details)
      - User ID (truncated, clickable)
      - Verification Code (4-digit code)
      - Status (badge with color coding):
        - pending = Yellow/Amber (#F59E0B)
        - needs_review = Orange (#F97316)
        - passed = Green (#10B981)
        - failed = Red (#EF4444)
      - Submitted At (formatted: "2 days ago" or "Jan 15, 2024")
      - Actions (View/Review button)
    - Pagination controls (bottom):
      - Previous/Next buttons
      - Page numbers
      - "Showing X-Y of Z results"
    - Empty state when no submissions
    - Loading skeleton while fetching

### Detail/Review Modal/Page

When clicking "View/Review" on a submission:

**Layout** (Full Page or Large Modal):
- **Header Section**:
  - Submission ID
  - Status badge
  - Back button to list
  - Close button (if modal)

- **Two-Column Layout**:
  
  **Left Column - User Photos**:
  - Section title: "User Profile Photos"
  - Grid of user photos (2-3 columns)
  - Each photo:
    - Thumbnail image
    - "Primary" badge if is_primary
    - Position number
    - Click to view full size (lightbox/modal)
  - Show message if no photos available
  
  **Right Column - Video Review**:
  - Section title: "Video Submission"
  - **Video Player Section**:
    - Video player (HTML5 video element)
    - Loading state while fetching presigned URL
    - Error state if video can't be loaded
    - Video controls (play, pause, volume, fullscreen, progress bar)
    - "Video expires in: X minutes" indicator
    - Refresh video URL button (if expired)
  
  - **Submission Details Section**:
    - User ID (with copy button)
    - Verification Code (prominent display)
    - Current Status (badge)
    - Submitted At
    - Rejection Reason (if rejected, shown in red)
  
  - **Review Actions Section**:
    - **Approve Button** (green, large):
      - Opens confirmation dialog
      - Optional notes textarea in dialog
      - Confirm/Cancel buttons
      - Shows success message on completion
    - **Reject Button** (red, large):
      - Opens rejection form:
        - Rejection reason (required, textarea, min 10 characters)
        - Additional notes (optional, textarea)
        - Confirm/Cancel buttons
      - Shows success message on completion

- **Comparison Helper** (optional):
  - Side-by-side view toggle
  - Zoom controls for photos
  - Video playback speed controls

### Design Requirements

1. **Color Scheme**:
   - Status badges: 
     - Pending: Yellow/Amber (#F59E0B)
     - Needs Review: Orange (#F97316)
     - Passed: Green (#10B981)
     - Failed: Red (#EF4444)
   - Primary actions: Blue (#3B82F6)
   - Danger actions: Red (#EF4444)
   - Background: Light gray (#F9FAFB)

2. **Responsive Design**:
   - Desktop: Two-column layout for photos and video
   - Tablet: Stacked layout
   - Mobile: Single column, full-width video player

3. **Loading States**:
   - Skeleton loaders for table rows
   - Spinner for video loading
   - Button loading states during API calls
   - Progress indicator for video buffering

4. **Error Handling**:
   - Toast notifications for errors
   - Inline error messages for forms
   - Retry buttons for failed API calls
   - "Video not available" message if URL fetch fails

5. **User Experience**:
   - Auto-refresh queue every 30 seconds (optional toggle)
   - Keyboard shortcuts (Enter to approve, Esc to close modal)
   - Smooth transitions and animations
   - Confirmation dialogs for destructive actions
   - Success feedback after approve/reject

### Photo Display

- Show all user photos in a grid
- Primary photo should be highlighted or shown first
- Photos should be clickable to view full size
- Use lazy loading for better performance
- Show loading placeholder while images load
- Handle broken image URLs gracefully

### Video Player Features

- Standard HTML5 video controls
- Fullscreen support
- Playback speed control (0.5x, 1x, 1.5x, 2x)
- Volume control
- Progress bar with seek functionality
- Auto-play on load (optional, user preference)
- Show video duration
- Handle different video formats (MP4, WebM)

---

## Implementation Notes

1. **Video Playback**:
   - Fetch presigned URL when opening detail view
   - Handle video loading errors gracefully
   - Show "Video not available" if URL fetch fails
   - Consider video format compatibility (MP4, WebM)
   - Cache video URLs (they expire, so handle refresh)

2. **State Management**:
   - Use React Query/SWR for API state management
   - Cache video URLs with expiration tracking
   - Optimistic updates for approve/reject actions
   - Invalidate cache after status changes

3. **Performance**:
   - Lazy load video only when detail view opens
   - Virtual scrolling for large submission lists
   - Debounce search/filter inputs
   - Paginate results (don't load all at once)
   - Image lazy loading for photos

4. **Security**:
   - Never expose S3 keys directly - always use presigned URLs
   - Validate admin token on every request
   - Clear sensitive data on logout

5. **Photo Comparison**:
   - Display photos prominently for easy comparison
   - Consider side-by-side layout option
   - Highlight primary photo
   - Show photo order/position

---

## Example Component Structure

```
/admin
  /video-verifications
    - VideoVerificationList.tsx
    - VideoVerificationDetail.tsx (Full Page)
    - VideoPlayer.tsx
    - PhotoGrid.tsx
    - ReviewActions.tsx
    - StatusBadge.tsx
    - FilterBar.tsx
  /components
    - Sidebar.tsx
    - LoadingSpinner.tsx
    - ConfirmationDialog.tsx
```

---

## Testing Checklist

- [ ] List submissions with filters
- [ ] Pagination works correctly
- [ ] User photos display correctly
- [ ] Video loads and plays
- [ ] Video URL refresh works when expired
- [ ] Approve action works
- [ ] Reject action works with required reason
- [ ] Status updates reflect immediately
- [ ] Error states handled gracefully
- [ ] Responsive on mobile/tablet
- [ ] Keyboard shortcuts work
- [ ] Confirmation dialogs appear
- [ ] Success messages show after actions

---

## Technical Stack Recommendations

- **Framework**: React with Next.js or Vue with Nuxt
- **UI Library**: Tailwind CSS + shadcn/ui or similar
- **State Management**: React Query/TanStack Query or SWR
- **HTTP Client**: Axios
- **Video Player**: HTML5 video or react-player
- **Image Gallery**: react-image-gallery or similar
- **Date Handling**: date-fns or dayjs

---

## Key Features Summary

1. **List View**: Filterable, paginated table of video submissions
2. **Detail View**: Full review interface with photos and video
3. **Photo Display**: Grid of user profile photos for comparison
4. **Video Player**: Full-featured video player with controls
5. **Review Actions**: Approve/reject with notes and reasons
6. **Status Management**: Real-time status updates
7. **Error Handling**: Comprehensive error states and recovery

---

**Note**: This feature is critical for platform safety. Ensure the UI makes it easy to compare user photos with video submissions accurately.
