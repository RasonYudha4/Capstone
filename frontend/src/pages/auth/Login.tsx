import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router'
import { useAuth } from '../../context/AuthContext'
import type { Role } from '../../context/types'
import { findUser } from '../../data/users'

const redirectMap: Record<Role, string> = {
    1: '/master-admin',
    2: '/admin',
    3: '/user',
}

export default function Login() {
    const { login } = useAuth()
    const navigate = useNavigate()
    const location = useLocation()

    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    const from = (location.state as { from?: Location })?.from?.pathname

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError('')
        setLoading(true)

        // simulate a small network delay
        await new Promise((r) => setTimeout(r, 600))

        const user = findUser(email, password)

        if (!user) {
            setError('Invalid email or password.')
            setLoading(false)
            return
        }

        login(user)
        navigate(from ?? redirectMap[user.role], { replace: true })
    }

    return (
        <div style={styles.page}>
            <div style={styles.card}>

                <div style={styles.header}>
                    <div style={styles.dot} />
                    <h1 style={styles.title}>Sign in</h1>
                    <p style={styles.subtitle}>Capstone project demo</p>
                </div>

                <form onSubmit={handleSubmit} style={styles.form}>
                    <div style={styles.field}>
                        <label style={styles.label}>Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="you@example.com"
                            required
                            style={styles.input}
                            onFocus={(e) => Object.assign(e.target.style, styles.inputFocus)}
                            onBlur={(e) => Object.assign(e.target.style, styles.inputBlur)}
                        />
                    </div>

                    <div style={styles.field}>
                        <label style={styles.label}>Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="••••••••"
                            required
                            style={styles.input}
                            onFocus={(e) => Object.assign(e.target.style, styles.inputFocus)}
                            onBlur={(e) => Object.assign(e.target.style, styles.inputBlur)}
                        />
                    </div>

                    {error && <p style={styles.error}>{error}</p>}

                    <button
                        type="submit"
                        disabled={loading}
                        style={{
                            ...styles.button,
                            ...(loading ? styles.buttonDisabled : {}),
                        }}
                    >
                        {loading ? 'Signing in…' : 'Sign in'}
                    </button>
                </form>

                <div style={styles.hints}>
                    <p style={styles.hintsTitle}>Demo accounts</p>
                    {[
                        { role: 'Admin', email: 'admin@example.com', password: 'admin123' },
                        { role: 'Master admin', email: 'master@example.com', password: 'master123' },
                        { role: 'User', email: 'user@example.com', password: 'user123' },
                    ].map((a) => (
                        <button
                            key={a.role}
                            type="button"
                            style={styles.hintRow}
                            onClick={() => { setEmail(a.email); setPassword(a.password) }}
                        >
                            <span style={styles.hintRole}>{a.role}</span>
                            <span style={styles.hintEmail}>{a.email}</span>
                        </button>
                    ))}
                </div>

            </div>
        </div>
    )
}

// ── styles ────────────────────────────────────────────────────────────────────

const styles: Record<string, React.CSSProperties> = {
    page: {
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#0f0f0f',
        fontFamily: "'DM Sans', sans-serif",
        padding: '24px',
    },
    card: {
        width: '100%',
        maxWidth: '400px',
        backgroundColor: '#1a1a1a',
        border: '1px solid #2a2a2a',
        borderRadius: '16px',
        padding: '40px 36px',
        display: 'flex',
        flexDirection: 'column',
        gap: '28px',
    },
    header: {
        display: 'flex',
        flexDirection: 'column',
        gap: '6px',
    },
    dot: {
        width: '10px',
        height: '10px',
        borderRadius: '50%',
        backgroundColor: '#4ade80',
        marginBottom: '12px',
    },
    title: {
        margin: 0,
        fontSize: '24px',
        fontWeight: 600,
        color: '#f5f5f5',
        letterSpacing: '-0.3px',
    },
    subtitle: {
        margin: 0,
        fontSize: '13px',
        color: '#666',
    },
    form: {
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
    },
    field: {
        display: 'flex',
        flexDirection: 'column',
        gap: '6px',
    },
    label: {
        fontSize: '12px',
        fontWeight: 500,
        color: '#999',
        textTransform: 'uppercase',
        letterSpacing: '0.5px',
    },
    input: {
        backgroundColor: '#111',
        border: '1px solid #2a2a2a',
        borderRadius: '8px',
        padding: '10px 14px',
        fontSize: '14px',
        color: '#f5f5f5',
        outline: 'none',
        transition: 'border-color 0.15s',
    },
    inputFocus: {
        borderColor: '#4ade80',
    },
    inputBlur: {
        borderColor: '#2a2a2a',
    },
    error: {
        margin: 0,
        fontSize: '13px',
        color: '#f87171',
        backgroundColor: '#1f1010',
        border: '1px solid #3a1a1a',
        borderRadius: '6px',
        padding: '8px 12px',
    },
    button: {
        marginTop: '4px',
        padding: '11px',
        borderRadius: '8px',
        border: 'none',
        backgroundColor: '#4ade80',
        color: '#0f0f0f',
        fontSize: '14px',
        fontWeight: 600,
        cursor: 'pointer',
        transition: 'opacity 0.15s',
    },
    buttonDisabled: {
        opacity: 0.5,
        cursor: 'not-allowed',
    },
    hints: {
        borderTop: '1px solid #2a2a2a',
        paddingTop: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '6px',
    },
    hintsTitle: {
        margin: '0 0 6px',
        fontSize: '11px',
        color: '#555',
        textTransform: 'uppercase',
        letterSpacing: '0.5px',
    },
    hintRow: {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '8px 10px',
        borderRadius: '6px',
        border: '1px solid #222',
        backgroundColor: 'transparent',
        cursor: 'pointer',
        transition: 'background-color 0.1s',
        textAlign: 'left',
    },
    hintRole: {
        fontSize: '12px',
        color: '#4ade80',
        fontWeight: 500,
    },
    hintEmail: {
        fontSize: '12px',
        color: '#555',
        fontFamily: 'monospace',
    },
}