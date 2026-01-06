import { Info, Copy, ExternalLink, Lock, AlertTriangle } from 'lucide-react';
import { Modal, Badge, Button, Card, CardHeader, CardContent } from './ui';
import { toast } from './ui/Toast';
import { cn } from '../utils/cn';
import type { Proxy, PluginType } from '../types';
import {
  PLUGIN_TYPE_LABELS,
  PLUGIN_TYPE_COLORS,
  PLUGIN_USAGE_DESCRIPTIONS,
  PLUGIN_USAGE_EXAMPLES,
  PLUGIN_USAGE_EXAMPLES_WITH_AUTH,
} from '../types';

interface PluginUsageModalProps {
  visible: boolean;
  onClose: () => void;
  proxy: Proxy | null;
  serverAddr: string;
}

// 解析插件配置
const parsePluginConfig = (configStr: string | undefined): Record<string, string> => {
  if (!configStr) return {};
  try {
    return JSON.parse(configStr);
  } catch {
    return {};
  }
};

// 替换模板变量
const replaceTemplateVars = (
  template: string,
  vars: Record<string, string>
): string => {
  let result = template;
  Object.entries(vars).forEach(([key, value]) => {
    result = result.replace(new RegExp(`\\{${key}\\}`, 'g'), value);
  });
  return result;
};

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    toast.success('已复制到剪贴板');
  }).catch(() => {
    toast.error('复制失败');
  });
};

// 插件颜色映射到 Badge variant
const getPluginBadgeVariant = (pluginType: PluginType): 'default' | 'success' | 'warning' | 'danger' | 'info' => {
  const colorMap: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'info'> = {
    blue: 'info',
    green: 'success',
    orange: 'warning',
    red: 'danger',
    purple: 'default',
    cyan: 'info',
  };
  return colorMap[PLUGIN_TYPE_COLORS[pluginType]] || 'default';
};

export function PluginUsageModal({ visible, onClose, proxy, serverAddr }: PluginUsageModalProps) {
  if (!proxy || !proxy.plugin_type) {
    return null;
  }

  const pluginType = proxy.plugin_type as PluginType;
  const pluginConfig = parsePluginConfig(proxy.plugin_config);
  const port = proxy.remote_port?.toString() || '未分配';
  const hasPort = !!proxy.remote_port;

  // 判断是否有认证信息
  const hasAuth = (() => {
    switch (pluginType) {
      case 'http_proxy':
        return !!(pluginConfig.httpUser && pluginConfig.httpPassword);
      case 'socks5':
        return !!(pluginConfig.username && pluginConfig.password);
      case 'static_file':
        return !!(pluginConfig.httpUser && pluginConfig.httpPassword);
      default:
        return false;
    }
  })();

  // 获取认证信息
  const getAuthInfo = () => {
    switch (pluginType) {
      case 'http_proxy':
      case 'static_file':
        return {
          user: pluginConfig.httpUser || '',
          password: pluginConfig.httpPassword || '',
        };
      case 'socks5':
        return {
          user: pluginConfig.username || '',
          password: pluginConfig.password || '',
        };
      default:
        return { user: '', password: '' };
    }
  };

  const authInfo = getAuthInfo();

  // 模板变量
  const templateVars: Record<string, string> = {
    server: serverAddr,
    port: port,
    user: authInfo.user,
    password: authInfo.password,
  };

  // 获取连接地址
  const getConnectionUrl = (): string => {
    if (!hasPort) return '端口未分配';
    
    switch (pluginType) {
      case 'http_proxy':
        return hasAuth
          ? `http://${authInfo.user}:${authInfo.password}@${serverAddr}:${port}`
          : `http://${serverAddr}:${port}`;
      case 'socks5':
        return hasAuth
          ? `socks5://${authInfo.user}:${authInfo.password}@${serverAddr}:${port}`
          : `socks5://${serverAddr}:${port}`;
      case 'static_file': {
        const stripPrefix = pluginConfig.stripPrefix;
        const path = stripPrefix ? `/${stripPrefix}/` : '/';
        return `http://${serverAddr}:${port}${path}`;
      }
      case 'unix_domain_socket':
        return `tcp://${serverAddr}:${port}`;
      default:
        return `${serverAddr}:${port}`;
    }
  };

  // 获取使用示例 - 根据是否有认证选择对应的示例集
  const getExamples = () => {
    if (hasAuth) {
      return PLUGIN_USAGE_EXAMPLES_WITH_AUTH[pluginType] || [];
    } else {
      return PLUGIN_USAGE_EXAMPLES[pluginType] || [];
    }
  };

  const connectionUrl = getConnectionUrl();
  const examples = getExamples();

  return (
    <Modal
      open={visible}
      onClose={onClose}
      title={
        <div className="flex items-center gap-2">
          <Badge variant={getPluginBadgeVariant(pluginType)}>
            {PLUGIN_TYPE_LABELS[pluginType]}
          </Badge>
          <span>使用说明</span>
        </div>
      }
      size="lg"
    >
      <div className="space-y-4">
        {/* 功能描述 */}
        <div className="flex items-start gap-3 rounded-lg border border-blue-500/30 bg-blue-500/10 p-4">
          <Info className="mt-0.5 h-5 w-5 flex-shrink-0 text-blue-400" />
          <div>
            <div className="font-medium text-blue-400">功能说明</div>
            <div className="mt-1 text-sm text-foreground-secondary">{PLUGIN_USAGE_DESCRIPTIONS[pluginType]}</div>
          </div>
        </div>

        {/* 连接信息 */}
        <Card>
          <CardHeader>连接信息</CardHeader>
          <CardContent>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-foreground-muted">代理名称</span>
                <span className="font-medium text-foreground">{proxy.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-foreground-muted">服务器地址</span>
                <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">{serverAddr}</code>
              </div>
              <div className="flex justify-between">
                <span className="text-foreground-muted">远程端口</span>
                {hasPort ? (
                  <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">{port}</code>
                ) : (
                  <span className="text-yellow-400">未分配</span>
                )}
              </div>
              <div className="flex items-center justify-between">
                <span className="text-foreground-muted">连接地址</span>
                <div className="flex items-center gap-2">
                  <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">{connectionUrl}</code>
                  {hasPort && (
                    <>
                      <button
                        onClick={() => copyToClipboard(connectionUrl)}
                        className="rounded p-1 text-foreground-muted hover:bg-surface-hover hover:text-foreground"
                        title="复制"
                      >
                        <Copy className="h-4 w-4" />
                      </button>
                      {pluginType === 'static_file' && (
                        <button
                          onClick={() => window.open(connectionUrl, '_blank')}
                          className="rounded p-1 text-foreground-muted hover:bg-surface-hover hover:text-foreground"
                          title="打开"
                        >
                          <ExternalLink className="h-4 w-4" />
                        </button>
                      )}
                    </>
                  )}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 认证信息 */}
        {hasAuth && (
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Lock className="h-4 w-4 text-yellow-400" />
                认证信息
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-foreground-muted">用户名</span>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">{authInfo.user}</code>
                    <button
                      onClick={() => copyToClipboard(authInfo.user)}
                      className="rounded p-1 text-foreground-muted hover:bg-surface-hover hover:text-foreground"
                    >
                      <Copy className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>
                <div>
                  <span className="text-foreground-muted">密码</span>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">{authInfo.password}</code>
                    <button
                      onClick={() => copyToClipboard(authInfo.password)}
                      className="rounded p-1 text-foreground-muted hover:bg-surface-hover hover:text-foreground"
                    >
                      <Copy className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* 插件特定配置 */}
        {pluginType === 'static_file' && (
          <Card>
            <CardHeader>静态文件配置</CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-foreground-muted">本地路径</span>
                  <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">
                    {pluginConfig.localPath || '未配置'}
                  </code>
                </div>
                {pluginConfig.stripPrefix && (
                  <div className="flex justify-between">
                    <span className="text-foreground-muted">URL前缀</span>
                    <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">
                      /{pluginConfig.stripPrefix}/
                    </code>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {pluginType === 'unix_domain_socket' && (
          <Card>
            <CardHeader>Unix套接字配置</CardHeader>
            <CardContent>
              <div className="flex justify-between text-sm">
                <span className="text-foreground-muted">套接字路径</span>
                <code className="rounded px-2 py-0.5 bg-surface-hover text-indigo-400">
                  {pluginConfig.unixPath || '未配置'}
                </code>
              </div>
            </CardContent>
          </Card>
        )}

        {/* 使用示例 */}
        {hasPort && examples.length > 0 && (
          <div>
            <div className="mb-3 flex items-center gap-2 text-sm font-medium text-foreground-secondary">
              <div className="h-px flex-1 bg-border" />
              <span>命令行示例</span>
              <div className="h-px flex-1 bg-border" />
            </div>
            <div className="space-y-3">
              {examples.map((example, index) => {
                const command = replaceTemplateVars(example.command, templateVars);
                return (
                  <Card key={index}>
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <span className="font-medium">{example.title}</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => copyToClipboard(command)}
                        >
                          <Copy className="mr-1 h-3.5 w-3.5" />
                          复制
                        </Button>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <p className="mb-2 text-sm text-foreground-muted">{example.description}</p>
                      <pre className={cn("overflow-auto rounded-lg p-3 text-sm", "bg-surface text-green-400")}>
                        {command}
                      </pre>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </div>
        )}

        {/* 端口未分配提示 */}
        {!hasPort && (
          <div className="flex items-start gap-3 rounded-lg border border-yellow-500/30 bg-yellow-500/10 p-4">
            <AlertTriangle className="mt-0.5 h-5 w-5 flex-shrink-0 text-yellow-400" />
            <div>
              <div className="font-medium text-yellow-400">端口未分配</div>
              <div className="mt-1 text-sm text-foreground-secondary">
                该代理的远程端口尚未分配，请确保代理已启用并且 frpc 客户端已连接到服务器。
              </div>
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
}

export default PluginUsageModal;