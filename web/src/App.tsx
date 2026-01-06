import { useEffect, useState, useRef } from 'react';
import { RouterProvider } from 'react-router-dom';
import { router } from './router';
import { useAuthStore } from './store/auth';
import { FullPageSpinner } from './components/ui';

function App() {
  const [loading, setLoading] = useState(true);
  const { token, fetchProfile } = useAuthStore();
  const initRef = useRef(false);

  useEffect(() => {
    // 使用 ref 确保只执行一次初始化
    if (initRef.current) return;
    initRef.current = true;

    const initAuth = async () => {
      if (token) {
        try {
          await fetchProfile();
        } catch {
          useAuthStore.getState().logout();
        }
      }
      setLoading(false);
    };
    initAuth();
  }, [token, fetchProfile]);

  if (loading) {
    return <FullPageSpinner label="正在加载..." />;
  }

  return <RouterProvider router={router} />;
}

export default App;
