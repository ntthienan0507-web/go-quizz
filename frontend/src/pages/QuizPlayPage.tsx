import React, { useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../hooks/useWebSocket';
import QuestionCard from '../components/QuestionCard';
import Timer, { useTimeRemaining } from '../components/Timer';
import Leaderboard from '../components/Leaderboard';

export default function QuizPlayPage() {
  const { code } = useParams<{ code: string }>();
  const { token, user } = useAuth();
  const navigate = useNavigate();
  const guestName = sessionStorage.getItem('guestName') || undefined;
  const {
    connected, reconnecting, welcome, currentQuestion, answerResult, answerProgress,
    leaderboard, participants, quizFinished, finalRankings,
    error, submitAnswer, nextQuestion,
  } = useWebSocket(code || null, token, guestName);

  const [answered, setAnswered] = useState(false);
  const [selectedIdx, setSelectedIdx] = useState<number | undefined>(undefined);
  const [questionKey, setQuestionKey] = useState(0);

  const isHost = user?.role === 'admin';
  const isSelfPaced = welcome?.mode === 'self_paced';

  const timeRemaining = useTimeRemaining(
    currentQuestion?.time_limit || 30,
    currentQuestion?.question_idx ?? 0
  );

  const handleAnswer = useCallback((idx: number) => {
    if (!currentQuestion || answered) return;
    setAnswered(true);
    setSelectedIdx(idx);
    submitAnswer(currentQuestion.question_idx, idx, timeRemaining);
  }, [currentQuestion, answered, submitAnswer, timeRemaining]);

  const handleTimeUp = useCallback(() => {
    if (!answered) {
      setAnswered(true);
    }
  }, [answered]);

  const handleNextQuestion = () => {
    setAnswered(false);
    setSelectedIdx(undefined);
    setQuestionKey((k) => k + 1);
    nextQuestion();
  };

  // Reset answered state when new question arrives
  React.useEffect(() => {
    if (currentQuestion) {
      setAnswered(false);
      setSelectedIdx(undefined);
      setQuestionKey((k) => k + 1);
    }
  }, [currentQuestion?.question_idx]); // eslint-disable-line

  if (quizFinished) {
    return (
      <div style={styles.container}>
        <div style={styles.main}>
          <h1 style={styles.heading}>Quiz Finished!</h1>
          <Leaderboard rankings={finalRankings} title="Final Results" />
        </div>
      </div>
    );
  }

  if (!connected) {
    return (
      <div style={styles.container}>
        <div style={styles.center}>
          {error ? (
            <div style={styles.errorPage}>
              <div style={styles.errorIcon}>!</div>
              <h2 style={styles.errorTitle}>Connection Failed</h2>
              <p style={styles.errorMsg}>{error}</p>
              <div style={styles.errorActions}>
                <button style={styles.retryBtn} onClick={() => window.location.reload()}>
                  Try Again
                </button>
                <button style={styles.backBtn} onClick={() => navigate(`/join?code=${code}`)}>
                  Go Back
                </button>
              </div>
            </div>
          ) : reconnecting ? (
            <h2 style={{ color: '#f59e0b' }}>Connection lost. Reconnecting...</h2>
          ) : (
            <h2 style={{ color: '#1e293b' }}>Connecting to quiz...</h2>
          )}
        </div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <div style={styles.main}>
        {!currentQuestion ? (
          <div style={styles.lobby}>
            <h1 style={styles.heading}>{welcome?.quiz_title || 'Quiz Lobby'}</h1>
            <p style={styles.info}>
              {welcome?.total_questions} questions | {participants} participants
            </p>
            {isSelfPaced ? (
              <>
                <p style={styles.waitText}>Self-paced quiz — start when you're ready!</p>
                <button style={styles.startBtn} onClick={handleNextQuestion}>
                  Start Quiz
                </button>
              </>
            ) : (
              <>
                <p style={styles.waitText}>
                  {isHost ? 'You are the host. Start when ready!' : 'Waiting for the host to start...'}
                </p>
                {isHost && (
                  <button style={styles.startBtn} onClick={handleNextQuestion}>
                    Start First Question
                  </button>
                )}
              </>
            )}
          </div>
        ) : (
          <div>
            {/* Host info bar */}
            {isHost && !isSelfPaced && (
              <div style={styles.hostBar}>
                <span style={styles.hostBadge}>HOST</span>
                <span style={styles.hostInfo}>
                  Q{currentQuestion.question_idx + 1}/{welcome?.total_questions}
                </span>
                <span style={styles.hostInfo}>
                  {answerProgress
                    ? `${answerProgress.answered_count}/${answerProgress.total_players} answered`
                    : `0/${participants} answered`
                  }
                </span>
              </div>
            )}

            {/* Player progress bar (non-host) */}
            {!isHost && !isSelfPaced && answerProgress && (
              <div style={styles.progressInfo}>
                {answerProgress.answered_count}/{answerProgress.total_players} players answered
              </div>
            )}

            <Timer
              duration={currentQuestion.time_limit}
              onTimeUp={handleTimeUp}
              resetKey={questionKey}
            />
            <QuestionCard
              questionIdx={currentQuestion.question_idx}
              text={currentQuestion.text}
              options={currentQuestion.options}
              onAnswer={handleAnswer}
              disabled={answered}
              selectedIdx={selectedIdx}
              correctIdx={answerResult?.correct_idx}
            />
            {answerResult && (
              <div style={{
                ...styles.resultBanner,
                backgroundColor: answerResult.is_correct ? '#10b981' : '#ef4444',
              }}>
                {answerResult.is_correct
                  ? `Correct! +${answerResult.points_awarded} points`
                  : 'Wrong answer!'}
                {' | Total: '}{Math.round(answerResult.your_total)}
              </div>
            )}
            {(isHost || (answered && isSelfPaced)) && (
              <div style={styles.adminBar}>
                <button style={styles.nextBtn} onClick={handleNextQuestion}>
                  Next Question
                </button>
              </div>
            )}
          </div>
        )}
        {reconnecting && (
          <div style={styles.reconnecting}>Reconnecting...</div>
        )}
        {error && <div style={styles.error}>{error}</div>}
      </div>
      <div style={styles.sidebar}>
        <div style={styles.statsCard}>
          <div style={styles.statRow}>
            <span style={styles.statLabel}>Players</span>
            <span style={styles.statValue}>{participants}</span>
          </div>
          {currentQuestion && !isSelfPaced && (
            <div style={styles.statRow}>
              <span style={styles.statLabel}>Answered</span>
              <span style={styles.statValue}>
                {answerProgress ? answerProgress.answered_count : 0}/{participants}
              </span>
            </div>
          )}
          {currentQuestion && (
            <div style={styles.statRow}>
              <span style={styles.statLabel}>Question</span>
              <span style={styles.statValue}>
                {currentQuestion.question_idx + 1}/{welcome?.total_questions}
              </span>
            </div>
          )}
        </div>
        <Leaderboard rankings={leaderboard} />
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex', minHeight: 'calc(100vh - 60px)',
  },
  main: { flex: 1, padding: '24px' },
  sidebar: { width: '300px', padding: '24px', borderLeft: '1px solid #e2e8f0', display: 'flex', flexDirection: 'column', gap: '16px' },
  center: { display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' },
  lobby: { textAlign: 'center', paddingTop: '60px' },
  heading: { color: '#1e293b', fontSize: '32px' },
  info: { color: '#64748b', fontSize: '16px', marginTop: '8px' },
  waitText: { color: '#6366f1', fontSize: '18px', marginTop: '24px' },
  startBtn: {
    marginTop: '24px', padding: '14px 32px', backgroundColor: '#10b981', color: '#fff',
    border: 'none', borderRadius: '8px', fontSize: '18px', cursor: 'pointer', fontWeight: 'bold',
  },
  hostBar: {
    display: 'flex', alignItems: 'center', gap: '12px', padding: '10px 16px',
    backgroundColor: '#fff', borderRadius: '8px', marginBottom: '12px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  hostBadge: {
    backgroundColor: '#6366f1', color: '#fff', padding: '2px 8px', borderRadius: '4px',
    fontSize: '11px', fontWeight: 'bold', letterSpacing: '1px',
  },
  hostInfo: {
    color: '#64748b', fontSize: '14px',
  },
  progressInfo: {
    textAlign: 'center', color: '#64748b', fontSize: '14px', marginBottom: '8px',
  },
  resultBanner: {
    textAlign: 'center', padding: '12px', borderRadius: '8px',
    color: '#fff', fontWeight: 'bold', fontSize: '18px', margin: '16px auto',
    maxWidth: '400px',
  },
  adminBar: { textAlign: 'center', marginTop: '16px' },
  nextBtn: {
    padding: '12px 28px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '8px', fontSize: '16px',
    cursor: 'pointer', fontWeight: 'bold',
  },
  statsCard: {
    backgroundColor: '#ffffff', borderRadius: '8px', padding: '16px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)',
  },
  statRow: {
    display: 'flex', justifyContent: 'space-between', padding: '6px 0',
    borderBottom: '1px solid #f1f5f9',
  },
  statLabel: { color: '#64748b', fontSize: '13px' },
  statValue: { color: '#1e293b', fontWeight: 'bold', fontSize: '13px' },
  errorPage: {
    textAlign: 'center', maxWidth: '420px',
  },
  errorIcon: {
    width: '64px', height: '64px', borderRadius: '50%', backgroundColor: '#fef2f2',
    color: '#dc2626', fontSize: '32px', fontWeight: 'bold', lineHeight: '64px',
    margin: '0 auto 16px', border: '2px solid #fecaca',
  },
  errorTitle: {
    color: '#1e293b', fontSize: '24px', margin: '0 0 8px',
  },
  errorMsg: {
    color: '#64748b', fontSize: '15px', lineHeight: '1.5', margin: '0 0 24px',
  },
  errorActions: {
    display: 'flex', gap: '12px', justifyContent: 'center',
  },
  retryBtn: {
    padding: '12px 28px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '8px', fontSize: '16px', cursor: 'pointer', fontWeight: 'bold',
  },
  backBtn: {
    padding: '12px 28px', backgroundColor: '#fff', color: '#64748b',
    border: '1px solid #e2e8f0', borderRadius: '8px', fontSize: '16px', cursor: 'pointer',
  },
  reconnecting: {
    backgroundColor: '#fffbeb', color: '#d97706', padding: '8px 12px',
    borderRadius: '6px', marginTop: '12px', textAlign: 'center',
    border: '1px solid #fde68a', fontWeight: 'bold',
  },
  error: {
    backgroundColor: '#fef2f2', color: '#dc2626', padding: '8px 12px',
    borderRadius: '6px', marginTop: '12px', textAlign: 'center',
    border: '1px solid #fecaca',
  },
};
