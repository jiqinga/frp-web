import { ArrowUp, Cloud, RefreshCw, Info } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Button } from '../../../components/ui/Button';
import { Input } from '../../../components/ui/Input';
import { Select, type SelectOption } from '../../../components/ui/Select';
import { Tooltip } from '../../../components/ui/Tooltip';
import { Checkbox } from '../../../components/ui/Checkbox';
import type { GithubMirror } from '../../../api/githubMirror';

interface RemoteOperationModalsProps {
  // 升级弹窗
  upgradeVisible: boolean;
  upgradeVersion: string;
  upgradeMirrorId?: number;
  onUpgradeVersionChange: (version: string) => void;
  onUpgradeMirrorChange: (mirrorId?: number) => void;
  onUpgradeConfirm: () => void;
  onUpgradeCancel: () => void;
  
  // 安装弹窗
  installVisible: boolean;
  installMirrorId?: number;
  onInstallMirrorChange: (mirrorId?: number) => void;
  onInstallConfirm: () => void;
  onInstallCancel: () => void;
  
  // 重装弹窗
  reinstallVisible: boolean;
  reinstallMirrorId?: number;
  reinstallRegenerateAuth: boolean;
  onReinstallMirrorChange: (mirrorId?: number) => void;
  onReinstallRegenerateAuthChange: (checked: boolean) => void;
  onReinstallConfirm: () => void;
  onReinstallCancel: () => void;
  
  mirrors: GithubMirror[];
}

/**
 * 远程操作弹窗组件集合
 * 包含：升级、安装、重装三个弹窗
 * 使用 Tailwind CSS + Lucide Icons 重构
 */
export function RemoteOperationModals({
  upgradeVisible,
  upgradeVersion,
  upgradeMirrorId,
  onUpgradeVersionChange,
  onUpgradeMirrorChange,
  onUpgradeConfirm,
  onUpgradeCancel,
  installVisible,
  installMirrorId,
  onInstallMirrorChange,
  onInstallConfirm,
  onInstallCancel,
  reinstallVisible,
  reinstallMirrorId,
  reinstallRegenerateAuth,
  onReinstallMirrorChange,
  onReinstallRegenerateAuthChange,
  onReinstallConfirm,
  onReinstallCancel,
  mirrors,
}: RemoteOperationModalsProps) {
  // 镜像选项
  const mirrorOptions: SelectOption[] = mirrors.map(m => ({
    value: m.id!,
    label: `${m.name}${m.is_default ? ' (默认)' : ''}`,
  }));

  return (
    <>
      {/* 升级版本弹窗 */}
      <Modal
        open={upgradeVisible}
        onClose={onUpgradeCancel}
        title={
          <div className="flex items-center gap-2">
            <ArrowUp className="h-5 w-5 text-green-400" />
            <span>升级FRP版本</span>
          </div>
        }
        size="sm"
        footer={
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={onUpgradeCancel}>
              取消
            </Button>
            <Button onClick={onUpgradeConfirm}>
              确认升级
            </Button>
          </div>
        }
      >
        <div className="space-y-4 py-2">
          <div>
            <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
              版本号
              <Tooltip content="留空使用默认版本 0.65.0">
                <Info className="h-3.5 w-3.5 text-foreground-muted" />
              </Tooltip>
            </label>
            <Input
              placeholder="例如: 0.65.0"
              value={upgradeVersion}
              onChange={(e) => onUpgradeVersionChange(e.target.value)}
            />
            <p className="text-xs text-foreground-muted mt-1">留空使用默认版本 0.65.0</p>
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
              GitHub加速源
              <Tooltip content="选择用于下载FRP的加速源">
                <Info className="h-3.5 w-3.5 text-foreground-muted" />
              </Tooltip>
            </label>
            <Select
              value={upgradeMirrorId}
              onChange={(value) => onUpgradeMirrorChange(value as number | undefined)}
              options={mirrorOptions}
              placeholder="使用默认加速源"
            />
          </div>
        </div>
      </Modal>

      {/* 远程安装弹窗 */}
      <Modal
        open={installVisible}
        onClose={onInstallCancel}
        title={
          <div className="flex items-center gap-2">
            <Cloud className="h-5 w-5 text-blue-400" />
            <span>远程安装FRP</span>
          </div>
        }
        size="sm"
        footer={
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={onInstallCancel}>
              取消
            </Button>
            <Button onClick={onInstallConfirm}>
              确认安装
            </Button>
          </div>
        }
      >
        <div className="space-y-4 py-2">
          <div className="p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <p className="text-sm text-blue-400 dark:text-blue-300">
              将在远程服务器上安装 FRP 服务端程序，安装完成后可通过本平台管理。
            </p>
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
              GitHub加速源
              <Tooltip content="选择用于下载FRP的加速源">
                <Info className="h-3.5 w-3.5 text-foreground-muted" />
              </Tooltip>
            </label>
            <Select
              value={installMirrorId}
              onChange={(value) => onInstallMirrorChange(value as number | undefined)}
              options={mirrorOptions}
              placeholder="使用默认加速源"
            />
          </div>
        </div>
      </Modal>

      {/* 重装弹窗 */}
      <Modal
        open={reinstallVisible}
        onClose={onReinstallCancel}
        title={
          <div className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5 text-orange-400" />
            <span>重装FRP</span>
          </div>
        }
        size="sm"
        footer={
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={onReinstallCancel}>
              取消
            </Button>
            <Button variant="danger" onClick={onReinstallConfirm}>
              确认重装
            </Button>
          </div>
        }
      >
        <div className="space-y-4 py-2">
          <div className="p-3 bg-orange-500/10 border border-orange-500/20 rounded-lg">
            <p className="text-sm text-orange-400 dark:text-orange-300">
              重装将重新下载并安装 FRP 程序，可选择是否重新生成认证信息。
            </p>
          </div>
          
          {/* 重新生成认证选项 */}
          <label className="flex items-start gap-3 cursor-pointer group">
            <Checkbox
              checked={reinstallRegenerateAuth}
              onChange={onReinstallRegenerateAuthChange}
            />
            <div>
              <span className="text-sm text-foreground-secondary group-hover:text-foreground">
                重新生成Dashboard密码和Token
              </span>
              <p className="text-xs mt-0.5 text-foreground-muted">
                推荐开启，但会导致现有客户端连接失效
              </p>
            </div>
          </label>
          
          <div>
            <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
              GitHub加速源
              <Tooltip content="选择用于下载FRP的加速源">
                <Info className="h-3.5 w-3.5 text-foreground-muted" />
              </Tooltip>
            </label>
            <Select
              value={reinstallMirrorId}
              onChange={(value) => onReinstallMirrorChange(value as number | undefined)}
              options={mirrorOptions}
              placeholder="使用默认加速源"
            />
          </div>
        </div>
      </Modal>
    </>
  );
}