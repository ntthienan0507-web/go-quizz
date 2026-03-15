import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { getPlayerDashboard } from '../api/http';

interface Stats {
  total_quizzes: number;
  total_score: number;
  avg_score: number;
  best_rank: number;
  wins: number;
}

interface RecentQuiz {
  quiz_id: string;
  quiz_title: string;
  quiz_code: string;
  score: number;
  rank: number;
  total_players: number;
  finished_at: string;
}

interface Dashboard {
  stats: Stats;
  global_rank: number;
  recent: RecentQuiz[];
}

export default function PlayerDashboardPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [data, setData] = useState<Dashboard | null>(null);
  const [joinCode, setJoinCode] = useState('');

  useEffect(() => {
    getPlayerDashboard().then((res) => setData(res.data));
  }, []);

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    if (joinCode.trim()) navigate(`/play/${joinCode.trim()}`);
  };

  if (!data) return <div style={styles.loading}>Loading...</div>;

  const { stats, global_rank, recent } = data;

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Welcome, {user?.username}!</h1>

      <div style={styles.statsGrid}>
        <div style={styles.statCard}>
          <div style={styles.statValue}>{stats.total_quizzes}</div>
          <div style={styles.statLabel}>Quizzes Played</div>
        </div>
        <div style={styles.statCard}>
          <div style={styles.statValue}>{stats.total_score}</div>
          <div style={styles.statLabel}>Total Score</div>
        </div>
        <div style={styles.statCard}>
          <div style={styles.statValue}>{stats.wins}</div>
          <div style={styles.statLabel}>Wins</div>
        </div>
        <div style={styles.statCard}>
          <div style={styles.statValue}>{stats.best_rank || '-'}</div>
          <div style={styles.statLabel}>Best Rank</div>
        </div>
        <div style={{ ...styles.statCard, backgroundColor: '#6366f1' }}>
          <div style={{ ...styles.statValue, color: '#fff' }}>#{global_rank || '-'}</div>
          <div style={{ ...styles.statLabel, color: '#c7d2fe' }}>Global Rank</div>
        </div>
        <div style={styles.statCard}>
          <div style={styles.statValue}>{Math.round(stats.avg_score)}</div>
          <div style={styles.statLabel}>Avg Score</div>
        </div>
      </div>

      <form onSubmit={handleJoin} style={styles.joinForm}>
        <input
          style={styles.input}
          placeholder="Enter quiz code to join..."
          value={joinCode}
          onChange={(e) => setJoinCode(e.target.value)}
        />
        <button type="submit" style={styles.joinBtn}>Join Quiz</button>
      </form>

      <div style={styles.section}>
        <div style={styles.sectionHeader}>
          <h2 style={styles.sectionTitle}>Recent Quizzes</h2>
          <span style={styles.viewAll} onClick={() => navigate('/player/history')}>View All</span>
        </div>
        {recent.length === 0 ? (
          <p style={styles.empty}>No quizzes played yet. Join one above!</p>
        ) : (
          <div style={styles.table}>
            <div style={styles.tableHeader}>
              <span style={{ flex: 2 }}>Quiz</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Score</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Rank</span>
              <span style={{ flex: 1, textAlign: 'right' }}>Date</span>
            </div>
            {recent.map((q) => (
              <div key={q.quiz_id + q.finished_at} style={styles.tableRow}>
                <span style={{ flex: 2, color: '#1e293b', fontWeight: '500' }}>{q.quiz_title}</span>
                <span style={{ flex: 1, textAlign: 'center', color: '#6366f1', fontWeight: 'bold' }}>{q.score}</span>
                <span style={{ flex: 1, textAlign: 'center', color: '#64748b' }}>
                  {q.rank}/{q.total_players}
                </span>
                <span style={{ flex: 1, textAlign: 'right', color: '#94a3b8', fontSize: '13px' }}>
                  {new Date(q.finished_at).toLocaleDateString()}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: { padding: '24px', maxWidth: '900px', margin: '0 auto' },
  loading: { textAlign: 'center', padding: '60px', color: '#64748b' },
  heading: { color: '#1e293b', marginBottom: '24px' },
  statsGrid: {
    display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px', marginBottom: '24px',
  },
  statCard: {
    backgroundColor: '#fff', padding: '20px', borderRadius: '8px', textAlign: 'center',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  statValue: { fontSize: '28px', fontWeight: 'bold', color: '#1e293b' },
  statLabel: { fontSize: '13px', color: '#64748b', marginTop: '4px' },
  joinForm: { display: 'flex', gap: '8px', marginBottom: '32px' },
  input: {
    flex: 1, padding: '12px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#fff', color: '#1e293b', fontSize: '14px',
  },
  joinBtn: {
    padding: '12px 24px', backgroundColor: '#10b981', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '14px',
  },
  section: {
    backgroundColor: '#fff', borderRadius: '8px', padding: '20px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  sectionHeader: {
    display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px',
  },
  sectionTitle: { color: '#1e293b', margin: 0, fontSize: '18px' },
  viewAll: { color: '#6366f1', cursor: 'pointer', fontSize: '14px', fontWeight: '500' },
  empty: { color: '#64748b', textAlign: 'center', padding: '20px' },
  table: { display: 'flex', flexDirection: 'column' },
  tableHeader: {
    display: 'flex', padding: '8px 0', borderBottom: '2px solid #e2e8f0',
    color: '#64748b', fontSize: '12px', fontWeight: 'bold', textTransform: 'uppercase',
  },
  tableRow: {
    display: 'flex', padding: '12px 0', borderBottom: '1px solid #f1f5f9',
    alignItems: 'center', fontSize: '14px',
  },
};
