import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { listQuestions, createQuestion, deleteQuestion } from '../api/http';

interface Question {
  id: string;
  text: string;
  options: string[];
  correct_idx: number;
  points: number;
  order_num: number;
}

export default function QuestionsPage() {
  const { id } = useParams<{ id: string }>();
  const [questions, setQuestions] = useState<Question[]>([]);
  const [text, setText] = useState('');
  const [options, setOptions] = useState(['', '', '', '']);
  const [correctIdx, setCorrectIdx] = useState(0);
  const [points, setPoints] = useState(10);

  const fetchQuestions = async () => {
    if (!id) return;
    const res = await listQuestions(id);
    setQuestions(res.data.items || []);
  };

  useEffect(() => { fetchQuestions(); }, [id]); // eslint-disable-line

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !text.trim()) return;
    const validOptions = options.filter((o) => o.trim());
    if (validOptions.length < 2) return;
    await createQuestion(id, {
      text,
      options: validOptions,
      correct_idx: correctIdx,
      points,
      order_num: questions.length,
    });
    setText('');
    setOptions(['', '', '', '']);
    setCorrectIdx(0);
    fetchQuestions();
  };

  const handleDelete = async (qid: string) => {
    await deleteQuestion(qid);
    fetchQuestions();
  };

  const updateOption = (idx: number, val: string) => {
    const newOpts = [...options];
    newOpts[idx] = val;
    setOptions(newOpts);
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.heading}>Manage Questions</h1>

      <form onSubmit={handleAdd} style={styles.form}>
        <input
          style={styles.input} placeholder="Question text" value={text}
          onChange={(e) => setText(e.target.value)} required
        />
        <div style={styles.optionsGrid}>
          {options.map((opt, idx) => (
            <div key={idx} style={styles.optionRow}>
              <input
                style={{ ...styles.input, flex: 1 }}
                placeholder={`Option ${idx + 1}`}
                value={opt}
                onChange={(e) => updateOption(idx, e.target.value)}
                required={idx < 2}
              />
              <label style={styles.radioLabel}>
                <input
                  type="radio" name="correct" checked={correctIdx === idx}
                  onChange={() => setCorrectIdx(idx)}
                />
                Correct
              </label>
            </div>
          ))}
        </div>
        <div style={styles.row}>
          <label style={styles.label}>Points:</label>
          <input
            style={{ ...styles.input, width: '80px' }} type="number" value={points}
            onChange={(e) => setPoints(Number(e.target.value))} min={1}
          />
          <button type="submit" style={styles.btn}>Add Question</button>
        </div>
      </form>

      <div style={styles.list}>
        {questions.map((q, idx) => (
          <div key={q.id} style={styles.card}>
            <div style={styles.cardHeader}>
              <strong style={{ color: '#1e293b' }}>Q{idx + 1}:</strong>
              <span style={{ color: '#1e293b', flex: 1, marginLeft: '8px' }}>{q.text}</span>
              <button style={styles.deleteBtn} onClick={() => handleDelete(q.id)}>Delete</button>
            </div>
            <div style={styles.optionsList}>
              {(q.options || []).map((opt: string, i: number) => (
                <span key={i} style={{
                  ...styles.optionTag,
                  backgroundColor: i === q.correct_idx ? '#10b981' : '#e2e8f0',
                  color: i === q.correct_idx ? '#fff' : '#475569',
                }}>
                  {opt}
                </span>
              ))}
            </div>
            <span style={styles.meta}>{q.points} pts</span>
          </div>
        ))}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: { padding: '24px', maxWidth: '800px', margin: '0 auto' },
  heading: { color: '#1e293b', marginBottom: '24px' },
  form: {
    backgroundColor: '#ffffff', padding: '20px', borderRadius: '8px', marginBottom: '24px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)',
  },
  input: {
    padding: '10px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#f8fafc', color: '#1e293b', fontSize: '14px',
    marginBottom: '8px', boxSizing: 'border-box' as const,
  },
  optionsGrid: { display: 'flex', flexDirection: 'column', gap: '4px', marginBottom: '12px' },
  optionRow: { display: 'flex', alignItems: 'center', gap: '8px' },
  radioLabel: { color: '#64748b', fontSize: '13px', whiteSpace: 'nowrap' },
  row: { display: 'flex', alignItems: 'center', gap: '8px' },
  label: { color: '#64748b', fontSize: '14px' },
  btn: {
    padding: '10px 20px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold',
    marginLeft: 'auto',
  },
  list: { display: 'flex', flexDirection: 'column', gap: '8px' },
  card: {
    backgroundColor: '#ffffff', padding: '12px 16px', borderRadius: '8px',
    boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
  },
  cardHeader: { display: 'flex', alignItems: 'center' },
  deleteBtn: {
    background: '#ef4444', color: '#fff', border: 'none', padding: '4px 10px',
    borderRadius: '4px', cursor: 'pointer', fontSize: '12px',
  },
  optionsList: { display: 'flex', gap: '6px', marginTop: '8px', flexWrap: 'wrap' },
  optionTag: {
    color: '#fff', padding: '4px 10px', borderRadius: '6px', fontSize: '13px',
  },
  meta: { color: '#64748b', fontSize: '12px' },
};
