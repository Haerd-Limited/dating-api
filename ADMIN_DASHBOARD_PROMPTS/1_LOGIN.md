# Admin Dashboard - Part 1: Login & Authentication

## Project Overview

Build a secure login system for an admin dashboard that manages a dating app backend. This is the first part of a multi-feature admin panel.

## API Base Configuration

- **Base URL**: `http://localhost:8080` (development) or your production URL
- **API Version**: `/api/v1`
- **Content-Type**: `application/json`
- **All responses**: JSON format

---

## Authentication System

### Authentication Method

**API Key Authentication via HTTP Header**
- **Header Name**: `X-Admin-Token`
- **Type**: Static API key (configured in backend environment)
- **Required for**: All `/api/v1/admin/*` endpoints

### Login Flow Implementation

1. **Login Page** (`/login`):
   - Clean, professional login form
   - Single input field: "Admin API Key" (password type for security)
   - Submit button
   - Error handling for invalid credentials
   - Loading state during authentication
   - Remember me option (optional)

2. **Authentication Logic**:
   - User enters admin API key
   - Validate by making a test API call to any admin endpoint (e.g., `GET /api/v1/admin/analytics/retention/global?date=2024-01-01`)
   - If successful (200 response): 
     - Store API key in `localStorage` with key `adminToken`
     - Redirect to dashboard home page
   - If failed (401/403): 
     - Show error message "Invalid API key. Please try again."
     - Clear any existing token
   - Handle network errors gracefully

3. **Token Storage**:
   - Store in `localStorage.setItem('adminToken', apiKey)`
   - Check for existing token on app load
   - Auto-redirect to dashboard if token exists and is valid
   - Clear token on logout

4. **HTTP Client Setup**:
   - Configure HTTP client (Axios/Fetch) to automatically include `X-Admin-Token` header
   - Intercept all requests to `/api/v1/admin/*` and add header
   - Handle 401/403 responses by clearing token and redirecting to login
   - Show loading indicators during API calls

5. **Logout**:
   - Clear `localStorage.removeItem('adminToken')`
   - Redirect to login page
   - Show confirmation message (optional)

6. **Token Validation on App Load**:
   - On app initialization, check if token exists in localStorage
   - If exists, make a test API call to validate it
   - If valid: proceed to dashboard
   - If invalid: clear token and show login page

### Example API Client Setup

```javascript
// Example with Axios
import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  }
});

// Add request interceptor
apiClient.interceptors.request.use((config) => {
  const adminToken = localStorage.getItem('adminToken');
  if (adminToken && config.url?.includes('/admin/')) {
    config.headers['X-Admin-Token'] = adminToken;
  }
  return config;
});

// Add response interceptor for auth errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 || error.response?.status === 403) {
      localStorage.removeItem('adminToken');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### Test Endpoint for Validation

**GET** `/api/v1/admin/analytics/retention/global`

**Query Parameters**:
- `date` (optional, default: today): Date in format `YYYY-MM-DD`

**Success Response** (200):
```json
{
  "total_users": 1000,
  "dau": 500,
  "wau": 2000,
  "mau": 8000,
  "average_time_to_first_return": "2h30m15s"
}
```

**Error Responses**:
- `401 Unauthorized`: Invalid or missing API key
- `403 Forbidden`: API key doesn't have access

---

## UI Requirements

### Login Page Design

**Layout**:
- Centered login form on the page
- Professional, clean design
- Brand/logo at the top (optional)
- "Admin Dashboard" title

**Form Elements**:
- **API Key Input**:
  - Type: password (masked input)
  - Placeholder: "Enter admin API key"
  - Required field
  - Auto-focus on page load
- **Submit Button**:
  - Primary action button
  - Shows loading spinner when authenticating
  - Disabled during authentication
- **Error Message**:
  - Red text below form
  - Only shown when authentication fails
  - Message: "Invalid API key. Please try again."

**User Experience**:
- Enter key and press Enter to submit
- Show loading state immediately on submit
- Clear error message when user starts typing again
- Responsive design (works on mobile/tablet)

### Design Requirements

1. **Color Scheme**:
   - Professional color palette
   - Primary button: Blue (#3B82F6 or similar)
   - Error text: Red (#EF4444)
   - Background: Light gray or white

2. **Typography**:
   - Clear, readable fonts
   - Appropriate heading sizes
   - Good contrast for accessibility

3. **Responsive Design**:
   - Works on desktop, tablet, and mobile
   - Form remains centered and usable on all screen sizes

4. **Loading States**:
   - Button shows spinner during authentication
   - Disable form inputs during loading
   - Prevent multiple submissions

5. **Error Handling**:
   - Display user-friendly error messages
   - Handle network errors gracefully
   - Show "Connection error. Please try again." for network failures

---

## Implementation Notes

1. **Security**:
   - Never log the API key to console
   - Clear token on any authentication error
   - Use HTTPS in production

2. **State Management**:
   - Store authentication state (isAuthenticated)
   - Use React Context or similar for global auth state
   - Persist token in localStorage

3. **Route Protection**:
   - Protect all dashboard routes
   - Redirect to login if not authenticated
   - Check token validity on route changes

4. **Error Handling**:
   - Handle all HTTP error codes
   - Show appropriate messages
   - Log errors for debugging (without sensitive data)

---

## Technical Stack Recommendations

- **Framework**: React with Next.js or Vue with Nuxt
- **UI Library**: Tailwind CSS + shadcn/ui or similar
- **State Management**: React Context or Zustand for auth state
- **HTTP Client**: Axios or Fetch API
- **Routing**: React Router or Next.js routing

---

## Testing Checklist

- [ ] Login with valid API key works
- [ ] Login with invalid API key shows error
- [ ] Token persists after page refresh
- [ ] Token validation on app load works
- [ ] Logout clears token and redirects
- [ ] Protected routes redirect to login when not authenticated
- [ ] 401/403 responses trigger logout
- [ ] Network errors handled gracefully
- [ ] Loading states work correctly
- [ ] Responsive on mobile/tablet

---

## Next Steps

After completing the login, you'll build:
- Part 2: Video Verification Review
- Part 3: Analytics Dashboard
- Part 4: Reports Management

Focus on making the login secure and user-friendly before moving to other features.


