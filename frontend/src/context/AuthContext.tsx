import React, { createContext, useContext, useState, ReactNode } from 'react';
import { login as apiLogin, register as apiRegister, logout as apiLogout } from '../api/http';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

function getInitialToken(): string | null {
  return localStorage.getItem('token');
}

function getInitialUser(): User | null {
  const saved = localStorage.getItem('user');
  if (saved) {
    try { return JSON.parse(saved); } catch { return null; }
  }
  return null;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(getInitialUser);
  const [token, setToken] = useState<string | null>(getInitialToken);

  const login = async (email: string, password: string) => {
    const res = await apiLogin(email, password);
    const { token: t, user: u } = res.data;
    setToken(t);
    setUser(u);
    localStorage.setItem('token', t);
    localStorage.setItem('user', JSON.stringify(u));
  };

  const register = async (username: string, email: string, password: string) => {
    await apiRegister(username, email, password);
  };

  const logout = () => {
    apiLogout().catch(() => {}); // best-effort server-side cookie clear
    setToken(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  };

  return (
    <AuthContext.Provider value={{ user, token, login, register, logout, isAuthenticated: !!token }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
