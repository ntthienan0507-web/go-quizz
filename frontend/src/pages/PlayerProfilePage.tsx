import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { getPlayerProfile, updatePlayerProfile } from '../api/http';

interface Stats {
  total_quizzes: number;
  total_score: number;
  avg_score: number;
  best_rank: number;
  wins: number;
}

interface ProfileUser {
  id: string;
  username: string;
  email: string;
  role: string;
  created_at: string;
}

export default function PlayerProfilePage() {
  const { user: authUser, logout } = useAuth();
  const [profileUser, setProfileUser] = useState<ProfileUser | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [editing, setEditing] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const fetchProfile = async () => {
    const res = await getPlayerProfile();
    setProfileUser(res.data.user);
    setStats(res.data.stats);
    setUsername(res.data.user.username);
    setEmail(res.data.user.email);
  };

  useEffect(() => { fetchProfile(); }, []);

  const handleSave = async () => {
    setError('');
    setSuccess('');
    try {
      await updatePlayerProfile({ username, email });
      setEditing(false);
      setSuccess('Profile updated!');
      fetchProfile();
      // Update localStorage so navbar reflects changes
      if (authUser) {
        const updated = { ...authUser, username, email };
        localStorage.setItem('user', JSON.stringify(updated));
      }
      setTimeout(() => setSuccess(''), 3000);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to update profile');
    }
  };

  if (!profileUser) return <div style={styles.loading}>Loading...</div>;

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Profile</h1>

      <div style={styles.card}>
        {error && <div style={styles.error}>{error}</div>}
        {success && <div style={styles.success}>{success}</div>}

        <div style={styles.field}>
          <label style={styles.label}>Username</label>
          {editing ? (
            <input
              style={styles.input}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          ) : (
            <div style={styles.value}>{profileUser.username}</div>
          )}
        </div>

        <div style={styles.field}>
          <label style={styles.label}>Email</label>
          {editing ? (
            <input
              style={styles.input}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          ) : (
            <div style={styles.value}>{profileUser.email}</div>
          )}
        </div>

        <div style={styles.field}>
          <label style={styles.label}>Role</label>
          <div style={styles.value}>
            <span style={styles.roleBadge}>{profileUser.role}</span>
          </div>
        </div>

        <div style={styles.field}>
          <label style={styles.label}>Member since</label>
          <div style={styles.value}>
            {new Date(profileUser.created_at).toLocaleDateString()}
          </div>
        </div>

        <div style={styles.actions}>
          {editing ? (
            <>
              <button style={styles.saveBtn} onClick={handleSave}>Save</button>
              <button style={styles.cancelBtn} onClick={() => { setEditing(false); setUsername(profileUser.username); setEmail(profileUser.email); }}>
                Cancel
              </button>
            </>
          ) : (
            <button style={styles.editBtn} onClick={() => setEditing(true)}>Edit Profile</button>
          )}
        </div>
      </div>

      {stats && (
        <div style={styles.card}>
          <h2 style={styles.cardTitle}>Stats</h2>
          <div style={styles.statsGrid}>
            <div style={styles.stat}>
              <div style={styles.statValue}>{stats.total_quizzes}</div>
              <div style={styles.statLabel}>Quizzes</div>
            </div>
            <div style={styles.stat}>
              <div style={styles.statValue}>{stats.total_score}</div>
              <div style={styles.statLabel}>Total Score</div>
            </div>
            <div style={styles.stat}>
              <div style={styles.statValue}>{Math.round(stats.avg_score)}</div>
              <div style={styles.statLabel}>Avg Score</div>
            </div>
            <div style={styles.stat}>
              <div style={styles.statValue}>{stats.wins}</div>
              <div style={styles.statLabel}>Wins</div>
            </div>
            <div style={styles.stat}>
              <div style={styles.statValue}>{stats.best_rank || '-'}</div>
              <div style={styles.statLabel}>Best Rank</div>
            </div>
          </div>
        </div>
      )}

      <div style={{ textAlign: 'center', marginTop: '24px' }}>
        <button style={styles.logoutBtn} onClick={logout}>Logout</button>
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: { padding: '24px', maxWidth: '600px', margin: '0 auto' },
  loading: { textAlign: 'center', padding: '60px', color: '#64748b' },
  heading: { color: '#1e293b', marginBottom: '24px' },
  card: {
    backgroundColor: '#fff', padding: '24px', borderRadius: '8px', marginBottom: '16px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  cardTitle: { color: '#1e293b', margin: '0 0 16px', fontSize: '18px' },
  field: { marginBottom: '16px' },
  label: { display: 'block', color: '#64748b', fontSize: '12px', fontWeight: 'bold', textTransform: 'uppercase', marginBottom: '4px' },
  value: { color: '#1e293b', fontSize: '16px' },
  input: {
    width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#f8fafc', color: '#1e293b', fontSize: '14px', boxSizing: 'border-box',
  },
  roleBadge: {
    backgroundColor: '#eef2ff', color: '#6366f1', padding: '4px 10px',
    borderRadius: '12px', fontSize: '12px', fontWeight: 'bold',
  },
  actions: { display: 'flex', gap: '8px', marginTop: '8px' },
  editBtn: {
    padding: '8px 20px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '500',
  },
  saveBtn: {
    padding: '8px 20px', backgroundColor: '#10b981', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '500',
  },
  cancelBtn: {
    padding: '8px 20px', backgroundColor: '#94a3b8', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '500',
  },
  logoutBtn: {
    padding: '8px 20px', backgroundColor: '#ef4444', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '500',
  },
  error: {
    backgroundColor: '#fef2f2', color: '#dc2626', padding: '8px 12px',
    borderRadius: '6px', marginBottom: '12px', fontSize: '14px', border: '1px solid #fecaca',
  },
  success: {
    backgroundColor: '#f0fdf4', color: '#16a34a', padding: '8px 12px',
    borderRadius: '6px', marginBottom: '12px', fontSize: '14px', border: '1px solid #bbf7d0',
  },
  statsGrid: { display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' },
  stat: { textAlign: 'center' },
  statValue: { fontSize: '24px', fontWeight: 'bold', color: '#1e293b' },
  statLabel: { fontSize: '12px', color: '#64748b', marginTop: '2px' },
};
