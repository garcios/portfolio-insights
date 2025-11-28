# Header Component - Quick Start Guide

## 🚀 Quick Setup (5 minutes)

### 1. Import the Header Component

```tsx
import Header from './components/Header';
```

### 2. Add to Your App

```tsx
function App() {
  return (
    <div>
      <Header currentPage="overview" />
      {/* Your content here */}
    </div>
  );
}
```

### 3. That's it! 🎉

The header is now fully functional with:
- ✅ Desktop navigation
- ✅ Mobile hamburger menu
- ✅ User dropdown menu
- ✅ Active page highlighting

## 📋 Component Props

### Header Component

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `currentPage` | `'overview' \| 'transactions' \| 'fundamentals'` | `'overview'` | Highlights the active navigation link |

## 🎯 Common Use Cases

### Change Active Page

```tsx
// Overview page
<Header currentPage="overview" />

// Transactions page
<Header currentPage="transactions" />

// Fundamentals page
<Header currentPage="fundamentals" />
```

### Add New Navigation Link

**Step 1**: Edit `Header.tsx` - Add to `navLinks` array:

```tsx
const navLinks = [
  { label: 'Overview', href: '/', id: 'overview' as const },
  { label: 'Transactions', href: '/transactions', id: 'transactions' as const },
  { label: 'Asset Fundamentals', href: '/fundamentals', id: 'fundamentals' as const },
  { label: 'Analytics', href: '/analytics', id: 'analytics' as const }, // New link
];
```

**Step 2**: Update the `HeaderProps` type:

```tsx
interface HeaderProps {
    currentPage?: 'overview' | 'transactions' | 'fundamentals' | 'analytics';
}
```

### Customize User Menu

Edit `UserMenu.tsx` - Modify `menuItems` array:

```tsx
const menuItems = [
  { icon: Settings, label: 'Settings', onClick: handleSettings },
  { icon: Bell, label: 'Notifications', onClick: handleNotifications },
  { icon: User, label: 'Profile', onClick: handleProfile }, // New item
  { icon: LogOut, label: 'Sign Out', onClick: handleSignOut, isDanger: true },
];
```

## 🎨 Styling Customization

### Change Header Colors

Edit `index.css`:

```css
:root {
  --color-bg-secondary: #your-color;  /* Header background */
  --color-primary: #your-color;       /* Brand color */
  --color-secondary: #your-color;     /* Accent color */
}
```

### Adjust Mobile Breakpoint

Edit `index.css`:

```css
/* Change from 768px to your preferred breakpoint */
@media (max-width: 768px) {
  .desktop-nav { display: none; }
  .mobile-menu-toggle { display: flex !important; }
}
```

## 🔌 Router Integration

### React Router Example

**Step 1**: Install React Router (if not already installed):

```bash
npm install react-router-dom
```

**Step 2**: Update `NavLink.tsx`:

```tsx
import { Link } from 'react-router-dom';

// Replace <a> tag with:
<Link to={href} aria-current={isActive ? 'page' : undefined}>
  {children}
</Link>
```

**Step 3**: Update `Header.tsx` to auto-detect current page:

```tsx
import { useLocation } from 'react-router-dom';

function Header() {
  const location = useLocation();
  const currentPage = 
    location.pathname === '/' ? 'overview' 
    : location.pathname.includes('transactions') ? 'transactions'
    : location.pathname.includes('fundamentals') ? 'fundamentals'
    : 'overview';
  
  // ... rest of component
}
```

## 📱 Mobile Testing

### Test Responsive Behavior

1. **Desktop**: Open browser at full width
   - ✅ Navigation links visible
   - ✅ Hamburger menu hidden

2. **Mobile**: Resize to ≤ 768px
   - ✅ Navigation links hidden
   - ✅ Hamburger menu visible
   - ✅ Click hamburger to open menu
   - ✅ Menu slides in from right
   - ✅ Click outside to close

### Chrome DevTools

```
1. Open DevTools (F12)
2. Click "Toggle Device Toolbar" (Ctrl+Shift+M)
3. Select "iPhone 12 Pro" or similar
4. Test mobile menu functionality
```

## 🐛 Troubleshooting

### Issue: Mobile menu not showing

**Solution**: Check CSS is loaded
```tsx
// In main.tsx, ensure:
import './index.css';
```

### Issue: Navigation links not working

**Solution**: Integrate with router (see Router Integration above)

### Issue: Styling looks wrong

**Solution**: Verify CSS variables are defined in `index.css`

## 📚 Component Files

| File | Purpose |
|------|---------|
| `Header.tsx` | Main header component |
| `NavLink.tsx` | Individual navigation link |
| `MobileMenu.tsx` | Mobile slide-in menu |
| `UserMenu.tsx` | User dropdown menu |
| `HEADER_README.md` | Full documentation |

## 🎓 Learn More

- **Full Documentation**: See `HEADER_README.md` in components folder
- **Implementation Guide**: See `/docs/frontend-header-implementation.md`
- **Live Demo**: Run `npm run dev` and visit `http://localhost:5174`

## ✨ Tips & Best Practices

1. **Always set `currentPage`** to highlight active navigation
2. **Use semantic HTML** for better accessibility
3. **Test on real mobile devices** when possible
4. **Keep navigation items to 3-5** for best UX
5. **Ensure touch targets are ≥ 44px** for mobile

## 🎉 You're Ready!

The header component is production-ready and fully accessible. Customize it to match your brand and integrate with your routing solution for a complete navigation experience.

---

**Need Help?** Check the full documentation in `HEADER_README.md`
