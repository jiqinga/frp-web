import { Globe, Plus, Edit, Trash2, Check } from 'lucide-react';
import { Button, Input, Modal, Table, Card, CardHeader, CardContent, Badge, Switch, Textarea, Tooltip } from '../../../components/ui';
import type { GithubMirror } from '../../../api/githubMirror';

interface MirrorSettingsProps {
  mirrors: GithubMirror[];
  mirrorModalVisible: boolean;
  setMirrorModalVisible: (visible: boolean) => void;
  editingMirror: GithubMirror | null;
  mirrorForm: {
    name: string;
    base_url: string;
    description: string;
    enabled: boolean;
  };
  setMirrorForm: React.Dispatch<React.SetStateAction<{
    name: string;
    base_url: string;
    description: string;
    enabled: boolean;
  }>>;
  onAdd: () => void;
  onEdit: (mirror: GithubMirror) => void;
  onDelete: (id: number) => void;
  onSetDefault: (id: number) => void;
  onSubmit: () => void;
}

export function MirrorSettings({
  mirrors,
  mirrorModalVisible,
  setMirrorModalVisible,
  editingMirror,
  mirrorForm,
  setMirrorForm,
  onAdd,
  onEdit,
  onDelete,
  onSetDefault,
  onSubmit,
}: MirrorSettingsProps) {
  const columns = [
    {
      key: 'name',
      title: '名称',
      render: (_: unknown, record: GithubMirror) => (
        <div className="flex items-center justify-center gap-2">
          <Globe className="h-4 w-4 text-indigo-400" />
          <span className="font-medium text-foreground">{record.name}</span>
        </div>
      )
    },
    {
      key: 'base_url',
      title: '地址',
      render: (_: unknown, record: GithubMirror) => (
        <span className="font-mono text-sm text-foreground-secondary">{record.base_url}</span>
      )
    },
    {
      key: 'status',
      title: '状态',
      render: (_: unknown, record: GithubMirror) => (
        <div className="flex items-center justify-center gap-2">
          {record.is_default && <Badge variant="primary">默认</Badge>}
          <Badge variant={record.enabled ? 'success' : 'default'}>
            {record.enabled ? '启用' : '禁用'}
          </Badge>
        </div>
      ),
    },
    { 
      key: 'description', 
      title: '描述',
      render: (_: unknown, record: GithubMirror) => (
        <span className="text-foreground-muted">{record.description || '-'}</span>
      )
    },
    {
      key: 'action',
      title: '操作',
      render: (_: unknown, record: GithubMirror) => (
        <div className="flex items-center justify-center gap-1">
          {!record.is_default && (
            <Tooltip content="设为默认">
              <Button size="sm" variant="ghost" onClick={() => onSetDefault(record.id)}>
                <Check className="h-4 w-4" />
              </Button>
            </Tooltip>
          )}
          <Tooltip content="编辑">
            <Button size="sm" variant="ghost" onClick={() => onEdit(record)}>
              <Edit className="h-4 w-4" />
            </Button>
          </Tooltip>
          <Tooltip content="删除">
            <Button size="sm" variant="ghost" onClick={() => onDelete(record.id)}>
              <Trash2 className="h-4 w-4 text-red-400" />
            </Button>
          </Tooltip>
        </div>
      ),
    },
  ];

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Globe className="h-5 w-5 text-green-400" />
              <span>GitHub加速源管理</span>
            </div>
            <Button size="sm" onClick={onAdd} icon={<Plus />}>
              添加加速源
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table
            columns={columns}
            data={mirrors}
            rowKey="id"
            emptyText="暂无加速源"
          />
        </CardContent>
      </Card>

      <Modal
        open={mirrorModalVisible}
        onClose={() => setMirrorModalVisible(false)}
        title={editingMirror ? '编辑加速源' : '添加加速源'}
      >
        <div className="space-y-4">
          <Input
            label="名称"
            value={mirrorForm.name}
            onChange={(e) => setMirrorForm(prev => ({ ...prev, name: e.target.value }))}
            required
          />
          <Input
            label="地址"
            value={mirrorForm.base_url}
            onChange={(e) => setMirrorForm(prev => ({ ...prev, base_url: e.target.value }))}
            placeholder="https://xget.183321.xyz/gh"
            required
          />
          <Textarea
            label="描述"
            value={mirrorForm.description}
            onChange={(e) => setMirrorForm(prev => ({ ...prev, description: e.target.value }))}
            rows={3}
          />
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-foreground-secondary">启用</span>
            <Switch
              checked={mirrorForm.enabled}
              onChange={(checked) => setMirrorForm(prev => ({ ...prev, enabled: checked }))}
            />
          </div>
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="secondary" onClick={() => setMirrorModalVisible(false)}>
              取消
            </Button>
            <Button onClick={onSubmit}>
              {editingMirror ? '更新' : '创建'}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}