import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  Server,
  Network,
  Activity,
  Bell,
  HardDrive,
  FileText,
  Settings,
  ChevronLeft,
  ChevronRight,
  User,
  LogOut,
  Key,
  Menu,
  X,
  Shield,
} from 'lucide-react';
import { useAuthStore } from '../store/auth';
import { ChangePasswordModal } from './ChangePasswordModal';
import { ThemeToggle } from './ThemeToggle';
import { Dropdown, Tooltip, type DropdownItem } from './ui';
import { cn } from '../utils/cn';

interface MenuItem {
  key: string;
  icon: React.ReactNode;
  label: string;
}

const menuItems: MenuItem[] = [
  { key: '/', icon: <LayoutDashboard className="h-5 w-5" />, label: '仪表盘' },
  { key: '/clients', icon: <Server className="h-5 w-5" />, label: '客户端管理' },
  { key: '/proxies', icon: <Network className="h-5 w-5" />, label: '代理管理' },
  { key: '/realtime', icon: <Activity className="h-5 w-5" />, label: '实时监控' },
  { key: '/alerts', icon: <Bell className="h-5 w-5" />, label: '告警管理' },
  { key: '/frp-servers', icon: <HardDrive className="h-5 w-5" />, label: 'FRP服务器' },
  { key: '/certificates', icon: <Shield className="h-5 w-5" />, label: '证书管理' },
  { key: '/logs', icon: <FileText className="h-5 w-5" />, label: '日志查看' },
  { key: '/settings', icon: <Settings className="h-5 w-5" />, label: '系统设置' },
];

export function MainLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const userMenuItems: DropdownItem[] = [
    {
      key: 'changePassword',
      icon: <Key className="h-4 w-4" />,
      label: '修改密码',
      onClick: () => setPasswordModalOpen(true),
    },
    {
      key: 'divider',
      label: '',
      divider: true,
    },
    {
      key: 'logout',
      icon: <LogOut className="h-4 w-4" />,
      label: '退出登录',
      danger: true,
      onClick: handleLogout,
    },
  ];

  const handleMenuClick = (key: string) => {
    navigate(key);
    setMobileMenuOpen(false);
  };

  return (
    <div className="flex h-screen overflow-hidden transition-colors duration-300 bg-background">
      {/* 侧边栏 - 桌面端 */}
      <aside
        className={cn(
          'hidden md:flex flex-col border-r transition-all duration-300 ease-in-out',
          'bg-surface-elevated border-border',
          collapsed ? 'w-16' : 'w-56'
        )}
      >
        {/* Logo */}
        <div className="flex items-center h-16 px-4 border-b border-border">
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg shadow-indigo-500/25">
              <Network className="h-5 w-5 text-white" />
            </div>
            {!collapsed && (
              <span className="text-lg font-bold whitespace-nowrap text-foreground">
                FRP Panel
              </span>
            )}
          </div>
        </div>

        {/* 菜单 */}
        <nav className="flex-1 py-4 overflow-y-auto overflow-x-visible">
          <ul className="space-y-1 px-2">
            {menuItems.map((item) => {
              const isActive = location.pathname === item.key;
              const menuButton = (
                <button
                  onClick={() => handleMenuClick(item.key)}
                  className={cn(
                    'flex items-center w-full gap-3 px-3 py-2.5 rounded-lg',
                    'transition-all duration-200',
                    'relative',
                    isActive
                      ? 'bg-indigo-600/20 text-indigo-400'
                      : 'text-foreground-muted hover:text-foreground hover:bg-surface-hover'
                  )}
                >
                  {/* 活动指示器 */}
                  {isActive && (
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 bg-indigo-500 rounded-r-full" />
                  )}
                  <span className={cn(isActive && 'text-indigo-400')}>
                    {item.icon}
                  </span>
                  {!collapsed && (
                    <span className="text-sm font-medium whitespace-nowrap">
                      {item.label}
                    </span>
                  )}
                </button>
              );

              return (
                <li key={item.key}>
                  {collapsed ? (
                    <Tooltip content={item.label} position="right" delay={100}>
                      {menuButton}
                    </Tooltip>
                  ) : (
                    menuButton
                  )}
                </li>
              );
            })}
          </ul>
        </nav>

        {/* 折叠按钮 */}
        <div className="p-4 border-t border-border">
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="flex items-center justify-center w-full h-10 rounded-lg transition-colors text-foreground-muted hover:text-foreground hover:bg-surface-hover"
          >
            {collapsed ? (
              <ChevronRight className="h-5 w-5" />
            ) : (
              <ChevronLeft className="h-5 w-5" />
            )}
          </button>
        </div>
      </aside>

      {/* 移动端侧边栏遮罩 */}
      {mobileMenuOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 md:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* 移动端侧边栏 */}
      <aside
        className={cn(
          'fixed inset-y-0 left-0 w-56 border-r z-50 md:hidden',
          'transform transition-transform duration-300 ease-in-out',
          'bg-surface-elevated border-border',
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        {/* Logo */}
        <div className="flex items-center justify-between h-16 px-4 border-b border-border">
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600">
              <Network className="h-5 w-5 text-white" />
            </div>
            <span className="text-lg font-bold text-foreground">FRP Panel</span>
          </div>
          <button
            onClick={() => setMobileMenuOpen(false)}
            className="p-2 text-foreground-muted hover:text-foreground"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* 菜单 */}
        <nav className="py-4 overflow-y-auto">
          <ul className="space-y-1 px-2">
            {menuItems.map((item) => {
              const isActive = location.pathname === item.key;
              return (
                <li key={item.key}>
                  <button
                    onClick={() => handleMenuClick(item.key)}
                    className={cn(
                      'flex items-center w-full gap-3 px-3 py-2.5 rounded-lg',
                      'transition-all duration-200',
                      isActive
                        ? 'bg-indigo-600/20 text-indigo-400'
                        : 'text-foreground-muted hover:text-foreground hover:bg-surface-hover'
                    )}
                  >
                    {item.icon}
                    <span className="text-sm font-medium">{item.label}</span>
                  </button>
                </li>
              );
            })}
          </ul>
        </nav>
      </aside>

      {/* 主内容区 */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* 顶部导航 */}
        <header className="flex items-center justify-between h-16 px-4 md:px-6 border-b bg-surface-elevated border-border">
          {/* 移动端菜单按钮 */}
          <button
            onClick={() => setMobileMenuOpen(true)}
            className="p-2 md:hidden text-foreground-muted hover:text-foreground"
          >
            <Menu className="h-6 w-6" />
          </button>

          {/* 页面标题 */}
          <div className="hidden md:block">
            <h1 className="text-lg font-semibold text-foreground">
              {menuItems.find((item) => item.key === location.pathname)?.label || ''}
            </h1>
          </div>

          {/* 用户菜单 */}
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <Dropdown
              trigger={
                <div className="flex items-center gap-2 px-3 py-2 rounded-lg transition-colors cursor-pointer hover:bg-surface-hover">
                  <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600">
                    <User className="h-4 w-4 text-white" />
                  </div>
                  <span className="text-sm font-medium hidden sm:block text-foreground-secondary">
                    {user?.username || '用户'}
                  </span>
                </div>
              }
              items={userMenuItems}
            />
          </div>
        </header>

        {/* 内容区 */}
        <main className="flex-1 overflow-auto p-4 md:p-6">
          <div className="animate-fade-in">
            <Outlet />
          </div>
        </main>
      </div>

      {/* 修改密码弹窗 */}
      <ChangePasswordModal
        open={passwordModalOpen}
        onClose={() => setPasswordModalOpen(false)}
      />
    </div>
  );
}