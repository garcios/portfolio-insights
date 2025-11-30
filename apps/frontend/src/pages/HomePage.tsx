import { useNavigate } from 'react-router-dom';
import {
    TrendingUp,
    Shield,
    BarChart3,
    Zap,
    ArrowRight,
    CheckCircle2,
    Sparkles,
    LineChart,
    PieChart,
    Activity
} from 'lucide-react';
import './HomePage.css';

const HomePage = () => {
    const navigate = useNavigate();

    const handleGetStarted = () => {
        navigate('/register');
    };

    const features = [
        {
            icon: <TrendingUp size={28} />,
            title: 'Real-Time Performance',
            description: 'Track your portfolio value with live market data and instant updates.'
        },
        {
            icon: <BarChart3 size={28} />,
            title: 'Advanced Analytics',
            description: 'Gain deep insights with comprehensive risk analysis and performance metrics.'
        },
        {
            icon: <LineChart size={28} />,
            title: 'Historical Tracking',
            description: 'Visualize your investment journey with detailed historical data and trends.'
        },
        {
            icon: <Shield size={28} />,
            title: 'Secure & Private',
            description: 'Bank-level security with OAuth 2.0 authentication and encrypted data storage.'
        },
        {
            icon: <PieChart size={28} />,
            title: 'Asset Allocation',
            description: 'Optimize your portfolio with intelligent asset allocation insights.'
        },
        {
            icon: <Activity size={28} />,
            title: 'Transaction Management',
            description: 'Effortlessly manage and track all your investment transactions in one place.'
        }
    ];

    const benefits = [
        'Real-time portfolio valuation',
        'Multi-currency support',
        'Comprehensive transaction history',
        'Performance analytics & insights',
        'Risk assessment tools',
        'Secure data encryption'
    ];

    return (
        <div className="homepage">
            {/* Hero Section */}
            <section className="hero-section">
                <div className="hero-background">
                    <div className="gradient-orb orb-1"></div>
                    <div className="gradient-orb orb-2"></div>
                    <div className="gradient-orb orb-3"></div>
                </div>

                <div className="hero-content">
                    <div className="hero-badge">
                        <Sparkles size={16} />
                        <span>Smarter Investing. Clearer Decisions.</span>
                    </div>

                    <h1 className="hero-title">
                        Portfolio Insights
                    </h1>

                    <p className="hero-subtitle">
                        Transform the way you manage your investments with powerful analytics,
                        real-time tracking, and actionable insights—all in one beautiful platform.
                    </p>

                    <div className="hero-cta">
                        <button
                            className="btn-primary-large"
                            onClick={handleGetStarted}
                        >
                            Get Started Free
                            <ArrowRight size={20} />
                        </button>

                        <button
                            className="btn-secondary-large"
                            onClick={() => navigate('/login')}
                        >
                            Sign In
                        </button>
                    </div>

                    <div className="hero-stats">
                        <div className="stat-item">
                            <div className="stat-value">
                                <Zap size={20} className="stat-icon" />
                                Real-Time
                            </div>
                            <div className="stat-label">Market Data</div>
                        </div>
                        <div className="stat-divider"></div>
                        <div className="stat-item">
                            <div className="stat-value">
                                <Shield size={20} className="stat-icon" />
                                Bank-Level
                            </div>
                            <div className="stat-label">Security</div>
                        </div>
                        <div className="stat-divider"></div>
                        <div className="stat-item">
                            <div className="stat-value">
                                <BarChart3 size={20} className="stat-icon" />
                                Advanced
                            </div>
                            <div className="stat-label">Analytics</div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Features Section */}
            <section className="features-section">
                <div className="section-header">
                    <h2 className="section-title">Everything You Need to Succeed</h2>
                    <p className="section-subtitle">
                        Powerful features designed to help you make informed investment decisions
                    </p>
                </div>

                <div className="features-grid">
                    {features.map((feature, index) => (
                        <div key={index} className="feature-card">
                            <div className="feature-icon">
                                {feature.icon}
                            </div>
                            <h3 className="feature-title">{feature.title}</h3>
                            <p className="feature-description">{feature.description}</p>
                        </div>
                    ))}
                </div>
            </section>

            {/* Benefits Section */}
            <section className="benefits-section">
                <div className="benefits-container">
                    <div className="benefits-content">
                        <h2 className="benefits-title">
                            Why Choose Portfolio Insights?
                        </h2>
                        <p className="benefits-subtitle">
                            Join thousands of investors who trust Portfolio Insights to manage
                            and grow their wealth with confidence.
                        </p>

                        <div className="benefits-list">
                            {benefits.map((benefit, index) => (
                                <div key={index} className="benefit-item">
                                    <CheckCircle2 size={20} className="benefit-check" />
                                    <span>{benefit}</span>
                                </div>
                            ))}
                        </div>

                        <button
                            className="btn-primary-large"
                            onClick={handleGetStarted}
                        >
                            Start Your Journey
                            <ArrowRight size={20} />
                        </button>
                    </div>

                    <div className="benefits-visual">
                        <div className="visual-card card-1">
                            <div className="card-header">
                                <TrendingUp size={24} />
                                <span>Portfolio Value</span>
                            </div>
                            <div className="card-value">$124,567.89</div>
                            <div className="card-change positive">
                                <span>+12.5%</span>
                                <span className="change-label">This month</span>
                            </div>
                        </div>

                        <div className="visual-card card-2">
                            <div className="card-header">
                                <Activity size={24} />
                                <span>Performance</span>
                            </div>
                            <div className="mini-chart">
                                <div className="chart-bar" style={{ height: '40%' }}></div>
                                <div className="chart-bar" style={{ height: '65%' }}></div>
                                <div className="chart-bar" style={{ height: '45%' }}></div>
                                <div className="chart-bar" style={{ height: '80%' }}></div>
                                <div className="chart-bar" style={{ height: '70%' }}></div>
                                <div className="chart-bar" style={{ height: '90%' }}></div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* CTA Section */}
            <section className="cta-section">
                <div className="cta-content">
                    <h2 className="cta-title">Ready to Take Control?</h2>
                    <p className="cta-subtitle">
                        Join Portfolio Insights today and start making smarter investment decisions.
                    </p>
                    <button
                        className="btn-cta"
                        onClick={handleGetStarted}
                    >
                        Get Started Now
                        <ArrowRight size={24} />
                    </button>
                </div>
            </section>

            {/* Footer */}
            <footer className="homepage-footer">
                <div className="footer-content">
                    <div className="footer-brand">
                        <h3>Portfolio Insights</h3>
                        <p>Smarter investing for everyone.</p>
                    </div>
                    <div className="footer-links">
                        <div className="footer-column">
                            <h4>Product</h4>
                            <a href="#features">Features</a>
                            <a href="#pricing">Pricing</a>
                            <a href="#security">Security</a>
                        </div>
                        <div className="footer-column">
                            <h4>Company</h4>
                            <a href="#about">About</a>
                            <a href="#blog">Blog</a>
                            <a href="#careers">Careers</a>
                        </div>
                        <div className="footer-column">
                            <h4>Support</h4>
                            <a href="#help">Help Center</a>
                            <a href="#contact">Contact</a>
                            <a href="#status">Status</a>
                        </div>
                    </div>
                </div>
                <div className="footer-bottom">
                    <p>© 2025 Portfolio Insights. All rights reserved.</p>
                    <div className="footer-legal">
                        <a href="#privacy">Privacy Policy</a>
                        <a href="#terms">Terms of Service</a>
                    </div>
                </div>
            </footer>
        </div>
    );
};

export default HomePage;
