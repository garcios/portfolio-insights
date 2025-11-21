import React from 'react';

const LoadingSpinner: React.FC = () => {
    return (
        <div style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '400px',
            gap: '16px',
        }}>
            <div className="spinner" />
            <p style={{
                color: 'var(--color-text-tertiary)',
                fontSize: '0.875rem',
            }}>
                Loading portfolio data...
            </p>
        </div>
    );
};

export default LoadingSpinner;
