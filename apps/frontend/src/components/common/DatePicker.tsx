import { Calendar } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';

interface DatePickerProps {
    id?: string;
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    label?: string;
    min?: string;
    max?: string;
}

const DatePicker = ({ id, value, onChange, placeholder, label, min, max }: DatePickerProps) => {
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    // Close dropdown when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [isOpen]);

    const formatDisplayDate = (dateString: string) => {
        if (!dateString) return '';
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    };

    const handleInputClick = () => {
        inputRef.current?.showPicker?.();
    };

    return (
        <div ref={containerRef} style={{ position: 'relative', display: 'inline-block' }}>
            {label && (
                <label
                    htmlFor={id}
                    style={{
                        display: 'block',
                        fontSize: '0.875rem',
                        color: 'var(--color-text-secondary)',
                        fontWeight: '500',
                        marginBottom: '6px'
                    }}
                >
                    {label}
                </label>
            )}
            <div style={{ position: 'relative' }}>
                <div
                    onClick={handleInputClick}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        padding: '10px 12px',
                        borderRadius: '8px',
                        border: '1px solid var(--color-border)',
                        background: 'var(--color-bg-primary)',
                        color: value ? 'var(--color-text-primary)' : 'var(--color-text-tertiary)',
                        fontSize: '0.875rem',
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                        minWidth: '160px',
                        userSelect: 'none'
                    }}
                    onMouseEnter={(e) => {
                        e.currentTarget.style.borderColor = 'var(--color-primary)';
                        e.currentTarget.style.boxShadow = '0 0 0 3px rgba(99, 102, 241, 0.1)';
                    }}
                    onMouseLeave={(e) => {
                        e.currentTarget.style.borderColor = 'var(--color-border)';
                        e.currentTarget.style.boxShadow = 'none';
                    }}
                >
                    <Calendar size={16} style={{ color: 'var(--color-text-tertiary)', flexShrink: 0 }} />
                    <span style={{ flex: 1 }}>
                        {value ? formatDisplayDate(value) : (placeholder || 'Select date')}
                    </span>
                </div>
                <input
                    ref={inputRef}
                    id={id}
                    type="date"
                    value={value}
                    onChange={(e) => onChange(e.target.value)}
                    min={min}
                    max={max}
                    style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        height: '100%',
                        opacity: 0,
                        cursor: 'pointer',
                        pointerEvents: 'all'
                    }}
                />
            </div>
        </div>
    );
};

export default DatePicker;
