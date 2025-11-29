import { InputHTMLAttributes, forwardRef } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    label: string;
    error?: string;
}

const Input = forwardRef<HTMLInputElement, InputProps>(
    ({ label, error, className = '', ...props }, ref) => {
        return (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', width: '100%' }}>
                <label
                    htmlFor={props.id}
                    style={{
                        fontSize: '0.875rem',
                        fontWeight: '500',
                        color: 'var(--color-text-secondary)'
                    }}
                >
                    {label}
                </label>
                <input
                    ref={ref}
                    className={className}
                    style={{
                        padding: '10px 12px',
                        borderRadius: '6px',
                        border: error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)',
                        background: 'var(--color-bg-primary)',
                        color: 'var(--color-text-primary)',
                        fontSize: '0.875rem',
                        outline: 'none',
                        transition: 'border-color 0.2s',
                        width: '100%',
                        boxSizing: 'border-box'
                    }}
                    onFocus={(e) => {
                        if (!error) {
                            e.currentTarget.style.borderColor = 'var(--color-primary)';
                        }
                    }}
                    onBlur={(e) => {
                        if (!error) {
                            e.currentTarget.style.borderColor = 'var(--color-border)';
                        }
                    }}
                    {...props}
                />
                {error && (
                    <span style={{ fontSize: '0.75rem', color: 'var(--color-danger)' }}>
                        {error}
                    </span>
                )}
            </div>
        );
    }
);

Input.displayName = 'Input';

export default Input;
