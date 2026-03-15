import React from 'react';

interface Props {
  questionIdx: number;
  text: string;
  options: string[];
  onAnswer: (selectedIdx: number) => void;
  disabled: boolean;
  selectedIdx?: number;
  correctIdx?: number;
}

const OPTION_COLORS = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b'];

export default function QuestionCard({ questionIdx, text, options, onAnswer, disabled, selectedIdx, correctIdx }: Props) {
  return (
    <div style={styles.container}>
      <div style={styles.questionNum}>Question {questionIdx + 1}</div>
      <h2 style={styles.text}>{text}</h2>
      <div style={styles.options}>
        {options.map((opt, idx) => {
          let bg = OPTION_COLORS[idx % OPTION_COLORS.length];
          if (correctIdx !== undefined) {
            if (idx === correctIdx) bg = '#10b981';
            else if (idx === selectedIdx) bg = '#ef4444';
          }
          return (
            <button
              key={idx}
              onClick={() => onAnswer(idx)}
              disabled={disabled}
              style={{ ...styles.option, backgroundColor: bg, opacity: disabled ? 0.7 : 1 }}
            >
              {opt}
            </button>
          );
        })}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: { textAlign: 'center', padding: '20px' },
  questionNum: { color: '#64748b', fontSize: '14px', marginBottom: '8px' },
  text: { color: '#1e293b', fontSize: '24px', marginBottom: '24px' },
  options: { display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', maxWidth: '600px', margin: '0 auto' },
  option: {
    color: '#fff', border: 'none', padding: '20px', borderRadius: '8px',
    fontSize: '16px', cursor: 'pointer', fontWeight: 'bold',
  },
};
