# Portfolio Insights Frontend

A modern, beautiful portfolio tracking dashboard built with React, TypeScript, and Vite.

## 🎨 Features

### ✨ Modern Design
- **Dark Theme**: Sleek dark mode with vibrant gradient accents
- **Glassmorphism Effects**: Modern glass-effect cards with backdrop blur
- **Smooth Animations**: Fade-in effects, hover states, and micro-interactions
- **Responsive Layout**: Fully responsive grid system that adapts to all screen sizes

### 📊 Dashboard Components

#### Stats Cards
- **Total Portfolio Value**: Real-time portfolio valuation
- **Day Change**: Daily performance with trend indicators
- **Holdings Count**: Number of positions in portfolio
- **Portfolio Name**: Active portfolio identifier

Each card features:
- Gradient background decorations
- Custom icons with color coding
- Trend indicators (up/down arrows)
- Percentage change badges

#### Performance Chart
- **Interactive Area Chart**: 30-day portfolio value visualization
- **Gradient Fill**: Dynamic gradient based on performance (green for gains, red for losses)
- **Custom Tooltips**: Glassmorphic tooltips with formatted values
- **Time Period Selector**: 1D, 1W, 1M, 3M, 1Y, ALL options
- **Responsive**: Adapts to container size

#### Holdings Table
- **Detailed Position View**: Symbol, quantity, price, change, value, allocation
- **Gradient Avatars**: Color-coded symbol icons
- **Trend Indicators**: Real-time price change with visual indicators
- **Allocation Bars**: Visual representation of portfolio allocation
- **Hover Effects**: Interactive row highlighting

### 🎯 User Interface

#### Header
- **Branding**: Logo with gradient background
- **Real-time Status**: Portfolio tracking indicator
- **Action Buttons**:
  - Refresh button with loading animation
  - Notifications bell
  - Settings gear
  - User profile avatar

#### Interactions
- **Refresh Data**: Click refresh button to reload portfolio data
- **Hover Effects**: All interactive elements have smooth hover states
- **Loading States**: Spinner with loading message
- **Responsive**: Mobile-friendly navigation

## 🛠️ Technology Stack

- **React 18**: Modern React with hooks
- **TypeScript**: Type-safe development
- **Vite**: Lightning-fast build tool
- **Apollo Client**: GraphQL client for API communication
- **Recharts**: Beautiful, composable charts
- **Lucide React**: Modern icon library
- **Custom CSS**: Design system with CSS variables

## 🚀 Getting Started

### Prerequisites
- Node.js 24.4.0 or higher
- npm or yarn

### Installation

```bash
# Navigate to frontend directory
cd apps/frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The application will be available at `http://localhost:5173/`

### Build for Production

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## 📁 Project Structure

```
src/
├── components/          # React components
│   ├── HoldingsTable.tsx    # Holdings table component
│   ├── LoadingSpinner.tsx   # Loading state component
│   ├── PortfolioChart.tsx   # Performance chart component
│   └── StatsCard.tsx        # Statistics card component
├── types/              # TypeScript type definitions
│   └── portfolio.ts         # Portfolio data types
├── utils/              # Utility functions
│   └── apolloClient.ts      # Apollo GraphQL client setup
├── App.tsx             # Main application component
├── main.tsx            # Application entry point
└── index.css           # Global styles and design system
```

## 🎨 Design System

### Color Palette

The application uses a carefully crafted color palette defined in CSS variables:

```css
--color-bg-primary: #0a0e1a      /* Main background */
--color-bg-secondary: #111827    /* Secondary background */
--color-bg-card: #1e293b         /* Card background */

--color-primary: #6366f1         /* Primary brand color */
--color-secondary: #ec4899       /* Secondary accent */
--color-accent: #14b8a6          /* Tertiary accent */
--color-success: #10b981         /* Success/positive */
--color-danger: #ef4444          /* Danger/negative */
```

### Typography

- **Font Family**: Inter (Google Fonts)
- **Font Weights**: 300, 400, 500, 600, 700, 800
- **Responsive Sizing**: Scales appropriately across devices

### Spacing System

```css
--spacing-xs: 0.25rem    /* 4px */
--spacing-sm: 0.5rem     /* 8px */
--spacing-md: 1rem       /* 16px */
--spacing-lg: 1.5rem     /* 24px */
--spacing-xl: 2rem       /* 32px */
--spacing-2xl: 3rem      /* 48px */
```

### Border Radius

```css
--radius-sm: 0.375rem    /* Small elements */
--radius-md: 0.5rem      /* Medium elements */
--radius-lg: 0.75rem     /* Large elements */
--radius-xl: 1rem        /* Cards */
--radius-2xl: 1.5rem     /* Hero elements */
```

## 🔌 API Integration

### GraphQL Endpoint

The application connects to the GraphQL Gateway at:
- **Default**: `http://localhost:8080/query`
- **Configurable**: Set `VITE_GRAPHQL_URL` environment variable

### Environment Variables

Create a `.env` file in the frontend directory:

```env
VITE_GRAPHQL_URL=http://localhost:8080/query
```

### GraphQL Queries (Future Implementation)

Currently using mock data. To connect to real API, implement these queries:

```graphql
query GetPortfolio($id: ID!) {
  portfolio(id: $id) {
    id
    userId
    name
    holdings {
      symbol
      quantity
      value
    }
  }
}

query GetCurrentUser {
  me {
    id
    username
    email
  }
}
```

## 📊 Mock Data

The application currently uses generated mock data for demonstration:

- **7 Holdings**: AAPL, GOOGL, MSFT, AMZN, TSLA, NVDA, META
- **30 Days Performance**: Historical value data with upward trend
- **Random Variations**: Realistic price movements and changes

## 🎯 Future Enhancements

### Planned Features
- [ ] Real GraphQL API integration
- [ ] User authentication
- [ ] Multiple portfolio support
- [ ] Transaction history
- [ ] Portfolio analytics
- [ ] Export functionality (PDF, CSV)
- [ ] Dark/Light theme toggle
- [ ] Real-time price updates via WebSocket
- [ ] Advanced filtering and sorting
- [ ] Portfolio comparison
- [ ] Performance benchmarking
- [ ] Tax reporting

### Technical Improvements
- [ ] Error boundary implementation
- [ ] Comprehensive error handling
- [ ] Unit tests with Jest
- [ ] E2E tests with Playwright
- [ ] Accessibility improvements (WCAG 2.1)
- [ ] Performance optimization
- [ ] Code splitting
- [ ] PWA support

## 🐛 Known Issues

- Currently using mock data (GraphQL integration pending)
- Price changes are randomly generated
- No persistent state (data resets on refresh)

## 📝 License

This project is part of the Portfolio Insights monorepo.

## 🤝 Contributing

1. Follow the existing code style
2. Use TypeScript for type safety
3. Follow the component structure
4. Update documentation for new features
5. Test responsiveness on multiple devices

## 📞 Support

For issues or questions, please refer to the main project documentation.

---

**Built with ❤️ using React, TypeScript, and modern web technologies**
