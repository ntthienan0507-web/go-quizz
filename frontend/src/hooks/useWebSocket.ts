import { useEffect, useRef, useState, useCallback } from 'react';
import { QuizSocket, WSMessage } from '../ws/socket';

interface LeaderboardEntry {
  user_id: string;
  username: string;
  score: number;
  rank: number;
}

interface QuestionData {
  question_idx: number;
  text: string;
  options: string[];
  time_limit: number;
}

interface AnswerResult {
  is_correct: boolean;
  points_awarded: number;
  correct_idx: number;
  your_total: number;
}

interface WelcomeData {
  quiz_title: string;
  total_questions: number;
  participants: number;
  mode: string;
}

interface AnswerProgress {
  question_idx: number;
  answered_count: number;
  total_players: number;
}

export function useWebSocket(code: string | null, token: string | null, guestName?: string) {
  const socketRef = useRef<QuizSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);
  const [welcome, setWelcome] = useState<WelcomeData | null>(null);
  const [currentQuestion, setCurrentQuestion] = useState<QuestionData | null>(null);
  const [answerResult, setAnswerResult] = useState<AnswerResult | null>(null);
  const [answerProgress, setAnswerProgress] = useState<AnswerProgress | null>(null);
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [participants, setParticipants] = useState(0);
  const [quizFinished, setQuizFinished] = useState(false);
  const [finalRankings, setFinalRankings] = useState<LeaderboardEntry[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!code || (!token && !guestName)) return;

    // Use latest token from localStorage (may have been refreshed by interceptor)
    const activeToken = token ? (localStorage.getItem('token') || token) : null;

    const socket = new QuizSocket(code, activeToken, guestName);
    socketRef.current = socket;

    socket.on('welcome', (msg: WSMessage) => {
      setWelcome(msg.payload);
      setParticipants(msg.payload.participants);
      setConnected(true);
      setReconnecting(false);
      setError(null);
    });

    socket.on('player_joined', (msg: WSMessage) => {
      setParticipants(msg.payload.participant_count);
    });

    socket.on('player_left', (msg: WSMessage) => {
      setParticipants(msg.payload.participant_count);
    });

    socket.on('new_question', (msg: WSMessage) => {
      setCurrentQuestion(msg.payload);
      setAnswerResult(null);
      setAnswerProgress(null);
    });

    socket.on('answer_result', (msg: WSMessage) => {
      setAnswerResult(msg.payload);
    });

    socket.on('answer_progress', (msg: WSMessage) => {
      setAnswerProgress(msg.payload);
    });

    socket.on('leaderboard_update', (msg: WSMessage) => {
      setLeaderboard(msg.payload.rankings);
    });

    socket.on('quiz_finished', (msg: WSMessage) => {
      setQuizFinished(true);
      setFinalRankings(msg.payload.rankings);
    });

    socket.on('error', (msg: WSMessage) => {
      setError(msg.payload.message);
    });

    socket.on('connection_error', (msg: WSMessage) => {
      setConnected(false);
      setReconnecting(false);
      setError(msg.payload.message);
    });

    socket.on('disconnected', () => {
      setConnected(false);
      setReconnecting(true);
      setError(null);
    });

    socket.on('reconnecting', (msg: WSMessage) => {
      setReconnecting(true);
    });

    socket.on('reconnected', () => {
      setReconnecting(false);
      setError(null);
      // connected will be set to true when 'welcome' is received
    });

    socket.connect();

    return () => {
      socket.disconnect();
      socketRef.current = null;
    };
  }, [code, token, guestName]);

  const submitAnswer = useCallback((questionIdx: number, selectedIdx: number, timeRemaining: number) => {
    socketRef.current?.send('submit_answer', {
      question_idx: questionIdx,
      selected_idx: selectedIdx,
      time_remaining: timeRemaining,
    });
  }, []);

  const nextQuestion = useCallback(() => {
    socketRef.current?.send('next_question');
  }, []);

  return {
    connected,
    reconnecting,
    welcome,
    currentQuestion,
    answerResult,
    answerProgress,
    leaderboard,
    participants,
    quizFinished,
    finalRankings,
    error,
    submitAnswer,
    nextQuestion,
  };
}
