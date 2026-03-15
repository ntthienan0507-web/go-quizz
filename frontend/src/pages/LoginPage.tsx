import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function LoginPage() {
  const [isRegister, setIsRegister] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login, register } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      if (isRegister) {
        await register(username, email, password);
        await login(email, password);
      } else {
        await login(email, password);
      }
      const savedUser = JSON.parse(localStorage.getItem('user') || '{}');
      navigate(savedUser.role === 'admin' ? '/dashboard' : '/player/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Something went wrong');
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h2 style={styles.title}>{isRegister ? 'Register' : 'Login'}</h2>
        {error && <div style={styles.error}>{error}</div>}
        <form onSubmit={handleSubmit}>
          {isRegister && (
            <>
              <input
                style={styles.input} placeholder="Username" value={username}
                onChange={(e) => setUsername(e.target.value)} required
              />
            </>
          )}
          <input
            style={styles.input} type="email" placeholder="Email" value={email}
            onChange={(e) => setEmail(e.target.value)} required
          />
          <input
            style={styles.input} type="password" placeholder="Password" value={password}
            onChange={(e) => setPassword(e.target.value)} required minLength={6}
          />
          <button type="submit" style={styles.btn}>
            {isRegister ? 'Register' : 'Login'}
          </button>
        </form>
        <p style={styles.toggle}>
          {isRegister ? 'Already have an account?' : "Don't have an account?"}{' '}
          <span style={styles.link} onClick={() => setIsRegister(!isRegister)}>
            {isRegister ? 'Login' : 'Register'}
          </span>
        </p>
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex', justifyContent: 'center', alignItems: 'center',
    minHeight: 'calc(100vh - 60px)',
  },
  card: {
    backgroundColor: '#ffffff', padding: '40px', borderRadius: '12px',
    width: '100%', maxWidth: '400px',
    boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -2px rgba(0,0,0,0.1)',
  },
  title: { color: '#1e293b', textAlign: 'center', marginBottom: '24px' },
  input: {
    width: '100%', padding: '12px', marginBottom: '12px', borderRadius: '6px',
    border: '1px solid #e2e8f0', backgroundColor: '#f8fafc', color: '#1e293b',
    fontSize: '14px', boxSizing: 'border-box',
  },
  btn: {
    width: '100%', padding: '12px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', fontSize: '16px', cursor: 'pointer',
    fontWeight: 'bold',
  },
  error: {
    backgroundColor: '#fef2f2', color: '#dc2626', padding: '8px 12px',
    borderRadius: '6px', marginBottom: '12px', fontSize: '14px',
    border: '1px solid #fecaca',
  },
  toggle: { color: '#64748b', textAlign: 'center', marginTop: '16px', fontSize: '14px' },
  link: { color: '#6366f1', cursor: 'pointer', fontWeight: '500' },
};
