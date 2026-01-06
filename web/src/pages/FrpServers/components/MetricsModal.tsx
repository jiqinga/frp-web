import { useState } from 'react';
import { Tab } from '@headlessui/react';
import {
  Cpu,
  HardDrive,
  Clock,
  ArrowDownToLine,
  ArrowUpFromLine,
  Activity,
  TrendingUp,
  Loader2
} from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Button } from '../../../components/ui/Button';
import type { FrpsMetrics } from '../../../api/frpServer';
import ServerMetricsChart from '../../../components/ServerMetricsChart';

interface MetricsModalProps {
  visible: boolean;
  loading: boolean;
  metrics: FrpsMetrics | null;
  serverId: number | null;
  onClose: () => void;
}

/**
 * 服务器指标查看弹窗
 * 显示实时指标和历史趋势图表
 */
export function MetricsModal({
  visible,
  loading,
  metrics,
  serverId,
  onClose,
}: MetricsModalProps) {
  const [selectedTab, setSelectedTab] = useState(0);

  // 格式化字节大小
  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // 格式化运行时长
  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}天 ${hours}小时 ${mins}分钟`;
    if (hours > 0) return `${hours}小时 ${mins}分钟`;
    return `${mins}分钟`;
  };

  // 计算 CPU 使用率
  const getCpuUsage = () => {
    if (!metrics || metrics.uptime <= 0) return '0.00';
    return ((metrics.cpu_seconds / metrics.uptime) * 100).toFixed(2);
  };

  // 统计卡片组件
  const StatCard = ({ 
    icon: Icon, 
    label, 
    value, 
    color = 'indigo' 
  }: { 
    icon: React.ElementType; 
    label: string; 
    value: string; 
    color?: 'indigo' | 'emerald' | 'amber' | 'cyan' | 'purple';
  }) => {
    const colorClasses = {
      indigo: 'from-indigo-500/20 to-indigo-600/10 border-indigo-500/30 text-indigo-400',
      emerald: 'from-emerald-500/20 to-emerald-600/10 border-emerald-500/30 text-emerald-400',
      amber: 'from-amber-500/20 to-amber-600/10 border-amber-500/30 text-amber-400',
      cyan: 'from-cyan-500/20 to-cyan-600/10 border-cyan-500/30 text-cyan-400',
      purple: 'from-purple-500/20 to-purple-600/10 border-purple-500/30 text-purple-400',
    };

    return (
      <div className={`
        relative overflow-hidden rounded-lg border p-4
        bg-gradient-to-br ${colorClasses[color]}
        transition-all duration-300 hover:scale-[1.02]
      `}>
        {/* 背景装饰 */}
        <div className="absolute top-0 right-0 w-20 h-20 opacity-10">
          <Icon className="w-full h-full" />
        </div>
        
        <div className="relative z-10">
          <div className="flex items-center gap-2 mb-2">
            <Icon className="w-4 h-4" />
            <span className="text-xs text-foreground-muted">{label}</span>
          </div>
          <div className="text-xl font-bold text-foreground">{value}</div>
        </div>
        
        {/* 底部发光线 */}
        <div className={`absolute bottom-0 left-0 right-0 h-0.5 bg-gradient-to-r ${colorClasses[color].split(' ')[0].replace('/20', '/50')} to-transparent`} />
      </div>
    );
  };

  // 分隔线组件
  const Divider = ({ children }: { children: React.ReactNode }) => (
    <div className="relative flex items-center py-4">
      <div className="flex-grow border-t border-border" />
      <span className="flex-shrink mx-4 text-sm text-foreground-muted flex items-center gap-2">
        {children}
      </span>
      <div className="flex-grow border-t border-border" />
    </div>
  );

  return (
    <Modal
      open={visible}
      onClose={onClose}
      title={
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-indigo-400" />
          <span>服务器运行指标</span>
        </div>
      }
      size="xl"
      footer={
        <div className="flex justify-end">
          <Button variant="secondary" onClick={onClose}>
            关闭
          </Button>
        </div>
      }
    >
      {loading ? (
        <div className="flex flex-col items-center justify-center py-16">
          <Loader2 className="w-10 h-10 text-indigo-400 animate-spin mb-4" />
          <span className="text-foreground-muted">加载指标数据中...</span>
        </div>
      ) : metrics ? (
        <Tab.Group selectedIndex={selectedTab} onChange={setSelectedTab}>
          <Tab.List className="flex space-x-1 rounded-lg p-1 mb-6 bg-surface-hover">
            <Tab
              className={({ selected }) =>
                `w-full rounded-md py-2.5 text-sm font-medium leading-5 transition-all duration-200
                 flex items-center justify-center gap-2
                 ${selected
                   ? 'bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 shadow-lg shadow-indigo-500/10'
                   : 'text-foreground-muted hover:bg-surface-active hover:text-foreground'
                 }`
              }
            >
              <Activity className="w-4 h-4" />
              实时指标
            </Tab>
            <Tab
              className={({ selected }) =>
                `w-full rounded-md py-2.5 text-sm font-medium leading-5 transition-all duration-200
                 flex items-center justify-center gap-2
                 ${selected
                   ? 'bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 shadow-lg shadow-indigo-500/10'
                   : 'text-foreground-muted hover:bg-surface-active hover:text-foreground'
                 }`
              }
            >
              <TrendingUp className="w-4 h-4" />
              历史趋势
            </Tab>
          </Tab.List>

          <Tab.Panels>
            {/* 实时指标面板 */}
            <Tab.Panel>
              <Divider>
                <Cpu className="w-4 h-4" />
                系统资源
              </Divider>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                <StatCard
                  icon={Cpu}
                  label="平均 CPU 使用率"
                  value={`${getCpuUsage()}%`}
                  color="indigo"
                />
                <StatCard
                  icon={HardDrive}
                  label="内存占用"
                  value={formatBytes(metrics.memory_bytes)}
                  color="emerald"
                />
                <StatCard
                  icon={Clock}
                  label="运行时长"
                  value={formatUptime(metrics.uptime)}
                  color="amber"
                />
              </div>

              <Divider>
                <Activity className="w-4 h-4" />
                流量统计
              </Divider>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <StatCard
                  icon={ArrowDownToLine}
                  label="入站流量"
                  value={formatBytes(metrics.traffic_in)}
                  color="cyan"
                />
                <StatCard
                  icon={ArrowUpFromLine}
                  label="出站流量"
                  value={formatBytes(metrics.traffic_out)}
                  color="purple"
                />
              </div>
            </Tab.Panel>

            {/* 历史趋势面板 */}
            <Tab.Panel>
              {serverId ? (
                <div className="min-h-[400px]">
                  <ServerMetricsChart serverId={serverId} />
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-16 text-foreground-muted">
                  <TrendingUp className="w-12 h-12 mb-4 opacity-50" />
                  <span>无法加载历史数据</span>
                </div>
              )}
            </Tab.Panel>
          </Tab.Panels>
        </Tab.Group>
      ) : (
        <div className="flex flex-col items-center justify-center py-16 text-foreground-muted">
          <Activity className="w-12 h-12 mb-4 opacity-50" />
          <span>暂无指标数据</span>
        </div>
      )}
    </Modal>
  );
}