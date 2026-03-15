import React, { useEffect, useState } from 'react';
import { getPlayerHistory } from '../api/http';

interface HistoryItem {
  quiz_id: string;
  quiz_title: string;
  quiz_code: string;
  score: number;
  rank: number;
  total_players: number;
  finished_at: string;
}

export default function QuizHistoryPage() {
  const [items, setItems] = useState<HistoryItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const limit = 10;

  useEffect(() => {
    getPlayerHistory(page, limit).then((res) => {
      const data = res.data;
      setItems(data.items || []);
      setHasMore((data.items || []).length === limit);
    });
  }, [page]);

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Quiz History</h1>

      {items.length === 0 && page === 1 ? (
        <p style={styles.empty}>You haven't played any quizzes yet.</p>
      ) : (
        <>
          <div style={styles.table}>
            <div style={styles.tableHeader}>
              <span style={{ flex: 2 }}>Quiz</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Code</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Score</span>
              <span style={{ flex: 1, textAlign: 'center' }}>Rank</span>
              <span style={{ flex: 1, textAlign: 'right' }}>Date</span>
            </div>
            {items.map((q, i) => (
              <div key={`${q.quiz_id}-${i}`} style={styles.tableRow}>
                <span style={{ flex: 2, color: '#1e293b', fontWeight: '500' }}>{q.quiz_title}</span>
                <span style={{ flex: 1, textAlign: 'center', color: '#6366f1', fontFamily: 'monospace' }}>
                  {q.quiz_code}
                </span>
                <span style={{ flex: 1, textAlign: 'center', fontWeight: 'bold', color: '#1e293b' }}>
                  {q.score}
                </span>
                <span style={{ flex: 1, textAlign: 'center', color: '#64748b' }}>
                  {q.rank}/{q.total_players}
                </span>
                <span style={{ flex: 1, textAlign: 'right', color: '#94a3b8', fontSize: '13px' }}>
                  {new Date(q.finished_at).toLocaleDateString()}
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
    alignItems: 'center', fontSize: '14px',
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
