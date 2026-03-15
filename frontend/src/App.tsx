import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Navbar from './components/Navbar';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import QuestionsPage from './pages/QuestionsPage';
import JoinPage from './pages/JoinPage';
import QuizPlayPage from './pages/QuizPlayPage';
import PlayerDashboardPage from './pages/PlayerDashboardPage';
import QuizHistoryPage from './pages/QuizHistoryPage';
import GlobalLeaderboardPage from './pages/GlobalLeaderboardPage';
import PlayerProfilePage from './pages/PlayerProfilePage';

function ProtectedRoute({ children }: { children: React.ReactElement }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? children : <Navigate to="/login" />;
}

function HomeRedirect() {
  const { isAuthenticated, user } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" />;
  return <Navigate to={user?.role === 'admin' ? '/dashboard' : '/player/dashboard'} />;
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <div style={{ backgroundColor: '#f8fafc', minHeight: '100vh' }}>
          <Navbar />
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/join" element={<JoinPage />} />
            <Route path="/play/:code" element={<QuizPlayPage />} />
            <Route path="/dashboard" element={
              <ProtectedRoute><DashboardPage /></ProtectedRoute>
            } />
            <Route path="/quiz/:id/questions" element={
              <ProtectedRoute><QuestionsPage /></ProtectedRoute>
            } />
            <Route path="/player/dashboard" element={
              <ProtectedRoute><PlayerDashboardPage /></ProtectedRoute>
            } />
            <Route path="/player/history" element={
              <ProtectedRoute><QuizHistoryPage /></ProtectedRoute>
            } />
            <Route path="/player/profile" element={
              <ProtectedRoute><PlayerProfilePage /></ProtectedRoute>
            } />
            <Route path="/leaderboard" element={
              <ProtectedRoute><GlobalLeaderboardPage /></ProtectedRoute>
            } />
            <Route path="/" element={<HomeRedirect />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
