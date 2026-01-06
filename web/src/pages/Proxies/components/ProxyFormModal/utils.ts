// 从 custom_domains 和 root_domain 反向解析出前缀
export const extractPrefixesFromDomains = (customDomains: string, rootDomain: string): string => {
  if (!customDomains || !rootDomain) return '';
  const domains = customDomains.split(',').map(d => d.trim()).filter(Boolean);
  const suffix = `.${rootDomain}`;
  const prefixes = domains
    .filter(d => d.endsWith(suffix))
    .map(d => d.slice(0, -suffix.length));
  return prefixes.join(', ');
};

// 从前缀和根域名生成完整域名
export const generateDomainsFromPrefixes = (prefixes: string, rootDomain: string): string => {
  if (!prefixes || !rootDomain) return '';
  return prefixes
    .split(',')
    .map(p => p.trim())
    .filter(Boolean)
    .map(p => `${p}.${rootDomain}`)
    .join(',');
};

// 验证域名前缀格式
export const validateDomainPrefix = (prefix: string): boolean => {
  if (!prefix) return true;
  const pattern = /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/;
  return pattern.test(prefix);
};

// 从证书域名解析根域名和前缀
export const parseCertificateDomain = (domain: string): { rootDomain: string; prefix: string; isWildcard: boolean } => {
  if (domain.startsWith('*.')) {
    return { rootDomain: domain.slice(2), prefix: '', isWildcard: true };
  }
  const parts = domain.split('.');
  if (parts.length >= 2) {
    return {
      rootDomain: parts.slice(1).join('.'),
      prefix: parts[0],
      isWildcard: false
    };
  }
  return { rootDomain: domain, prefix: '', isWildcard: false };
};

// 检查域名是否与证书匹配
export const checkCertificateMatch = (
  certDomain: string,
  rootDomain: string,
  prefixes: string
): { matches: boolean; warning?: string } => {
  const { rootDomain: certRoot, prefix: certPrefix, isWildcard } = parseCertificateDomain(certDomain);
  
  if (certRoot !== rootDomain) {
    return { matches: false, warning: `证书域名 ${certDomain} 与根域名 ${rootDomain} 不匹配` };
  }
  
  if (isWildcard) {
    return { matches: true };
  }
  
  const prefixList = prefixes.split(',').map(p => p.trim()).filter(Boolean);
  if (prefixList.length === 0) {
    return { matches: false, warning: `证书 ${certDomain} 需要前缀 "${certPrefix}"` };
  }
  
  if (!prefixList.includes(certPrefix)) {
    return { matches: false, warning: `证书 ${certDomain} 仅适用于前缀 "${certPrefix}"，当前前缀不匹配` };
  }
  
  return { matches: true };
};

// 生成证书文件路径
export const generateCertPaths = (domain: string): { crtPath: string; keyPath: string } => {
  return {
    crtPath: `/opt/frpc/certs/${domain}.crt`,
    keyPath: `/opt/frpc/certs/${domain}.key`,
  };
};