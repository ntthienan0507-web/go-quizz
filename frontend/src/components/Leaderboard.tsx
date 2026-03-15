import React from 'react';

interface LeaderboardEntry {
  user_id: string;
  username: string;
  score: number;
  rank: number;
}

interface Props {
  rankings: LeaderboardEntry[];
  title?: string;
}

export default function Leaderboard({ rankings, title = 'Leaderboard' }: Props) {
  return (
    <div style={styles.container}>
      <h3 style={styles.title}>{title}</h3>
      {rankings.length === 0 ? (
        <p style={styles.empty}>No scores yet</p>
      ) : (
        <div>
          {rankings.map((entry) => (
            <div key={entry.user_id} style={styles.row}>
              <span style={styles.rank}>#{entry.rank}</span>
              <span style={styles.name}>{entry.username}</span>
              <span style={styles.score}>{Math.round(entry.score)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    backgroundColor: '#ffffff', borderRadius: '8px', padding: '16px',
    minWidth: '250px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)',
  },
  title: { color: '#6366f1', margin: '0 0 12px 0', fontSize: '18px' },
  empty: { color: '#64748b', fontSize: '14px' },
  row: {
    display: 'flex', alignItems: 'center', padding: '8px 0',
    borderBottom: '1px solid #f1f5f9',
  },
  rank: { color: '#f59e0b', width: '40px', fontWeight: 'bold' },
  name: { color: '#1e293b', flex: 1 },
  score: { color: '#fff', backgroundColor: '#6366f1', padding: '2px 10px', borderRadius: '12px', fontWeight: 'bold' },
};
