# Frontend Unit Tests

I have successfully created unit tests for the frontend components using Vitest and React Testing Library.

## Setup

1.  **Dependencies**: Installed `vitest`, `@vitest/coverage-v8`, `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`, `jsdom`.
2.  **Configuration**: Created `vitest.config.ts` and `src/test/setup.ts`.
3.  **Scripts**: Added `test` and `test:coverage` scripts to `package.json`.

## Tests Created

### 1. StatsCard (`src/test/StatsCard.test.tsx`)
-   Verifies rendering of title and value.
-   Checks positive and negative change indicators.
-   Ensures correct percentage formatting.
-   Tests optional props handling.

### 2. HoldingsTable (`src/test/HoldingsTable.test.tsx`)
-   Verifies table headers rendering.
-   Checks currency grouping logic.
-   Validates subtotal calculations.
-   Tests currency formatting (handling environment differences with flexible regex).

### 3. PortfolioChart (`src/test/PortfolioChart.test.tsx`)
-   Mocks `recharts` to avoid JSDOM issues with SVG/Canvas.
-   Verifies rendering of chart components (`AreaChart`, `XAxis`, `YAxis`, etc.).
-   Tests rendering with empty data.

## Running Tests

To run the tests:

```bash
cd apps/frontend
npm test
```

To run with coverage:

```bash
npm run test:coverage
```
