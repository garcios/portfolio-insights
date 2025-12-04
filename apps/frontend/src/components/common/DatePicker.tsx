import { Calendar } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { DayPicker } from 'react-day-picker';
import { format } from 'date-fns';
import 'react-day-picker/style.css';

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

    // Convert string value to Date object
    const selectedDate = value ? new Date(value) : undefined;

    // Convert min/max strings to Date objects
    const minDate = min ? new Date(min) : undefined;
    const maxDate = max ? new Date(max) : undefined;

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
        return format(date, 'MMM d, yyyy');
    };

    const handleDaySelect = (date: Date | undefined) => {
        if (date) {
            // Convert to YYYY-MM-DD format
            const year = date.getFullYear();
            const month = String(date.getMonth() + 1).padStart(2, '0');
            const day = String(date.getDate()).padStart(2, '0');
            onChange(`${year}-${month}-${day}`);
            setIsOpen(false);
        }
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
            <div
                onClick={() => setIsOpen(!isOpen)}
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
                    userSelect: 'none',
                    position: 'relative'
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

            {/* Calendar Popup */}
            {isOpen && (
                <div
                    style={{
                        position: 'absolute',
                        top: 'calc(100% + 8px)',
                        left: 0,
                        zIndex: 1000,
                        background: 'var(--color-bg-card)',
                        border: '1px solid var(--color-border)',
                        borderRadius: '12px',
                        boxShadow: 'var(--shadow-lg)',
                        padding: '16px',
                        minWidth: '300px'
                    }}
                >
                    <DayPicker
                        mode="single"
                        selected={selectedDate}
                        onSelect={handleDaySelect}
                        disabled={(date) => {
                            if (minDate && date < minDate) return true;
                            if (maxDate && date > maxDate) return true;
                            return false;
                        }}
                        modifiersStyles={{
                            selected: {
                                backgroundColor: 'var(--color-primary)',
                                color: 'white'
                            }
                        }}
                        styles={{
                            root: {
                                color: 'var(--color-text-primary)'
                            },
                            month_caption: {
                                color: 'var(--color-text-primary)',
                                fontWeight: '600',
                                fontSize: '0.95rem'
                            },
                            weekday: {
                                color: 'var(--color-text-secondary)',
                                fontSize: '0.875rem',
                                fontWeight: '500'
                            },
                            day: {
                                color: 'var(--color-text-primary)',
                                fontSize: '0.875rem',
                                borderRadius: '6px',
                                cursor: 'pointer'
                            },
                            day_button: {
                                borderRadius: '6px',
                                transition: 'all 0.2s'
                            },
                            nav_button: {
                                color: 'var(--color-text-secondary)',
                                cursor: 'pointer'
                            }
                        }}
                    />
                </div>
            )}
        </div>
    );
};

export default DatePicker;
