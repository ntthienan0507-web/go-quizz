import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

const API_URL = process.env.REACT_APP_API_URL || (process.env.NODE_ENV === 'development' ? 'http://localhost:8080/api' : `https://${window.location.host}/api`);

const http = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true, // send httpOnly cookies
});

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

let isRefreshing = false;
let failedQueue: { resolve: (token: string) => void; reject: (err: unknown) => void }[] = [];

const processQueue = (error: unknown, token: string | null) => {
  failedQueue.forEach((p) => {
    if (token) p.resolve(token);
    else p.reject(error);
  });
  failedQueue = [];
};

http.interceptors.response.use(
  (response) => {
    // Unwrap the {status, data} envelope so callers access res.data directly
    if (response.data && typeof response.data === 'object' && 'data' in response.data) {
      response.data = response.data.data;
    }
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Don't try to refresh for auth or public endpoints
    const url = originalRequest?.url || '';
    const isPublicEndpoint = url.includes('/auth/') || url.includes('/quizzes/join/');

    if (error.response?.status !== 401 || originalRequest._retry || isPublicEndpoint) {
      if (error.response?.status === 401 && !isPublicEndpoint) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
      }
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({
          resolve: (token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            resolve(http(originalRequest));
          },
          reject,
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      // Refresh token is sent via httpOnly cookie automatically
      const res = await axios.post(`${API_URL}/auth/refresh`, {}, { withCredentials: true });
      const data = res.data?.data || res.data;
      const newToken = data.token;

      localStorage.setItem('token', newToken);
      if (data.user) {
        localStorage.setItem('user', JSON.stringify(data.user));
      }

      processQueue(null, newToken);
      originalRequest.headers.Authorization = `Bearer ${newToken}`;
      return http(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError, null);
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);


export default http;

// Auth
export const register = (username: string, email: string, password: string) =>
  http.post('/auth/register', { username, email, password });

export const login = (email: string, password: string) =>
  http.post('/auth/login', { email, password });

export const logout = () => http.post('/auth/logout');

// Quizzes
export const listQuizzes = () => http.get('/quizzes');
export const createQuiz = (title: string, time_per_question: number, mode: string = 'live') =>
  http.post('/quizzes', { title, time_per_question, mode });
export const getQuiz = (id: string) => http.get(`/quizzes/${id}`);
export const updateQuiz = (id: string, title: string, time_per_question: number) =>
  http.put(`/quizzes/${id}`, { title, time_per_question });
export const deleteQuiz = (id: string) => http.delete(`/quizzes/${id}`);
export const startQuiz = (id: string) => http.post(`/quizzes/${id}/start`);
export const finishQuiz = (id: string) => http.post(`/quizzes/${id}/finish`);
export const joinQuiz = (code: string) => http.get(`/quizzes/join/${code}`);

// Questions
export const listQuestions = (quizId: string) => http.get(`/quizzes/${quizId}/questions`);
export const createQuestion = (quizId: string, data: {
  text: string; options: string[]; correct_idx: number; points: number; order_num: number;
}) => http.post(`/quizzes/${quizId}/questions`, data);
export const updateQuestion = (qid: string, data: {
  text: string; options: string[]; correct_idx: number; points: number; order_num: number;
}) => http.put(`/questions/${qid}`, data);
export const deleteQuestion = (qid: string) => http.delete(`/questions/${qid}`);

// Player
export const getPlayerDashboard = () => http.get('/player/dashboard');
export const getPlayerHistory = (page = 1, limit = 10) =>
  http.get(`/player/history?page=${page}&limit=${limit}`);
export const getGlobalLeaderboard = (page = 1, limit = 20) =>
  http.get(`/player/leaderboard?page=${page}&limit=${limit}`);
export const getPlayerProfile = () => http.get('/player/profile');
export const updatePlayerProfile = (data: { username?: string; email?: string }) =>
  http.put('/player/profile', data);
