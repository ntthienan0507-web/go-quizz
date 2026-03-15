import React, { useEffect, useState, useRef } from 'react';

interface Props {
  duration: number;
  onTimeUp: () => void;
  resetKey: number;
}

export default function Timer({ duration, onTimeUp, resetKey }: Props) {
  const [timeLeft, setTimeLeft] = useState(duration);
  const callbackRef = useRef(onTimeUp);
  callbackRef.current = onTimeUp;

  useEffect(() => {
    setTimeLeft(duration);
  }, [duration, resetKey]);

  useEffect(() => {
    if (timeLeft <= 0) {
      callbackRef.current();
      return;
    }
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft]);

  const pct = (timeLeft / duration) * 100;
  const color = timeLeft > 10 ? '#10b981' : timeLeft > 5 ? '#f59e0b' : '#ef4444';

  return (
    <div style={styles.container}>
      <div style={{ ...styles.bar, width: `${pct}%`, backgroundColor: color }} />
      <span style={styles.text}>{timeLeft}s</span>
    </div>
  );
}

export function useTimeRemaining(duration: number, resetKey: number): number {
  const [timeLeft, setTimeLeft] = useState(duration);

  useEffect(() => {
    setTimeLeft(duration);
  }, [duration, resetKey]);

  useEffect(() => {
    if (timeLeft <= 0) return;
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft]);

  return timeLeft;
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    position: 'relative', height: '30px', backgroundColor: '#e2e8f0',
    borderRadius: '15px', overflow: 'hidden', margin: '16px auto', maxWidth: '400px',
  },
  bar: {
    height: '100%', borderRadius: '15px', transition: 'width 1s linear',
  },
  text: {
    position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)',
    color: '#fff', fontWeight: 'bold', fontSize: '14px',
  },
};
