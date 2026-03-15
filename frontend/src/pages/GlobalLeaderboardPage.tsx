import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { getGlobalLeaderboard } from '../api/http';

interface LeaderboardEntry {
  user_id: string;
  username: string;
  total_score: number;
  quizzes_played: number;
  avg_score: number;
  rank: number;
}

export default function GlobalLeaderboardPage() {
  const { user } = useAuth();
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const limit = 20;

  useEffect(() => {
    getGlobalLeaderboard(page, limit).then((res) => {
      const data = res.data;
      setEntries(data.items || []);
      setHasMore((data.items || []).length === limit);
    });
  }, [page]);

  const getRankStyle = (rank: number): React.CSSProperties => {
    if (rank === 1) return { color: '#f59e0b', fontWeight: 'bold', fontSize: '18px' };
    if (rank === 2) return { color: '#94a3b8', fontWeight: 'bold', fontSize: '16px' };
    if (rank === 3) return { color: '#cd7f32', fontWeight: 'bold', fontSize: '16px' };
    return { color: '#64748b' };
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Global Leaderboard</h1>

      {entries.length === 0 ? (
        <p style={styles.empty}>No results yet. Play some quizzes to get on the board!</p>
      ) : (
        <>
          <div style={styles.table}>
            <div style={styles.tableHeader}>
              <span style={{ width: '60px', textAlign: 'center' }}>Rank</span>
              <span style={{ flex: 2 }}>Player</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Total Score</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Quizzes</span>
              <span style={{ flex: 1, textAlign: 'right' }}>Avg Score</span>
            </div>
            {entries.map((e) => (
              <div
                key={e.user_id}
                style={{
                  ...styles.tableRow,
                  backgroundColor: user?.id === e.user_id ? '#eef2ff' : undefined,
                }}
              >
                <span style={{ width: '60px', textAlign: 'center', ...getRankStyle(e.rank) }}>
                  {e.rank <= 3 ? ['', '1st', '2nd', '3rd'][e.rank] : `#${e.rank}`}
                </span>
                <span style={{ flex: 2, color: '#1e293b', fontWeight: '500' }}>
                  {e.username}
                  {user?.id === e.user_id && <span style={styles.youBadge}>YOU</span>}
                </span>
                <span style={{ flex: 1, textAlign: 'center', fontWeight: 'bold', color: '#6366f1' }}>
                  {e.total_score}
                </span>
                <span style={{ flex: 1, textAlign: 'center', color: '#64748b' }}>
                  {e.quizzes_played}
                </span>
                <span style={{ flex: 1, textAlign: 'right', color: '#64748b' }}>
                  {Math.round(e.avg_score)}
                </span>
              </div>
            ))}
          </div>

          <div style={styles.pagination}>
            <button
              style={{ ...styles.pageBtn, opacity: page === 1 ? 0.5 : 1 }}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              Previous
            </button>
            <span style={styles.pageInfo}>Page {page}</span>
            <button
              style={{ ...styles.pageBtn, opacity: !hasMore ? 0.5 : 1 }}
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
            >
              Next
            </button>
          </div>
        </>
      )}
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: { padding: '24px', maxWidth: '900px', margin: '0 auto' },
  heading: { color: '#1e293b', marginBottom: '24px' },
  empty: { color: '#64748b', textAlign: 'center', padding: '40px' },
  table: {
    backgroundColor: '#fff', borderRadius: '8px', padding: '16px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  tableHeader: {
    display: 'flex', padding: '8px 0', borderBottom: '2px solid #e2e8f0',
    color: '#64748b', fontSize: '12px', fontWeight: 'bold', textTransform: 'uppercase',
  },
  tableRow: {
    display: 'flex', padding: '12px 0', borderBottom: '1px solid #f1f5f9',
    alignItems: 'center', fontSize: '14px', borderRadius: '4px',
  },
  youBadge: {
    marginLeft: '8px', backgroundColor: '#6366f1', color: '#fff',
    padding: '2px 6px', borderRadius: '4px', fontSize: '10px', fontWeight: 'bold',
  },
  pagination: {
    display: 'flex', justifyContent: 'center', alignItems: 'center',
    gap: '16px', marginTop: '20px',
  },
  pageBtn: {
    padding: '8px 16px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '500',
  },
  pageInfo: { color: '#64748b', fontSize: '14px' },
};
