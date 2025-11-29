import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { Transaction, TransactionType } from '../../types/transaction';

interface AddTransactionModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSave: (transaction: Omit<Transaction, 'id'>) => void;
}

const AddTransactionModal = ({ isOpen, onClose, onSave }: AddTransactionModalProps) => {
    const [formData, setFormData] = useState({
        date: new Date().toISOString().split('T')[0],
        ticker: '',
        type: 'BUY' as TransactionType,
        quantity: 0,
        price: 0,
        brokerage: 0,
        currency: 'USD',
        notes: ''
    });

    const [total, setTotal] = useState(0);

    // Calculate total automatically
    useEffect(() => {
        const qty = Number(formData.quantity) || 0;
        const price = Number(formData.price) || 0;
        const brokerage = Number(formData.brokerage) || 0;

        // For BUY: (qty * price) + brokerage
        // For SELL: (qty * price) - brokerage
        // For SPLIT: 0 (usually just quantity adjustment)
        // For DIVIDEND: (qty * price) (price here is dividend per share)

        let calculatedTotal = 0;
        if (formData.type === 'BUY') {
            calculatedTotal = (qty * price) + brokerage;
        } else if (formData.type === 'SELL') {
            calculatedTotal = (qty * price) - brokerage;
        } else if (formData.type === 'DIVIDEND') {
            calculatedTotal = (qty * price);
        }

        setTotal(calculatedTotal);
    }, [formData.quantity, formData.price, formData.brokerage, formData.type]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        // Basic validation
        if (!formData.ticker || formData.quantity <= 0 || formData.price <= 0) {
            alert('Please fill in all required fields correctly.');
            return;
        }

        onSave({
            ...formData,
            quantity: Number(formData.quantity),
            price: Number(formData.price),
            brokerage: Number(formData.brokerage),
            total
        });

        // Reset form
        setFormData({
            date: new Date().toISOString().split('T')[0],
            ticker: '',
            type: 'BUY',
            quantity: 0,
            price: 0,
            brokerage: 0,
            currency: 'USD',
            notes: ''
        });
        onClose();
    };

    if (!isOpen) return null;

    return (
        <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'rgba(0, 0, 0, 0.5)',
            backdropFilter: 'blur(4px)',
            zIndex: 1000,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '16px'
        }}>
            <div style={{
                background: 'var(--color-bg-card)',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--color-border)',
                width: '100%',
                maxWidth: '500px',
                boxShadow: 'var(--shadow-xl)',
                animation: 'scaleIn 0.2s ease-out',
            }}>
                <div style={{
                    padding: '20px',
                    borderBottom: '1px solid var(--color-border)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                }}>
                    <h2 style={{
                        fontSize: '1.25rem',
                        fontWeight: '600',
                        color: 'var(--color-text-primary)'
                    }}>Add Transaction</h2>
                    <button
                        onClick={onClose}
                        style={{
                            background: 'transparent',
                            border: 'none',
                            color: 'var(--color-text-secondary)',
                            cursor: 'pointer',
                            padding: '4px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            borderRadius: '4px'
                        }}
                    >
                        <X size={20} />
                    </button>
                </div>

                <form onSubmit={handleSubmit} style={{ padding: '20px' }}>
                    <div style={{ display: 'grid', gap: '16px' }}>
                        {/* Ticker & Date */}
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Ticker Symbol *
                                </label>
                                <input
                                    type="text"
                                    name="ticker"
                                    value={formData.ticker}
                                    onChange={handleChange}
                                    placeholder="e.g. AAPL"
                                    required
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Trade Date *
                                </label>
                                <input
                                    type="date"
                                    name="date"
                                    value={formData.date}
                                    onChange={handleChange}
                                    required
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                />
                            </div>
                        </div>

                        {/* Type & Currency */}
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Type *
                                </label>
                                <select
                                    name="type"
                                    value={formData.type}
                                    onChange={handleChange}
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                >
                                    <option value="BUY">Buy</option>
                                    <option value="SELL">Sell</option>
                                    <option value="SPLIT">Split</option>
                                    <option value="DIVIDEND">Dividend</option>
                                </select>
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Currency
                                </label>
                                <select
                                    name="currency"
                                    value={formData.currency}
                                    onChange={handleChange}
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                >
                                    <option value="USD">USD</option>
                                    <option value="AUD">AUD</option>
                                    <option value="EUR">EUR</option>
                                    <option value="GBP">GBP</option>
                                </select>
                            </div>
                        </div>

                        {/* Quantity & Price */}
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Quantity *
                                </label>
                                <input
                                    type="number"
                                    name="quantity"
                                    value={formData.quantity}
                                    onChange={handleChange}
                                    min="0.000001"
                                    step="any"
                                    required
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Price per Share *
                                </label>
                                <input
                                    type="number"
                                    name="price"
                                    value={formData.price}
                                    onChange={handleChange}
                                    min="0"
                                    step="0.01"
                                    required
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                />
                            </div>
                        </div>

                        {/* Brokerage & Total */}
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Brokerage Fee
                                </label>
                                <input
                                    type="number"
                                    name="brokerage"
                                    value={formData.brokerage}
                                    onChange={handleChange}
                                    min="0"
                                    step="0.01"
                                    style={{
                                        padding: '8px 12px',
                                        borderRadius: '6px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem'
                                    }}
                                />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                    Total Amount
                                </label>
                                <div style={{
                                    padding: '8px 12px',
                                    borderRadius: '6px',
                                    background: 'var(--color-bg-tertiary)',
                                    color: 'var(--color-text-primary)',
                                    fontSize: '0.875rem',
                                    fontWeight: '600',
                                    border: '1px solid transparent'
                                }}>
                                    {new Intl.NumberFormat('en-US', { style: 'currency', currency: formData.currency }).format(total)}
                                </div>
                            </div>
                        </div>

                        {/* Notes */}
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                            <label style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-secondary)' }}>
                                Notes
                            </label>
                            <input
                                type="text"
                                name="notes"
                                value={formData.notes}
                                onChange={handleChange}
                                placeholder="Optional notes..."
                                style={{
                                    padding: '8px 12px',
                                    borderRadius: '6px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-primary)',
                                    color: 'var(--color-text-primary)',
                                    fontSize: '0.875rem'
                                }}
                            />
                        </div>
                    </div>

                    <div style={{
                        marginTop: '24px',
                        display: 'flex',
                        justifyContent: 'flex-end',
                        gap: '12px'
                    }}>
                        <button
                            type="button"
                            onClick={onClose}
                            style={{
                                padding: '8px 16px',
                                borderRadius: '6px',
                                border: '1px solid var(--color-border)',
                                background: 'transparent',
                                color: 'var(--color-text-primary)',
                                fontSize: '0.875rem',
                                fontWeight: '500',
                                cursor: 'pointer'
                            }}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            style={{
                                padding: '8px 16px',
                                borderRadius: '6px',
                                border: 'none',
                                background: 'var(--color-primary)',
                                color: 'white',
                                fontSize: '0.875rem',
                                fontWeight: '500',
                                cursor: 'pointer'
                            }}
                        >
                            Save Transaction
                        </button>
                    </div>
                </form>
            </div>
            <style>{`
                @keyframes scaleIn {
                    from {
                        opacity: 0;
                        transform: scale(0.95);
                    }
                    to {
                        opacity: 1;
                        transform: scale(1);
                    }
                }
            `}</style>
        </div>
    );
};

export default AddTransactionModal;
