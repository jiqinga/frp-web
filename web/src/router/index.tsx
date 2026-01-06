import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ProtectedRoute } from '../components/ProtectedRoute';
import { MainLayout } from '../components/MainLayout';
import { LoadingFallback } from '../components/LoadingFallback';

export const router = createBrowserRouter([
  {
    path: '/login',
    lazy: () => import('../pages/Login'),
    HydrateFallback: LoadingFallback,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <MainLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        lazy: () => import('../pages/Dashboard'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'clients',
        lazy: () => import('../pages/Clients'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'proxies',
        lazy: () => import('../pages/Proxies'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'frp-servers',
        lazy: () => import('../pages/FrpServers'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'logs',
        lazy: () => import('../pages/Logs'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'settings',
        lazy: () => import('../pages/Settings'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'realtime',
        lazy: () => import('../pages/RealtimeMonitor'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'alerts',
        lazy: () => import('../pages/AlertRules'),
        HydrateFallback: LoadingFallback,
      },
      {
        path: 'certificates',
        lazy: () => import('../pages/Certificates'),
        HydrateFallback: LoadingFallback,
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);