import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const isAdmin = user?.role === 'admin';

  return (
    <nav style={styles.nav}>
      <Link to="/" style={styles.brand}>QuizApp</Link>
      <div style={styles.links}>
        {isAuthenticated ? (
          <>
            {isAdmin ? (
              <Link to="/dashboard" style={styles.link}>Dashboard</Link>
            ) : (
              <Link to="/player/dashboard" style={styles.link}>Home</Link>
            )}
            <Link to="/leaderboard" style={styles.link}>Leaderboard</Link>
            {!isAdmin && (
              <Link to="/player/history" style={styles.link}>History</Link>
            )}
            <Link to="/join" style={styles.link}>Join Quiz</Link>
            <Link to="/player/profile" style={styles.linkUser}>{user?.username}</Link>
            <button onClick={handleLogout} style={styles.btn}>Logout</button>
          </>
        ) : (
          <>
            <Link to="/login" style={styles.link}>Login</Link>
            <Link to="/join" style={styles.link}>Join Quiz</Link>
          </>
        )}
      </div>
    </nav>
  );
}

const styles: Record<string, React.CSSProperties> = {
  nav: {
    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
    padding: '12px 24px', backgroundColor: '#1e293b', color: '#fff',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  brand: { color: '#818cf8', fontSize: '20px', fontWeight: 'bold', textDecoration: 'none' },
  links: { display: 'flex', alignItems: 'center', gap: '16px' },
  link: { color: '#e2e8f0', textDecoration: 'none' },
  linkUser: { color: '#818cf8', textDecoration: 'none', fontSize: '14px', fontWeight: '500' },
  btn: {
    background: '#6366f1', color: '#fff', border: 'none', padding: '6px 14px',
    borderRadius: '6px', cursor: 'pointer', fontWeight: '500',
  },
};
