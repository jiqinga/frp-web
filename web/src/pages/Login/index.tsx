import { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Network, User, Lock } from 'lucide-react';
import { useAuthStore } from '../../store/auth';
import { Input, Button, Alert } from '../../components/ui';

export function Component() {
  const navigate = useNavigate();
  const login = useAuthStore((state) => state.login);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    
    if (!username || !password) {
      setError('请输入用户名和密码');
      return;
    }

    setLoading(true);
    try {
      await login(username, password);
      navigate('/');
    } catch (err) {
      setError('登录失败，请检查用户名和密码');
      console.error('[Login] 登录失败:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative min-h-screen w-full overflow-hidden bg-surface-elevated">
      {/* 动态背景 */}
      <div className="absolute inset-0">
        {/* 网格背景 */}
        <div
          className="absolute inset-0 opacity-20"
          style={{
            backgroundImage: `
              linear-gradient(rgba(99, 102, 241, 0.1) 1px, transparent 1px),
              linear-gradient(90deg, rgba(99, 102, 241, 0.1) 1px, transparent 1px)
            `,
            backgroundSize: '50px 50px'
          }}
        />
        
        {/* 渐变光晕 */}
        <div className="absolute -left-40 -top-40 h-80 w-80 rounded-full bg-indigo-500/20 blur-[100px]" />
        <div className="absolute -bottom-40 -right-40 h-80 w-80 rounded-full bg-cyan-500/20 blur-[100px]" />
        <div className="absolute left-1/2 top-1/2 h-96 w-96 -translate-x-1/2 -translate-y-1/2 rounded-full bg-purple-500/10 blur-[120px]" />
        
        {/* 扫描线动画 */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="scan-line absolute left-0 right-0 h-px bg-gradient-to-r from-transparent via-indigo-500/50 to-transparent" />
        </div>
        
        {/* 浮动粒子 */}
        <div className="absolute inset-0">
          {[...Array(20)].map((_, i) => (
            <div
              key={i}
              className="absolute h-1 w-1 rounded-full bg-indigo-400/30"
              style={{
                left: `${Math.random() * 100}%`,
                top: `${Math.random() * 100}%`,
                animation: `float ${3 + Math.random() * 4}s ease-in-out infinite`,
                animationDelay: `${Math.random() * 2}s`
              }}
            />
          ))}
        </div>
      </div>

      {/* 登录表单 */}
      <div className="relative z-10 flex min-h-screen items-center justify-center p-4">
        <div className="w-full max-w-md">
          {/* Logo 和标题 */}
          <div className="mb-8 text-center">
            <div className="mb-4 inline-flex items-center justify-center">
              <div className="relative">
                <div className="absolute inset-0 animate-pulse rounded-2xl bg-indigo-500/20 blur-xl" />
                <div className="relative flex h-16 w-16 items-center justify-center rounded-2xl border border-indigo-500/30 bg-surface/80 backdrop-blur-sm">
                  <Network className="h-8 w-8 text-indigo-400" />
                </div>
              </div>
            </div>
            <h1 className="mb-2 bg-gradient-to-r from-foreground via-indigo-300 to-indigo-400 bg-clip-text text-3xl font-bold tracking-tight text-transparent">
              FRP 管理面板
            </h1>
            <p className="text-foreground-muted">
              安全登录到您的控制台
            </p>
          </div>

          {/* 登录卡片 */}
          <div className="relative">
            {/* 卡片发光边框 */}
            <div className="absolute -inset-px rounded-2xl bg-gradient-to-r from-indigo-500/50 via-purple-500/50 to-cyan-500/50 opacity-50 blur-sm" />
            
            <div className="relative rounded-2xl border border-border bg-surface/80 p-8 backdrop-blur-xl">
              <form className="space-y-6" onSubmit={onSubmit}>
                {error && <Alert type="error" message={error} />}

                {/* 用户名输入 */}
                <Input
                  label="用户名"
                  id="username"
                  type="text"
                  placeholder="请输入用户名"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  prefix={<User className="h-5 w-5" />}
                  size="lg"
                />

                {/* 密码输入 */}
                <Input
                  label="密码"
                  id="password"
                  type="password"
                  placeholder="请输入密码"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  prefix={<Lock className="h-5 w-5" />}
                  size="lg"
                />

                {/* 登录按钮 */}
                <Button
                  type="submit"
                  variant="primary"
                  size="lg"
                  fullWidth
                  loading={loading}
                >
                  {loading ? '登录中...' : '登 录'}
                </Button>
              </form>

              {/* 装饰线条 */}
              <div className="mt-8 flex items-center gap-4">
                <div className="h-px flex-1 bg-gradient-to-r from-transparent via-border to-transparent" />
                <span className="text-xs text-foreground-subtle">安全连接</span>
                <div className="h-px flex-1 bg-gradient-to-r from-transparent via-border to-transparent" />
              </div>

              {/* 安全提示 */}
              <div className="mt-4 flex items-center justify-center gap-2 text-xs text-foreground-subtle">
                <div className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
                <span>256-bit SSL 加密保护</span>
              </div>
            </div>
          </div>

          {/* 版权信息 */}
          <div className="mt-8 text-center">
            <p className="text-sm text-foreground-subtle">
              © {new Date().getFullYear()} FRP Services Inc. All Rights Reserved.
            </p>
          </div>
        </div>
      </div>

      {/* 添加自定义动画样式 */}
      <style>{`
        @keyframes float {
          0%, 100% {
            transform: translateY(0) translateX(0);
            opacity: 0.3;
          }
          50% {
            transform: translateY(-20px) translateX(10px);
            opacity: 0.6;
          }
        }
        
        @keyframes shimmer {
          100% {
            transform: translateX(100%);
          }
        }
        
        .scan-line {
          animation: scan 4s linear infinite;
        }
        
        @keyframes scan {
          0% {
            top: -2px;
          }
          100% {
            top: 100%;
          }
        }
      `}</style>
    </div>
  );
}