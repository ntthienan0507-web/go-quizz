import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { joinQuiz } from '../api/http';

const ADJECTIVES = [
  'Swift', 'Brave', 'Clever', 'Happy', 'Lucky', 'Mighty', 'Quick', 'Sharp',
  'Bold', 'Calm', 'Cool', 'Eager', 'Fast', 'Gentle', 'Kind', 'Wise',
  'Bright', 'Fierce', 'Noble', 'Proud', 'Silent', 'Witty', 'Cosmic', 'Epic',
];

const ANIMALS = [
  'Panda', 'Fox', 'Eagle', 'Tiger', 'Wolf', 'Bear', 'Hawk', 'Lion',
  'Otter', 'Raven', 'Falcon', 'Lynx', 'Shark', 'Owl', 'Dragon', 'Phoenix',
  'Cobra', 'Dolphin', 'Jaguar', 'Panther', 'Viper', 'Koala', 'Bunny', 'Penguin',
];

function generateRandomName(): string {
  const adj = ADJECTIVES[Math.floor(Math.random() * ADJECTIVES.length)];
  const animal = ANIMALS[Math.floor(Math.random() * ANIMALS.length)];
  const num = Math.floor(Math.random() * 100);
  return `${adj}${animal}${num}`;
}

declare const google: any;

const GOOGLE_CLIENT_ID = process.env.REACT_APP_GOOGLE_CLIENT_ID;

export default function JoinPage() {
  const [searchParams] = useSearchParams();
  const [code, setCode] = useState(searchParams.get('code') || '');
  const [guestName, setGuestName] = useState(sessionStorage.getItem('guestName') || '');
  const [error, setError] = useState('');
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const googleBtnRef = useRef<HTMLDivElement>(null);

  // Initialize Google Sign-In
  useEffect(() => {
    if (isAuthenticated || !GOOGLE_CLIENT_ID) return;

    const initGoogle = () => {
      if (typeof google === 'undefined' || !google.accounts) return;

      google.accounts.id.initialize({
        client_id: GOOGLE_CLIENT_ID,
        callback: (response: any) => {
          // Decode JWT to get name
          const payload = JSON.parse(atob(response.credential.split('.')[1]));
          const name = payload.name || payload.email?.split('@')[0] || 'GoogleUser';
          setGuestName(name);
          sessionStorage.setItem('guestName', name);
        },
      });

      if (googleBtnRef.current) {
        google.accounts.id.renderButton(googleBtnRef.current, {
          theme: 'outline',
          size: 'large',
          width: 320,
          text: 'continue_with',
        });
      }
    };

    // GSI script might load after component mounts
    if (typeof google !== 'undefined' && google.accounts) {
      initGoogle();
    } else {
      const interval = setInterval(() => {
        if (typeof google !== 'undefined' && google.accounts) {
          clearInterval(interval);
          initGoogle();
        }
      }, 200);
      return () => clearInterval(interval);
    }
  }, [isAuthenticated]);

  const handleRandomName = () => {
    const name = generateRandomName();
    setGuestName(name);
  };

  const handleJoin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!isAuthenticated && !guestName.trim()) {
      setError('Please enter your name');
      return;
    }

    // Save guest name in session
    if (!isAuthenticated) {
      sessionStorage.setItem('guestName', guestName.trim());
    }

    try {
      await joinQuiz(code.trim());
      navigate(`/play/${code.trim()}`);
    } catch (err: any) {
      const status = err.response?.status;
      const msg = err.response?.data?.message;
      if (status === 400) {
        setError(msg || 'Quiz is not active yet. Ask the host to start it.');
      } else if (status === 404) {
        setError('Quiz not found. Check the code and try again.');
      } else {
        setError(msg || 'Something went wrong. Please try again.');
      }
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h2 style={styles.title}>Join a Quiz</h2>
        {error && <div style={styles.error}>{error}</div>}
        <form onSubmit={handleJoin}>
          {!isAuthenticated && (
            <>
              <div style={styles.nameRow}>
                <input
                  style={styles.nameInput}
                  placeholder="Your name"
                  value={guestName}
                  onChange={(e) => setGuestName(e.target.value)}
                  required
                  maxLength={30}
                />
                <button type="button" style={styles.randomBtn} onClick={handleRandomName} title="Random name">
                  🎲
                </button>
              </div>
              {GOOGLE_CLIENT_ID && (
                <div style={styles.googleWrapper}>
                  <div style={styles.divider}>
                    <span style={styles.dividerText}>or</span>
                  </div>
                  <div ref={googleBtnRef} style={styles.googleBtn} />
                </div>
              )}
            </>
          )}
          <input
            style={styles.input}
            placeholder="Enter Quiz Code"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            required
            maxLength={10}
          />
          <button type="submit" style={styles.btn}>Join</button>
        </form>
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex', justifyContent: 'center', alignItems: 'center',
    minHeight: 'calc(100vh - 60px)',
  },
  card: {
    backgroundColor: '#ffffff', padding: '40px', borderRadius: '12px',
    width: '100%', maxWidth: '400px', textAlign: 'center',
    boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -2px rgba(0,0,0,0.1)',
  },
  title: { color: '#1e293b', marginBottom: '24px' },
  nameRow: {
    display: 'flex', gap: '8px', marginBottom: '12px',
  },
  nameInput: {
    flex: 1, padding: '12px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#f8fafc', color: '#1e293b', fontSize: '16px', textAlign: 'center',
    boxSizing: 'border-box',
  },
  randomBtn: {
    padding: '12px 14px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#f8fafc', cursor: 'pointer', fontSize: '18px',
  },
  googleWrapper: {
    marginBottom: '16px',
  },
  divider: {
    display: 'flex', alignItems: 'center', margin: '12px 0',
  },
  dividerText: {
    color: '#94a3b8', fontSize: '13px', padding: '0 12px',
    backgroundColor: '#fff', position: 'relative' as const, zIndex: 1,
    margin: '0 auto',
  },
  googleBtn: {
    display: 'flex', justifyContent: 'center',
  },
  input: {
    width: '100%', padding: '16px', borderRadius: '6px', border: '1px solid #e2e8f0',
    backgroundColor: '#f8fafc', color: '#1e293b', fontSize: '20px', textAlign: 'center',
    letterSpacing: '4px', textTransform: 'uppercase', boxSizing: 'border-box',
    marginBottom: '16px',
  },
  btn: {
    width: '100%', padding: '14px', backgroundColor: '#6366f1', color: '#fff',
    border: 'none', borderRadius: '6px', fontSize: '18px', cursor: 'pointer',
    fontWeight: 'bold',
  },
  error: {
    backgroundColor: '#fef2f2', color: '#dc2626', padding: '8px 12px',
    borderRadius: '6px', marginBottom: '12px', fontSize: '14px',
    border: '1px solid #fecaca',
  },
};
