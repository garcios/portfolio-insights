# Header Component Documentation

## Overview

The Header component is a responsive navigation bar for the Portfolio Insights application. It includes:
- Logo and branding
- Desktop navigation menu
- Mobile hamburger menu
- User menu with dropdown
- Full accessibility support (ARIA attributes, keyboard navigation)

## Components

### 1. **Header.tsx**
Main header component that orchestrates all navigation elements.

**Props:**
- `currentPage?: 'overview' | 'transactions' | 'fundamentals'` - Highlights the active page

**Usage:**
```tsx
import Header from './components/Header';

function App() {
  return (
    <div>
      <Header currentPage="overview" />
      {/* Your page content */}
    </div>
  );
}
```

### 2. **NavLink.tsx**
Individual navigation link component with active state styling.

**Props:**
- `href: string` - Link destination
- `isActive?: boolean` - Whether this link is currently active
- `children: ReactNode` - Link text/content
- `onClick?: () => void` - Optional click handler

**Features:**
- Active state with gradient underline
- Smooth hover transitions
- Accessible with `aria-current` attribute

### 3. **MobileMenu.tsx**
Slide-in mobile navigation menu.

**Props:**
- `isOpen: boolean` - Controls menu visibility
- `onClose: () => void` - Callback when menu should close
- `navLinks: Array<{label, href, id}>` - Navigation items
- `currentPage: string` - Current active page

**Features:**
- Slide-in animation from right
- Backdrop overlay with blur effect
- Closes on:
  - Backdrop click
  - Escape key press
  - Link selection
- Prevents body scroll when open

### 4. **UserMenu.tsx**
User avatar with dropdown menu.

**Features:**
- User info display (name, email)
- Settings link
- Notifications link
- Theme toggle (Light/Dark mode)
- Sign out option
- Click-outside to close
- Keyboard navigation (Escape to close)

## Responsive Behavior

### Desktop (> 768px)
- Full navigation menu visible in header
- Hamburger menu hidden
- User menu always visible

### Mobile (≤ 768px)
- Navigation menu hidden
- Hamburger menu button visible
- Slide-in menu on hamburger click
- User menu always visible

## Accessibility Features

### ARIA Attributes
- `aria-label` on buttons and navigation
- `aria-expanded` on menu toggles
- `aria-current="page"` on active links
- `aria-haspopup` on dropdown triggers
- `role="menu"` and `role="menuitem"` on dropdowns

### Keyboard Navigation
- **Escape**: Closes mobile menu and user dropdown
- **Tab**: Navigate through interactive elements
- **Enter/Space**: Activate buttons and links

### Screen Reader Support
- Semantic HTML (`<nav>`, `<header>`, `<button>`)
- Descriptive labels on all interactive elements
- Proper heading hierarchy

## Styling

All components use the existing CSS variable system from `index.css`:

### Colors Used
- `--color-bg-secondary`: Header background
- `--color-bg-tertiary`: Button backgrounds
- `--color-bg-hover`: Hover states
- `--color-border`: Borders
- `--color-text-primary`: Primary text
- `--color-text-secondary`: Secondary text
- `--color-primary`: Brand color
- `--color-secondary`: Accent color

### Animations
- Fade in/out for overlays
- Slide-in from right for mobile menu
- Smooth hover transitions
- Gradient underline for active links

## Customization

### Adding New Navigation Links

Edit the `navLinks` array in `Header.tsx`:

```tsx
const navLinks = [
  { label: 'Overview', href: '/', id: 'overview' as const },
  { label: 'Transactions', href: '/transactions', id: 'transactions' as const },
  { label: 'Asset Fundamentals', href: '/fundamentals', id: 'fundamentals' as const },
  // Add your new link here
  { label: 'Analytics', href: '/analytics', id: 'analytics' as const },
];
```

Don't forget to update the `currentPage` type in the `HeaderProps` interface.

### Customizing User Menu Items

Edit the `menuItems` array in `UserMenu.tsx`:

```tsx
const menuItems = [
  { icon: Settings, label: 'Settings', onClick: () => console.log('Settings') },
  { icon: Bell, label: 'Notifications', onClick: () => console.log('Notifications') },
  // Add your custom items here
];
```

### Changing Theme

The header automatically uses your CSS variables. To customize:

1. Update colors in `index.css`:
```css
:root {
  --color-primary: #your-color;
  --color-secondary: #your-color;
}
```

2. The header will automatically reflect these changes.

## Integration with Routing

Currently, the navigation uses `preventDefault()` on link clicks. To integrate with a router:

### React Router Example

```tsx
import { Link, useLocation } from 'react-router-dom';

// In NavLink.tsx, replace the <a> tag with:
<Link
  to={href}
  aria-current={isActive ? 'page' : undefined}
  // ... rest of props
>
  {children}
</Link>

// In Header.tsx, determine currentPage from route:
import { useLocation } from 'react-router-dom';

function Header() {
  const location = useLocation();
  const currentPage = location.pathname === '/' ? 'overview' 
    : location.pathname.includes('transactions') ? 'transactions'
    : location.pathname.includes('fundamentals') ? 'fundamentals'
    : 'overview';
  
  // ... rest of component
}
```

## Performance Considerations

- Mobile menu uses CSS animations (GPU-accelerated)
- Event listeners are properly cleaned up in `useEffect`
- Body scroll is restored when mobile menu closes
- Minimal re-renders with proper state management

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Mobile browsers (iOS Safari, Chrome Mobile)
- Requires CSS Grid and Flexbox support
- Uses CSS custom properties (CSS variables)

## Future Enhancements

Potential improvements:
- [ ] Add search bar in header
- [ ] Breadcrumb navigation
- [ ] Notification badge on bell icon
- [ ] User profile picture upload
- [ ] Multi-level dropdown menus
- [ ] Sticky header with scroll behavior
- [ ] Header transparency on scroll

## Troubleshooting

### Mobile menu not appearing
- Check that viewport width is ≤ 768px
- Verify CSS is loaded correctly
- Check browser console for errors

### Links not working
- Ensure you've integrated with your routing solution
- Check that `onClick` handlers are properly defined

### Styling issues
- Verify all CSS variables are defined in `index.css`
- Check for CSS specificity conflicts
- Ensure `index.css` is imported in `main.tsx`

## Example Implementation

See `App.tsx` for a complete working example of the Header component integrated into the application.
