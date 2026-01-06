// 插件类型枚举
export type PluginType = 'http_proxy' | 'socks5' | 'static_file' | 'unix_domain_socket' | 'https2http' | 'https2https';

// 插件类型标签
export const PLUGIN_TYPE_LABELS: Record<PluginType, string> = {
  http_proxy: 'HTTP代理',
  socks5: 'SOCKS5代理',
  static_file: '静态文件',
  unix_domain_socket: 'Unix套接字',
  https2http: 'HTTPS转HTTP',
  https2https: 'HTTPS转HTTPS',
};

// 插件类型颜色
export const PLUGIN_TYPE_COLORS: Record<PluginType, string> = {
  http_proxy: 'magenta',
  socks5: 'cyan',
  static_file: 'lime',
  unix_domain_socket: 'gold',
  https2http: 'red',
  https2https: 'orange',
};

// HTTP代理插件配置
export interface HTTPProxyPluginConfig {
  httpUser?: string;
  httpPassword?: string;
}

// SOCKS5代理插件配置
export interface Socks5PluginConfig {
  username?: string;
  password?: string;
}

// 静态文件插件配置
export interface StaticFilePluginConfig {
  localPath: string;
  stripPrefix?: string;
  httpUser?: string;
  httpPassword?: string;
}

// Unix域套接字插件配置
export interface UnixDomainSocketPluginConfig {
  unixPath: string;
}

// HTTPS转HTTP插件配置
export interface HTTPS2HTTPPluginConfig {
  localAddr: string;
  crtPath: string;
  keyPath: string;
  hostHeaderRewrite?: string;
}

// HTTPS转HTTPS插件配置
export interface HTTPS2HTTPSPluginConfig {
  localAddr: string;
  crtPath: string;
  keyPath: string;
  hostHeaderRewrite?: string;
}

// 插件配置联合类型
export type PluginConfig = HTTPProxyPluginConfig | Socks5PluginConfig | StaticFilePluginConfig | UnixDomainSocketPluginConfig | HTTPS2HTTPPluginConfig | HTTPS2HTTPSPluginConfig;

// 插件字段配置
export const PLUGIN_TYPE_FIELDS: Record<PluginType, string[]> = {
  http_proxy: ['httpUser', 'httpPassword'],
  socks5: ['username', 'password'],
  static_file: ['localPath', 'stripPrefix', 'httpUser', 'httpPassword'],
  unix_domain_socket: ['unixPath'],
  https2http: ['localAddr', 'crtPath', 'keyPath', 'hostHeaderRewrite'],
  https2https: ['localAddr', 'crtPath', 'keyPath', 'hostHeaderRewrite'],
};

// 插件必填字段
export const PLUGIN_REQUIRED_FIELDS: Record<PluginType, string[]> = {
  http_proxy: [],
  socks5: [],
  static_file: ['localPath'],
  unix_domain_socket: ['unixPath'],
  https2http: ['localAddr', 'crtPath', 'keyPath'],
  https2https: ['localAddr', 'crtPath', 'keyPath'],
};

// 插件字段标签
export const PLUGIN_FIELD_LABELS: Record<string, string> = {
  httpUser: 'HTTP用户名',
  httpPassword: 'HTTP密码',
  username: '用户名',
  password: '密码',
  localPath: '本地路径',
  stripPrefix: 'URL前缀剥离',
  unixPath: 'Unix套接字路径',
  localAddr: '本地服务地址',
  crtPath: '证书文件路径',
  keyPath: '私钥文件路径',
  hostHeaderRewrite: 'Host Header重写',
};

// 插件字段提示
export const PLUGIN_FIELD_TOOLTIPS: Record<string, string> = {
  httpUser: 'HTTP代理认证用户名，留空表示不启用认证',
  httpPassword: 'HTTP代理认证密码',
  username: 'SOCKS5代理认证用户名，留空表示不启用认证',
  password: 'SOCKS5代理认证密码',
  localPath: '要共享的本地文件目录路径',
  stripPrefix: '从URL中剥离的前缀，如 static',
  unixPath: 'Unix域套接字文件路径，如 /var/run/docker.sock',
  localAddr: '本地HTTP/HTTPS服务地址，格式为 IP:端口',
  crtPath: 'TLS证书文件的绝对路径，如 /etc/ssl/certs/server.crt',
  keyPath: 'TLS私钥文件的绝对路径，如 /etc/ssl/private/server.key',
  hostHeaderRewrite: '重写发送到后端服务的Host头，留空则保持原样',
};

// 插件字段占位符
export const PLUGIN_FIELD_PLACEHOLDERS: Record<string, string> = {
  httpUser: 'admin',
  httpPassword: '请输入密码',
  username: 'user',
  password: '请输入密码',
  localPath: '/tmp/shared_files',
  stripPrefix: 'static',
  unixPath: '/var/run/docker.sock',
  localAddr: '127.0.0.1:8080',
  crtPath: '/etc/ssl/certs/server.crt',
  keyPath: '/etc/ssl/private/server.key',
  hostHeaderRewrite: 'localhost',
};

// 插件使用说明描述
export const PLUGIN_USAGE_DESCRIPTIONS: Record<PluginType, string> = {
  http_proxy: '提供 HTTP 代理服务，可用于浏览器或应用程序的 HTTP 代理设置。支持 HTTP/HTTPS 流量转发。',
  socks5: '提供 SOCKS5 代理服务，支持 TCP/UDP 流量转发。适用于需要全局代理或特定应用代理的场景。',
  static_file: '通过 HTTP 协议提供静态文件服务，可用于共享本地文件目录。支持 HTTP Basic 认证保护。',
  unix_domain_socket: '将 Unix 域套接字代理为 TCP 端口，可用于远程访问 Docker、MySQL 等使用 Unix 套接字的服务。',
  https2http: '在 frpc 端进行 TLS 终止，将外部 HTTPS 请求转换为 HTTP 请求转发到本地服务。适用于本地服务是 HTTP 但需要对外提供 HTTPS 访问的场景。',
  https2https: '在 frpc 端进行 TLS 终止后，再以 HTTPS 方式转发到本地 HTTPS 服务。适用于本地服务也是 HTTPS 的场景。',
};

// 插件使用示例模板
export interface PluginUsageExample {
  title: string;
  command: string;
  description: string;
}

// 插件使用示例模板
export const PLUGIN_USAGE_EXAMPLES: Record<PluginType, PluginUsageExample[]> = {
  http_proxy: [
    { title: 'curl 命令', command: 'curl -x http://{server}:{port} http://example.com', description: '使用 curl 通过 HTTP 代理访问网站' },
    { title: 'Linux/Mac 环境变量', command: 'export http_proxy=http://{server}:{port}\nexport https_proxy=http://{server}:{port}', description: '设置系统环境变量，使所有支持代理的程序使用此代理' },
    { title: 'Windows 环境变量', command: 'set http_proxy=http://{server}:{port}\nset https_proxy=http://{server}:{port}', description: '在 Windows CMD 中设置代理环境变量' },
    { title: 'wget 命令', command: 'wget -e http_proxy={server}:{port} http://example.com', description: '使用 wget 通过 HTTP 代理下载文件' },
  ],
  socks5: [
    { title: 'curl 命令', command: 'curl --socks5 {server}:{port} http://example.com', description: '使用 curl 通过 SOCKS5 代理访问网站' },
    { title: 'SSH ProxyCommand', command: 'ssh -o ProxyCommand="nc -x {server}:{port} %h %p" user@target-host', description: '通过 SOCKS5 代理进行 SSH 连接' },
    { title: 'Git 配置', command: 'git config --global http.proxy socks5://{server}:{port}', description: '配置 Git 使用 SOCKS5 代理' },
    { title: 'Chrome 启动参数', command: 'chrome --proxy-server="socks5://{server}:{port}"', description: '启动 Chrome 浏览器并使用 SOCKS5 代理' },
  ],
  static_file: [
    { title: '浏览器访问', command: 'http://{server}:{port}/', description: '在浏览器中直接访问静态文件服务' },
    { title: 'curl 下载', command: 'curl -O http://{server}:{port}/filename.txt', description: '使用 curl 下载文件' },
    { title: 'wget 下载', command: 'wget http://{server}:{port}/filename.txt', description: '使用 wget 下载文件' },
  ],
  unix_domain_socket: [
    { title: 'Docker 远程访问', command: 'docker -H tcp://{server}:{port} ps', description: '通过 TCP 端口远程访问 Docker 守护进程' },
    { title: 'Docker 环境变量', command: 'export DOCKER_HOST=tcp://{server}:{port}', description: '设置 DOCKER_HOST 环境变量以使用远程 Docker' },
    { title: 'curl 访问 Docker API', command: 'curl http://{server}:{port}/v1.24/containers/json', description: '使用 curl 直接调用 Docker API' },
  ],
  https2http: [
    { title: '浏览器访问', command: 'https://{domain}/', description: '通过 HTTPS 访问您的服务' },
    { title: 'curl 命令', command: 'curl https://{domain}/', description: '使用 curl 通过 HTTPS 访问服务' },
  ],
  https2https: [
    { title: '浏览器访问', command: 'https://{domain}/', description: '通过 HTTPS 访问您的服务' },
    { title: 'curl 命令', command: 'curl https://{domain}/', description: '使用 curl 通过 HTTPS 访问服务' },
  ],
};

// 带认证的插件使用示例模板
export const PLUGIN_USAGE_EXAMPLES_WITH_AUTH: Record<PluginType, PluginUsageExample[]> = {
  http_proxy: [
    { title: 'curl 命令', command: 'curl -x http://{user}:{password}@{server}:{port} http://example.com', description: '使用 curl 通过 HTTP 代理访问网站' },
    { title: 'Linux/Mac 环境变量', command: 'export http_proxy=http://{user}:{password}@{server}:{port}\nexport https_proxy=http://{user}:{password}@{server}:{port}', description: '设置系统环境变量，使所有支持代理的程序使用此代理' },
    { title: 'Windows 环境变量', command: 'set http_proxy=http://{user}:{password}@{server}:{port}\nset https_proxy=http://{user}:{password}@{server}:{port}', description: '在 Windows CMD 中设置代理环境变量' },
    { title: 'wget 命令', command: 'wget -e http_proxy=http://{user}:{password}@{server}:{port} http://example.com', description: '使用 wget 通过 HTTP 代理下载文件' },
  ],
  socks5: [
    { title: 'curl 命令', command: 'curl --socks5 {server}:{port} --proxy-user {user}:{password} http://example.com', description: '使用 curl 通过 SOCKS5 代理访问网站' },
    { title: 'SSH ProxyCommand', command: 'ssh -o ProxyCommand="nc -X 5 -x {server}:{port} -P {user} %h %p" user@target-host', description: '通过 SOCKS5 代理进行 SSH 连接（需要输入密码）' },
    { title: 'Git 配置', command: 'git config --global http.proxy socks5://{user}:{password}@{server}:{port}', description: '配置 Git 使用 SOCKS5 代理' },
    { title: 'Chrome 启动参数', command: 'chrome --proxy-server="socks5://{server}:{port}"', description: '启动 Chrome 浏览器并使用 SOCKS5 代理（浏览器会提示输入认证信息）' },
  ],
  static_file: [
    { title: '浏览器访问', command: 'http://{server}:{port}/', description: '在浏览器中访问静态文件服务（浏览器会弹出认证对话框）' },
    { title: 'curl 下载', command: 'curl -u {user}:{password} -O http://{server}:{port}/filename.txt', description: '使用 curl 下载文件' },
    { title: 'wget 下载', command: 'wget --http-user={user} --http-password={password} http://{server}:{port}/filename.txt', description: '使用 wget 下载文件' },
  ],
  unix_domain_socket: [],
  https2http: [],
  https2https: [],
};