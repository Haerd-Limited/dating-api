# Admin Dashboard - Part 3: Analytics Dashboard

## Project Overview

Build an analytics dashboard for admins to view user retention statistics, cohort analysis, and user engagement metrics. This is part 3 of the admin dashboard.

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

### 1. Get Retention Stats

**GET** `/api/v1/admin/analytics/retention`

**Query Parameters**:
- `from` (required): Start date in format `YYYY-MM-DD`
- `to` (required): End date in format `YYYY-MM-DD`

**Response**:
```json
{
  "total_users": 10000,
  "dau": 5000,
  "wau": 15000,
  "mau": 45000,
  "average_time_to_first_return": "2h30m15s"
}
```

**Field Descriptions**:
- `total_users`: Total number of registered users
- `dau`: Daily Active Users (users who opened the app on the specified date range)
- `wau`: Weekly Active Users
- `mau`: Monthly Active Users
- `average_time_to_first_return`: Average time between signup and first app open (ISO 8601 duration format)

### 2. Get Retention Cohorts

**GET** `/api/v1/admin/analytics/retention/cohorts`

**Query Parameters**:
- `signup_date` (required): Date in format `YYYY-MM-DD` - the signup date of the cohort
- `days_after` (optional, default: 7): Number of days after signup to measure retention

**Response**:
```json
{
  "signup_date": "2024-01-01",
  "day_1_retention": 45.5,
  "day_7_retention": 30.2,
  "day_30_retention": 15.8,
  "cohort_size": 1000
}
```

**Field Descriptions**:
- `signup_date`: The date users in this cohort signed up
- `day_1_retention`: Percentage of users who returned 1 day after signup
- `day_7_retention`: Percentage of users who returned 7 days after signup
- `day_30_retention`: Percentage of users who returned 30 days after signup
- `cohort_size`: Number of users in this cohort

### 3. Get Global Retention Stats

**GET** `/api/v1/admin/analytics/retention/global`

**Query Parameters**:
- `date` (optional, default: today): Date in format `YYYY-MM-DD`

**Response**: Same structure as Get Retention Stats

### 4. Get User Retention Profile

**GET** `/api/v1/admin/analytics/retention/users/{userID}`

**Response**:
```json
{
  "user_id": "uuid-string",
  "signup_date": "2024-01-01T00:00:00Z",
  "first_app_open": "2024-01-01T12:00:00Z",
  "latest_app_open": "2024-01-15T10:30:00Z",
  "total_app_opens": 50,
  "average_time_between_opens": "2h30m15s",
  "days_since_last_open": 5
}
```

**Field Descriptions**:
- `user_id`: The user's unique identifier
- `signup_date`: When the user registered (ISO 8601 datetime)
- `first_app_open`: First time user opened the app (ISO 8601 datetime, nullable)
- `latest_app_open`: Most recent app open (ISO 8601 datetime, nullable)
- `total_app_opens`: Total number of times user opened the app
- `average_time_between_opens`: Average time between app opens (ISO 8601 duration format, nullable)
- `days_since_last_open`: Number of days since last app open (nullable)

---

## UI Requirements

### Main Analytics Dashboard Page

**Layout**:
- **Sidebar Navigation** (left side):
  - Logo/Brand name
  - "Analytics" (active)
  - Navigation to other sections (Video Verifications, Reports)
  - Logout button (bottom)
  
- **Main Content Area**:

  **Page Header**:
  - Title: "Analytics Dashboard"
  - Last updated timestamp
  - Refresh button

  **Section 1: Retention Overview**
  - **Date Range Picker**:
    - From date input (YYYY-MM-DD format)
    - To date input (YYYY-MM-DD format)
    - "Apply" button
    - "Today", "Last 7 Days", "Last 30 Days" quick select buttons
    - Default: Last 30 days
  
  - **Key Metrics Cards** (4 cards in a row):
    - **Total Users**:
      - Large number display
      - Icon: Users icon
      - Color: Blue
      - Subtitle: "Registered users"
    
    - **Daily Active Users (DAU)**:
      - Large number display
      - Icon: Activity icon
      - Color: Green
      - Subtitle: "Active today"
      - Trend indicator (up/down arrow, optional)
    
    - **Weekly Active Users (WAU)**:
      - Large number display
      - Icon: Calendar icon
      - Color: Purple
      - Subtitle: "Active this week"
      - Trend indicator (optional)
    
    - **Monthly Active Users (MAU)**:
      - Large number display
      - Icon: Chart icon
      - Color: Orange
      - Subtitle: "Active this month"
      - Trend indicator (optional)
  
  - **Average Time to First Return**:
    - Card or section showing the duration
    - Format: "2 hours 30 minutes 15 seconds"
    - Visual representation (optional): Progress bar or gauge

  **Section 2: Cohort Analysis**
  - **Input Section**:
    - Signup date picker (YYYY-MM-DD format)
    - Days after selector (dropdown: 1, 7, 14, 30, 60, 90)
    - "Analyze Cohort" button
  
  - **Cohort Results Display**:
    - **Cohort Size**: Large number with label
    - **Retention Metrics Cards** (3 cards):
      - **Day 1 Retention**:
        - Percentage with large font
        - Progress bar visualization
        - Color: Green for high, Yellow for medium, Red for low
      
      - **Day 7 Retention**:
        - Percentage with large font
        - Progress bar visualization
        - Color coding
      
      - **Day 30 Retention**:
        - Percentage with large font
        - Progress bar visualization
        - Color coding
    
    - **Visual Chart** (optional but recommended):
      - Line chart or bar chart showing retention over time
      - X-axis: Days after signup
      - Y-axis: Retention percentage
      - Multiple lines for different cohorts (if comparing)

  **Section 3: User Retention Profile**
  - **Search Section**:
    - User ID input field
    - Search button
    - "Clear" button
  
  - **User Profile Display** (shown after search):
    - **User Information Card**:
      - User ID (with copy button)
      - Signup Date (formatted: "January 1, 2024")
      - Account Age (calculated: "15 days")
    
    - **Activity Metrics**:
      - **First App Open**: Date/time or "Never" if null
      - **Latest App Open**: Date/time or "Never" if null
      - **Total App Opens**: Large number
      - **Average Time Between Opens**: Duration or "N/A"
      - **Days Since Last Open**: Number with color coding:
        - Green: < 1 day
        - Yellow: 1-7 days
        - Orange: 7-30 days
        - Red: > 30 days
    
    - **Activity Timeline** (optional):
      - Visual timeline of app opens
      - Dots or bars representing opens over time

### Design Requirements

1. **Color Scheme**:
   - Metric cards: Use distinct colors for each metric
   - Retention percentages: Color-coded based on value
   - Charts: Use accessible color palette
   - Background: Light gray (#F9FAFB)

2. **Data Visualization**:
   - Use charts library (Recharts, Chart.js, or similar)
   - Ensure charts are responsive
   - Include tooltips on hover
   - Show data labels where appropriate
   - Use consistent color scheme

3. **Responsive Design**:
   - Desktop: Multi-column layout
   - Tablet: 2-column layout for metric cards
   - Mobile: Single column, stacked cards

4. **Loading States**:
   - Skeleton loaders for metric cards
   - Spinner for charts
   - Loading overlay for date range changes
   - Disable inputs during loading

5. **Error Handling**:
   - Toast notifications for errors
   - Inline error messages
   - Retry buttons for failed API calls
   - Empty states with helpful messages

6. **User Experience**:
   - Auto-format date inputs
   - Validate date ranges (from < to)
   - Show helpful placeholder text
   - Disable future dates
   - Format large numbers with commas (e.g., 10,000)
   - Format durations in human-readable format

### Date Formatting

- Display dates in user-friendly format: "January 15, 2024"
- Show relative time where appropriate: "2 days ago"
- Use consistent date format throughout
- Handle timezone properly (display in user's timezone or UTC)

### Number Formatting

- Format large numbers: 1,000 instead of 1000
- Show percentages with 1 decimal place: 45.5%
- Format durations: "2h 30m 15s" or "2 hours 30 minutes"

---

## Implementation Notes

1. **Date Handling**:
   - Use date-fns or dayjs for date manipulation
   - Validate date ranges before API calls
   - Handle timezone conversions properly
   - Format dates consistently

2. **State Management**:
   - Use React Query/SWR for API state management
   - Cache analytics data (refresh every 5 minutes)
   - Optimize re-renders for charts

3. **Performance**:
   - Lazy load charts
   - Debounce date range changes
   - Memoize expensive calculations
   - Virtualize long lists if needed

4. **Charts**:
   - Use a reliable charting library
   - Ensure accessibility (ARIA labels, keyboard navigation)
   - Make charts responsive
   - Include legends and axis labels

5. **Data Refresh**:
   - Auto-refresh every 5 minutes (optional toggle)
   - Manual refresh button
   - Show last updated timestamp

---

## Example Component Structure

```
/admin
  /analytics
    - AnalyticsDashboard.tsx
    - RetentionOverview.tsx
    - CohortAnalysis.tsx
    - UserRetentionProfile.tsx
    - MetricCard.tsx
    - RetentionChart.tsx
    - DateRangePicker.tsx
  /components
    - Sidebar.tsx
    - LoadingSpinner.tsx
```

---

## Testing Checklist

- [ ] Date range picker works correctly
- [ ] Retention stats display correctly
- [ ] Cohort analysis works with different dates
- [ ] User retention profile search works
- [ ] Charts render correctly
- [ ] Number formatting is correct
- [ ] Date formatting is correct
- [ ] Duration formatting is correct
- [ ] Error states handled gracefully
- [ ] Loading states work
- [ ] Responsive on mobile/tablet
- [ ] Auto-refresh works (if implemented)

---

## Technical Stack Recommendations

- **Framework**: React with Next.js or Vue with Nuxt
- **UI Library**: Tailwind CSS + shadcn/ui or similar
- **State Management**: React Query/TanStack Query or SWR
- **HTTP Client**: Axios
- **Charts**: Recharts, Chart.js, or Victory
- **Date Handling**: date-fns or dayjs
- **Number Formatting**: numeral.js or similar

---

## Key Features Summary

1. **Retention Overview**: Key metrics with date range filtering
2. **Cohort Analysis**: Analyze user retention by signup cohort
3. **User Retention Profile**: Individual user engagement metrics
4. **Data Visualization**: Charts and graphs for better insights
5. **Date Range Selection**: Flexible date filtering
6. **Real-time Updates**: Auto-refresh capabilities

---

**Note**: Analytics data helps understand user behavior and platform health. Ensure the dashboard is clear, accurate, and easy to interpret.


