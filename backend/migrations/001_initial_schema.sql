-- ============================================================================
-- FRP Web 管理系统 - 数据库初始化脚本
-- 版本: 1.0.0
-- 数据库: SQLite (GORM AutoMigrate 会自动处理表创建)
-- 说明: 此文件仅作为数据库结构参考文档，实际表结构由 GORM 自动迁移
-- ============================================================================

-- ============================================================================
-- 用户表 (users)
-- 用途: 存储系统用户信息
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    role VARCHAR(20) DEFAULT 'admin',
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- 系统设置表 (settings)
-- 用途: 存储系统配置项
-- ============================================================================
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT,
    description TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- GitHub镜像表 (github_mirrors)
-- 用途: 存储 GitHub 下载镜像配置
-- ============================================================================
CREATE TABLE IF NOT EXISTS github_mirrors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    description TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- FRP服务器表 (frp_servers)
-- 用途: 存储 frps 服务器配置信息
-- ============================================================================
CREATE TABLE IF NOT EXISTS frp_servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    server_type VARCHAR(20) DEFAULT 'local',        -- local: 本地服务器, remote: 远程服务器
    host VARCHAR(255) NOT NULL,
    dashboard_port INTEGER NOT NULL DEFAULT 7500,
    dashboard_user VARCHAR(100),
    dashboard_pwd VARCHAR(255),
    bind_port INTEGER DEFAULT 7000,
    token VARCHAR(64),
    ssh_host VARCHAR(255),                          -- 远程服务器 SSH 地址
    ssh_port INTEGER DEFAULT 22,
    ssh_user VARCHAR(100),
    ssh_password VARCHAR(500),
    install_path VARCHAR(500) DEFAULT '/opt/frps',
    mirror_id INTEGER,
    enabled BOOLEAN DEFAULT TRUE,
    status VARCHAR(20) DEFAULT 'stopped',           -- stopped/starting/running/stopping/error
    pid INTEGER DEFAULT 0,
    version VARCHAR(50),
    binary_path VARCHAR(500),
    config_path VARCHAR(500),
    last_sync_time DATETIME,
    last_error TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- 客户端表 (clients)
-- 用途: 存储 frpc 客户端配置信息
-- ============================================================================
CREATE TABLE IF NOT EXISTS clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    remark TEXT,
    server_addr VARCHAR(255) NOT NULL,
    server_port INTEGER NOT NULL,
    token VARCHAR(255),
    protocol VARCHAR(10) DEFAULT 'tcp',
    frpc_admin_host VARCHAR(255),
    frpc_admin_port INTEGER,
    frpc_admin_user VARCHAR(100),
    frpc_admin_pwd VARCHAR(255),
    frp_server_id INTEGER,
    online_status VARCHAR(20) DEFAULT 'unknown',
    last_heartbeat DATETIME,
    -- 配置同步字段
    config_version INTEGER DEFAULT 1,               -- 配置版本号，每次代理配置变更时自增
    ws_connected BOOLEAN DEFAULT FALSE,             -- WebSocket 连接状态
    last_config_sync DATETIME,                      -- 最后配置同步时间
    config_sync_status VARCHAR(20) DEFAULT 'pending', -- 同步状态: synced/failed/pending/rolled_back
    config_sync_error TEXT,                         -- 同步错误信息
    config_sync_time DATETIME,                      -- 最后同步时间
    -- 版本信息字段
    frpc_version VARCHAR(50),                       -- frpc 版本
    daemon_version VARCHAR(50),                     -- daemon 版本
    os VARCHAR(20),                                 -- 操作系统
    arch VARCHAR(20),                               -- 系统架构
    created_at DATETIME,
    updated_at DATETIME
);

-- 客户端索引
CREATE INDEX IF NOT EXISTS idx_clients_ws_connected ON clients(ws_connected);
CREATE INDEX IF NOT EXISTS idx_clients_config_version ON clients(config_version);

-- ============================================================================
-- 代理表 (proxies)
-- 用途: 存储代理隧道配置
-- ============================================================================
CREATE TABLE IF NOT EXISTS proxies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,                      -- tcp/udp/http/https/stcp/sudp/xtcp/tcpmux
    enabled BOOLEAN DEFAULT TRUE,                   -- 是否启用
    local_ip VARCHAR(50) DEFAULT '127.0.0.1',
    local_port INTEGER NOT NULL,
    remote_port INTEGER,
    custom_domains TEXT,                            -- 自定义域名，逗号分隔
    subdomain VARCHAR(100),
    locations TEXT,
    host_header_rewrite VARCHAR(200),
    http_user VARCHAR(100),
    http_password VARCHAR(100),
    secret_key VARCHAR(100),
    allow_users TEXT,
    use_encryption BOOLEAN DEFAULT FALSE,
    use_compression BOOLEAN DEFAULT FALSE,
    health_check_type VARCHAR(20),
    health_check_timeout INTEGER,
    health_check_interval INTEGER,
    bandwidth_limit VARCHAR(20),
    bandwidth_limit_mode VARCHAR(10) DEFAULT 'client',
    -- 流量统计字段
    total_bytes_in INTEGER DEFAULT 0,
    total_bytes_out INTEGER DEFAULT 0,
    last_online_time DATETIME,
    current_bytes_in_rate INTEGER DEFAULT 0,
    current_bytes_out_rate INTEGER DEFAULT 0,
    last_traffic_update DATETIME,
    -- FRP 状态字段
    frp_status VARCHAR(20) DEFAULT 'unknown',
    frp_cur_conns INTEGER DEFAULT 0,
    frp_last_start_time DATETIME,
    frp_last_close_time DATETIME,
    -- 插件配置字段
    plugin_type VARCHAR(50),                        -- 插件类型: http_proxy/socks5/static_file/unix_domain_socket/https2http/https2https
    plugin_config TEXT,                             -- 插件配置 JSON
    -- DNS 同步字段
    enable_dns_sync BOOLEAN DEFAULT FALSE,          -- 是否启用 DNS 同步
    dns_provider_id INTEGER,                        -- DNS 提供商 ID
    dns_root_domain VARCHAR(100),                   -- 根域名
    -- 自动证书字段
    auto_cert BOOLEAN DEFAULT FALSE,                -- 是否自动申请证书
    cert_id INTEGER,                                -- 关联的证书 ID
    created_at DATETIME,
    updated_at DATETIME
);

-- 代理索引
CREATE INDEX IF NOT EXISTS idx_proxies_client_id ON proxies(client_id);
CREATE INDEX IF NOT EXISTS idx_proxies_dns_provider_id ON proxies(dns_provider_id);
CREATE INDEX IF NOT EXISTS idx_proxies_cert_id ON proxies(cert_id);

-- ============================================================================
-- 告警规则表 (alert_rules)
-- 用途: 存储告警规则配置
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_type VARCHAR(20) NOT NULL DEFAULT 'proxy', -- proxy/frpc/frps/system
    target_id INTEGER NOT NULL,                     -- 对应 proxy_id/client_id/frp_server_id
    proxy_id INTEGER NOT NULL,                      -- 保留兼容旧数据
    rule_type VARCHAR(20) NOT NULL,                 -- daily/monthly/rate/offline
    threshold_value INTEGER DEFAULT 0,              -- 流量告警阈值
    threshold_unit VARCHAR(10) DEFAULT 'bytes',     -- bytes/MB/GB
    cooldown_minutes INTEGER DEFAULT 60,            -- 告警冷却时间（分钟）
    offline_delay_seconds INTEGER DEFAULT 60,       -- 离线延迟确认时间（秒）
    notify_on_recovery BOOLEAN DEFAULT TRUE,        -- 恢复在线时是否发送通知
    enabled BOOLEAN DEFAULT TRUE,
    notify_recipient_ids VARCHAR(500),              -- 接收人 ID 列表，逗号分隔
    notify_group_ids VARCHAR(500),                  -- 分组 ID 列表，逗号分隔
    notify_webhook VARCHAR(500),
    created_at DATETIME,
    updated_at DATETIME
);

-- 告警规则索引
CREATE INDEX IF NOT EXISTS idx_alert_rules_target_type ON alert_rules(target_type);
CREATE INDEX IF NOT EXISTS idx_alert_rules_target_id ON alert_rules(target_id);

-- ============================================================================
-- 告警日志表 (alert_logs)
-- 用途: 存储告警触发记录
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    target_type VARCHAR(20) NOT NULL DEFAULT 'proxy', -- proxy/frpc/frps/system
    target_id INTEGER NOT NULL,                     -- 对应 proxy_id/client_id/frp_server_id，系统告警为 0
    proxy_id INTEGER NOT NULL,                      -- 保留兼容旧数据
    alert_type VARCHAR(30) NOT NULL,
    current_value INTEGER DEFAULT 0,                -- 流量告警使用
    threshold_value INTEGER DEFAULT 0,
    message TEXT,
    event_data TEXT,                                -- 事件详情 JSON，用于系统告警
    notified BOOLEAN DEFAULT FALSE,
    created_at DATETIME
);

-- 告警日志索引
CREATE INDEX IF NOT EXISTS idx_alert_logs_target_type ON alert_logs(target_type);
CREATE INDEX IF NOT EXISTS idx_alert_logs_target_id ON alert_logs(target_id);

-- ============================================================================
-- 告警接收人表 (alert_recipients)
-- 用途: 存储告警通知接收人
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- 告警接收人分组表 (alert_recipient_groups)
-- 用途: 存储告警接收人分组
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_recipient_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- 分组-接收人关联表 (alert_group_recipients)
-- 用途: 存储分组与接收人的多对多关系
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_group_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    recipient_id INTEGER NOT NULL,
    UNIQUE(group_id, recipient_id)
);

-- ============================================================================
-- 客户端注册令牌表 (client_register_tokens)
-- 用途: 存储客户端一键注册令牌
-- ============================================================================
CREATE TABLE IF NOT EXISTS client_register_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token VARCHAR(64) NOT NULL UNIQUE,
    client_name VARCHAR(100) NOT NULL,
    frp_server_id INTEGER DEFAULT 1,
    server_addr VARCHAR(255) NOT NULL,
    server_port INTEGER NOT NULL,
    token_str VARCHAR(255),
    admin_password VARCHAR(64),
    protocol VARCHAR(10) DEFAULT 'tcp',
    remark TEXT,
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at DATETIME,
    created_by INTEGER,
    created_at DATETIME
);

-- ============================================================================
-- DNS提供商表 (dns_providers)
-- 用途: 存储 DNS 服务提供商配置
-- ============================================================================
CREATE TABLE IF NOT EXISTS dns_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,                      -- aliyun/cloudflare/tencent
    access_key VARCHAR(100),
    secret_key VARCHAR(200),
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME,
    updated_at DATETIME
);

-- ============================================================================
-- DNS记录表 (dns_records)
-- 用途: 存储 DNS 解析记录
-- ============================================================================
CREATE TABLE IF NOT EXISTS dns_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    proxy_id INTEGER,
    provider_id INTEGER,
    domain VARCHAR(255),
    root_domain VARCHAR(100),
    record_type VARCHAR(10) DEFAULT 'A',
    record_value VARCHAR(50),
    record_id VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',           -- pending/synced/failed
    last_error TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- DNS 记录索引
CREATE INDEX IF NOT EXISTS idx_dns_records_proxy_id ON dns_records(proxy_id);

-- ============================================================================
-- SSL证书表 (certificates)
-- 用途: 存储 SSL/TLS 证书
-- ============================================================================
CREATE TABLE IF NOT EXISTS certificates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    proxy_id INTEGER,
    domain VARCHAR(255) NOT NULL,
    provider_id INTEGER,
    status VARCHAR(20) DEFAULT 'pending',           -- pending/active/expiring/expired/failed
    cert_pem TEXT,                                  -- 证书内容 (PEM 格式)
    key_pem TEXT,                                   -- 私钥内容 (PEM 格式)
    issuer_cert_pem TEXT,                           -- 颁发者证书
    not_before DATETIME,                            -- 生效时间
    not_after DATETIME,                             -- 过期时间
    last_error TEXT,
    auto_renew BOOLEAN DEFAULT TRUE,
    acme_account_id VARCHAR(255),
    created_at DATETIME,
    updated_at DATETIME
);

-- 证书索引
CREATE INDEX IF NOT EXISTS idx_certificates_proxy_id ON certificates(proxy_id);
CREATE INDEX IF NOT EXISTS idx_certificates_provider_id ON certificates(provider_id);

-- ============================================================================
-- 操作日志表 (operation_logs)
-- 用途: 存储用户操作日志
-- ============================================================================
CREATE TABLE IF NOT EXISTS operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    operation_type VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id INTEGER,
    description TEXT,
    ip_address VARCHAR(50),
    ip_location VARCHAR(100),                       -- IP 归属地
    created_at DATETIME
);

-- 操作日志索引
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);

-- ============================================================================
-- 服务器指标历史表 (server_metrics_history)
-- 用途: 存储 frps 服务器性能指标历史
-- ============================================================================
CREATE TABLE IF NOT EXISTS server_metrics_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    cpu_percent REAL,
    memory_bytes INTEGER,
    traffic_in INTEGER,
    traffic_out INTEGER,
    record_time DATETIME NOT NULL,
    created_at DATETIME
);

-- 服务器指标索引
CREATE INDEX IF NOT EXISTS idx_server_time ON server_metrics_history(server_id, record_time);

-- ============================================================================
-- 代理指标历史表 (proxy_metrics_history)
-- 用途: 存储代理隧道流量历史
-- ============================================================================
CREATE TABLE IF NOT EXISTS proxy_metrics_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    proxy_name VARCHAR(100) NOT NULL,
    proxy_type VARCHAR(20),
    traffic_in INTEGER,                             -- 累计入站流量
    traffic_out INTEGER,                            -- 累计出站流量
    rate_in INTEGER,                                -- 入站速率 bytes/s
    rate_out INTEGER,                               -- 出站速率 bytes/s
    record_time DATETIME NOT NULL,
    created_at DATETIME
);

-- 代理指标索引
CREATE INDEX IF NOT EXISTS idx_proxy_metrics ON proxy_metrics_history(server_id, proxy_name, record_time);

-- ============================================================================
-- 插件配置 JSON 格式说明
-- ============================================================================
-- 
-- 1. HTTP代理插件 (http_proxy):
--    {"httpUser": "用户名", "httpPassword": "密码"}
--
-- 2. SOCKS5代理插件 (socks5):
--    {"username": "用户名", "password": "密码"}
--
-- 3. 静态文件服务插件 (static_file):
--    {"localPath": "/path/to/files", "stripPrefix": "static", "httpUser": "用户名", "httpPassword": "密码"}
--
-- 4. Unix域套接字插件 (unix_domain_socket):
--    {"unixPath": "/var/run/docker.sock"}
--
-- 5. HTTPS转HTTP插件 (https2http):
--    {"localAddr": "127.0.0.1:8080", "crtPath": "/path/to/cert.pem", "keyPath": "/path/to/key.pem", "hostHeaderRewrite": "example.com"}
--
-- 6. HTTPS转HTTPS插件 (https2https):
--    {"localAddr": "127.0.0.1:8443", "crtPath": "/path/to/cert.pem", "keyPath": "/path/to/key.pem", "hostHeaderRewrite": "example.com"}