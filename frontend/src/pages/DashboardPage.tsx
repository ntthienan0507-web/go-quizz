import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listQuizzes, createQuiz, deleteQuiz, startQuiz, finishQuiz } from '../api/http';

interface Quiz {
  id: string;
  title: string;
  quiz_code: string;
  status: string;
  mode: string;
  time_per_question: number;
  created_at: string;
}

export default function DashboardPage() {
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [title, setTitle] = useState('');
  const [mode, setMode] = useState('live');
  const [toast, setToast] = useState('');
  const [tpq, setTpq] = useState(30);
  const navigate = useNavigate();

  const fetchQuizzes = async () => {
    const res = await listQuizzes();
    setQuizzes(res.data.items || []);
  };

  useEffect(() => { fetchQuizzes(); }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    await createQuiz(title, tpq, mode);
    setTitle('');
    setMode('live');
    fetchQuizzes();
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Delete this quiz?')) return;
    await deleteQuiz(id);
    fetchQuizzes();
  };

  const handleStart = async (id: string) => {
    await startQuiz(id);
    fetchQuizzes();
  };

  const handleFinish = async (id: string) => {
    await finishQuiz(id);
    fetchQuizzes();
  };

  const handleShare = (code: string) => {
    const link = `${window.location.origin}/join?code=${code}`;
    navigator.clipboard.writeText(link).then(() => {
      setToast('Link copied to clipboard!');
      setTimeout(() => setToast(''), 3000);
    });
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>My Quizzes</h1>

      <form onSubmit={handleCreate} style={styles.form}>
        <input
          style={styles.input} placeholder="Quiz title" value={title}
          onChange={(e) => setTitle(e.target.value)} required
        />
        <select
          style={{ ...styles.input, flex: 'none', width: '130px' }}
          value={mode} onChange={(e) => setMode(e.target.value)}
        >
          <option value="live">Live</option>
          <option value="self_paced">Self-paced</option>
        </select>
        <input
          style={{ ...styles.input, flex: 'none', width: '100px' }} type="number" placeholder="Seconds"
          value={tpq} onChange={(e) => setTpq(Number(e.target.value))} min={5} max={120}
        />
        <button type="submit" style={styles.btn}>Create</button>
      </form>

      <div style={styles.list}>
        {quizzes.map((q) => (
          <div key={q.id} style={styles.card}>
            <div style={styles.cardHeader}>
              <h3 style={styles.cardTitle}>{q.title}</h3>
              <span style={{
                ...styles.badge,
                backgroundColor: q.status === 'active' ? '#10b981' : q.status === 'finished' ? '#94a3b8' : '#f59e0b',
              }}>
                {q.status}
              </span>
            </div>
            <p style={styles.code}>Code: <strong>{q.quiz_code}</strong></p>
            <p style={styles.meta}>
              {q.time_per_question}s per question
              {' · '}
              <span style={{ color: q.mode === 'self_paced' ? '#8b5cf6' : '#6366f1' }}>
                {q.mode === 'self_paced' ? 'Self-paced' : 'Live'}
              </span>
            </p>
            <div style={styles.actions}>
              <button style={styles.actionBtn} onClick={() => navigate(`/quiz/${q.id}/questions`)}>
                Questions
              </button>
              {q.status === 'draft' && (
                <button style={{ ...styles.actionBtn, backgroundColor: '#10b981' }} onClick={() => handleStart(q.id)}>
                  Start
                </button>
              )}
              {q.status === 'active' && (
                <button style={{ ...styles.actionBtn, backgroundColor: '#f59e0b' }} onClick={() => handleFinish(q.id)}>
                  Finish
                </button>
              )}
              <button style={{ ...styles.actionBtn, backgroundColor: '#0ea5e9' }} onClick={() => handleShare(q.quiz_code)}>
                Share
              </button>
              <button style={{ ...styles.actionBtn, backgroundColor: '#ef4444' }} onClick={() => handleDelete(q.id)}>
                Delete
              </button>
            </div>
          </div>
        ))}
        {quizzes.length === 0 && <p style={styles.empty}>No quizzes yet. Create one above!</p>}
      </div>
      {toast && <div style={styles.toast}>{toast}</div>}
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: { padding: '24px', maxWidth: '800px', margin: '0 auto' },
  heading: { color: '#1e293b', marginBottom: '24px' },
  form: { display: 'flex', gap: '8px', marginBottom: '24px' },
  input: {
    padding: '10px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#fff', color: '#1e293b', fontSize: '14px', flex: 1,
  },
  btn: {
    padding: '10px 20px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold',
  },
  list: { display: 'flex', flexDirection: 'column', gap: '12px' },
  card: {
    backgroundColor: '#ffffff', padding: '16px', borderRadius: '8px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)',
  },
  cardHeader: { display: 'flex', justifyContent: 'space-between', alignItems: 'center' },
  cardTitle: { color: '#1e293b', margin: 0 },
  badge: {
    color: '#fff', padding: '4px 10px', borderRadius: '12px', fontSize: '12px',
    fontWeight: 'bold',
  },
  code: { color: '#6366f1', fontSize: '14px', margin: '8px 0 4px', fontWeight: '500' },
  meta: { color: '#64748b', fontSize: '13px', margin: '0 0 12px' },
  actions: { display: 'flex', gap: '8px' },
  actionBtn: {
    padding: '6px 14px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontSize: '13px',
  },
  empty: { color: '#64748b', textAlign: 'center' },
  toast: {
    position: 'fixed', bottom: '24px', left: '50%', transform: 'translateX(-50%)',
    backgroundColor: '#1e293b', color: '#fff', padding: '12px 24px', borderRadius: '8px',
    fontSize: '14px', fontWeight: '500', boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
    zIndex: 1000, animation: 'fadeIn 0.2s ease-out',
  },
};
