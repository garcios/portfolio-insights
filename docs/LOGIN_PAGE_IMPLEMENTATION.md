# Login & Register Page Implementation

## Overview
A fully responsive Login and Register page has been implemented for the Portfolio Insights application.

## Components Created

### 1. Input Component (`src/components/ui/Input.tsx`)
A reusable input component that handles:
- Label rendering
- Error message display
- Focus/Blur states with styling
- All standard HTML input attributes

### 2. AuthPage (`src/pages/AuthPage.tsx`)
The main authentication page containing:
- **Tabs**: Toggle between Login and Register views.
- **Login Form**:
  - Email and Password fields
  - "Remember me" checkbox
  - "Forgot password?" link
  - Form validation
  - Mock API integration with loading state
  - Redirects to `/` (Overview) on success
- **Register Form**:
  - Name, Email, Password, Confirm Password fields
  - Password strength hint
  - Form validation (including password match)
  - Mock API integration with loading state
  - Switches to Login tab on success

## Integration
- Updated `src/App.tsx` to include the `/login` route.
- The page uses the existing design system variables (colors, shadows, etc.) to match the application's look and feel.

## Usage
Navigate to `/login` to view the authentication page.

## Future Improvements
- Integrate with real backend API endpoints (`/auth/login`, `/auth/register`).
- Implement actual token storage (localStorage/cookie).
- Add route protection (PrivateRoutes) to redirect unauthenticated users to `/login`.
