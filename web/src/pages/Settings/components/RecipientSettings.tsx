import { useState, useEffect } from 'react';
import { Users, UserPlus, Trash2, Edit, Save, X, UsersRound } from 'lucide-react';
import { Button, Input, Card, CardHeader, CardContent, Modal, Switch, Transfer } from '../../../components/ui';
import { ConfirmModal } from '../../../components/ui/ConfirmModal';
import { toast } from '../../../components/ui/Toast';
import { alertRecipientApi, type AlertRecipient, type AlertRecipientGroup } from '../../../api/alertRecipient';

export function RecipientSettings() {
  const [recipients, setRecipients] = useState<AlertRecipient[]>([]);
  const [groups, setGroups] = useState<AlertRecipientGroup[]>([]);
  const [loading, setLoading] = useState(false);
  
  // 接收人表单
  const [recipientModal, setRecipientModal] = useState(false);
  const [editingRecipient, setEditingRecipient] = useState<AlertRecipient | null>(null);
  const [recipientForm, setRecipientForm] = useState({ name: '', email: '', enabled: true });
  
  // 分组表单
  const [groupModal, setGroupModal] = useState(false);
  const [editingGroup, setEditingGroup] = useState<AlertRecipientGroup | null>(null);
  const [groupForm, setGroupForm] = useState({ name: '', description: '', enabled: true, recipientIds: [] as number[] });

  // 删除确认
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [deleteType, setDeleteType] = useState<'recipient' | 'group'>('recipient');
  const [deletingId, setDeletingId] = useState<number | null>(null);

  useEffect(() => { loadData(); }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [r, g] = await Promise.all([alertRecipientApi.getRecipients(), alertRecipientApi.getGroups()]);
      setRecipients(r || []);
      setGroups(g || []);
    } catch { toast.error('加载失败'); }
    setLoading(false);
  };

  // 接收人操作
  const openRecipientModal = (r?: AlertRecipient) => {
    setEditingRecipient(r || null);
    setRecipientForm(r ? { name: r.name, email: r.email, enabled: r.enabled } : { name: '', email: '', enabled: true });
    setRecipientModal(true);
  };

  const saveRecipient = async () => {
    if (!recipientForm.name || !recipientForm.email) return toast.error('请填写完整');
    try {
      if (editingRecipient?.id) {
        await alertRecipientApi.updateRecipient(editingRecipient.id, { ...recipientForm, id: editingRecipient.id });
      } else {
        await alertRecipientApi.createRecipient(recipientForm);
      }
      toast.success('保存成功');
      setRecipientModal(false);
      loadData();
    } catch { toast.error('保存失败'); }
  };

  const handleDeleteRecipient = (id: number) => {
    setDeleteType('recipient');
    setDeletingId(id);
    setDeleteConfirmVisible(true);
  };

  const handleDeleteGroup = (id: number) => {
    setDeleteType('group');
    setDeletingId(id);
    setDeleteConfirmVisible(true);
  };

  const confirmDelete = async () => {
    if (deletingId === null) return;
    try {
      if (deleteType === 'recipient') {
        await alertRecipientApi.deleteRecipient(deletingId);
      } else {
        await alertRecipientApi.deleteGroup(deletingId);
      }
      toast.success('删除成功');
      loadData();
    } catch { toast.error('删除失败'); }
    setDeleteConfirmVisible(false);
    setDeletingId(null);
  };

  // 分组操作
  const openGroupModal = (g?: AlertRecipientGroup) => {
    setEditingGroup(g || null);
    setGroupForm(g ? { 
      name: g.name, 
      description: g.description, 
      enabled: g.enabled,
      recipientIds: g.recipients?.map(r => r.id!) || []
    } : { name: '', description: '', enabled: true, recipientIds: [] });
    setGroupModal(true);
  };

  const saveGroup = async () => {
    if (!groupForm.name) return toast.error('请填写分组名称');
    try {
      if (editingGroup?.id) {
        await alertRecipientApi.updateGroup(editingGroup.id, { ...groupForm, id: editingGroup.id });
        await alertRecipientApi.setGroupRecipients(editingGroup.id, groupForm.recipientIds);
      } else {
        const res = await alertRecipientApi.createGroup(groupForm) as AlertRecipientGroup;
        if (res?.id) {
          await alertRecipientApi.setGroupRecipients(res.id, groupForm.recipientIds);
        }
      }
      toast.success('保存成功');
      setGroupModal(false);
      loadData();
    } catch { toast.error('保存失败'); }
  };


  return (
    <div className="space-y-6">
      {/* 接收人列表 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Users className="h-5 w-5 text-indigo-400" />
              <span>告警接收人</span>
            </div>
            <Button size="sm" icon={<UserPlus />} onClick={() => openRecipientModal()}>添加</Button>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? <div className="text-foreground-muted">加载中...</div> : (
            <div className="space-y-2">
              {recipients.length === 0 ? <div className="text-foreground-subtle">暂无接收人</div> : recipients.map(r => (
                <div key={r.id} className="flex items-center justify-between p-3 rounded-lg bg-surface-hover">
                  <div>
                    <div className="font-medium text-foreground">{r.name}</div>
                    <div className="text-sm text-foreground-muted">{r.email}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-xs px-2 py-1 rounded ${r.enabled ? 'bg-green-500/20 text-green-400' : 'bg-surface-active text-foreground-muted'}`}>
                      {r.enabled ? '启用' : '禁用'}
                    </span>
                    <Button size="sm" variant="ghost" icon={<Edit className="h-4 w-4" />} onClick={() => openRecipientModal(r)} />
                    <Button size="sm" variant="ghost" icon={<Trash2 className="h-4 w-4 text-red-400" />} onClick={() => handleDeleteRecipient(r.id!)} />
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 分组列表 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <UsersRound className="h-5 w-5 text-indigo-400" />
              <span>接收人分组</span>
            </div>
            <Button size="sm" icon={<UserPlus />} onClick={() => openGroupModal()}>添加</Button>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? <div className="text-foreground-muted">加载中...</div> : (
            <div className="space-y-2">
              {groups.length === 0 ? <div className="text-foreground-subtle">暂无分组</div> : groups.map(g => (
                <div key={g.id} className="p-3 rounded-lg bg-surface-hover">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium text-foreground">{g.name}</div>
                      <div className="text-sm text-foreground-muted">{g.description || '无描述'}</div>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className={`text-xs px-2 py-1 rounded ${g.enabled ? 'bg-green-500/20 text-green-400' : 'bg-surface-active text-foreground-muted'}`}>
                        {g.enabled ? '启用' : '禁用'}
                      </span>
                      <Button size="sm" variant="ghost" icon={<Edit className="h-4 w-4" />} onClick={() => openGroupModal(g)} />
                      <Button size="sm" variant="ghost" icon={<Trash2 className="h-4 w-4 text-red-400" />} onClick={() => handleDeleteGroup(g.id!)} />
                    </div>
                  </div>
                  {g.recipients && g.recipients.length > 0 && (
                    <div className="mt-2 flex flex-wrap gap-1">
                      {g.recipients.map(r => (
                        <span key={r.id} className="text-xs px-2 py-1 rounded bg-surface-active text-foreground-secondary">{r.name}</span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 接收人编辑弹窗 */}
      <Modal open={recipientModal} onClose={() => setRecipientModal(false)} title={editingRecipient ? '编辑接收人' : '添加接收人'}>
        <div className="space-y-4">
          <Input label="名称" value={recipientForm.name} onChange={e => setRecipientForm(f => ({ ...f, name: e.target.value }))} />
          <Input label="邮箱" type="email" value={recipientForm.email} onChange={e => setRecipientForm(f => ({ ...f, email: e.target.value }))} />
          <div className="flex items-center justify-between">
            <span className="text-sm text-foreground-secondary">启用</span>
            <Switch checked={recipientForm.enabled} onChange={v => setRecipientForm(f => ({ ...f, enabled: v }))} />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" icon={<X />} onClick={() => setRecipientModal(false)}>取消</Button>
            <Button icon={<Save />} onClick={saveRecipient}>保存</Button>
          </div>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        open={deleteConfirmVisible}
        onClose={() => setDeleteConfirmVisible(false)}
        onConfirm={confirmDelete}
        title={deleteType === 'recipient' ? '删除接收人' : '删除分组'}
        content="确定删除？删除后无法恢复。"
        type="warning"
        confirmText="删除"
        cancelText="取消"
      />

      {/* 分组编辑弹窗 */}
      <Modal open={groupModal} onClose={() => setGroupModal(false)} title={editingGroup ? '编辑分组' : '添加分组'}>
        <div className="space-y-4">
          <Input label="分组名称" value={groupForm.name} onChange={e => setGroupForm(f => ({ ...f, name: e.target.value }))} />
          <Input label="描述" value={groupForm.description} onChange={e => setGroupForm(f => ({ ...f, description: e.target.value }))} />
          <div className="flex items-center justify-between">
            <span className="text-sm text-foreground-secondary">启用</span>
            <Switch checked={groupForm.enabled} onChange={v => setGroupForm(f => ({ ...f, enabled: v }))} />
          </div>
          <div>
            <label className="text-sm mb-2 block text-foreground-secondary">选择成员</label>
            <Transfer
              dataSource={recipients.map(r => ({ key: r.id!, title: r.name, description: r.email }))}
              targetKeys={groupForm.recipientIds}
              onChange={keys => setGroupForm(f => ({ ...f, recipientIds: keys as number[] }))}
              titles={['可选成员', '已选成员']}
              searchPlaceholder="搜索成员..."
              height={180}
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" icon={<X />} onClick={() => setGroupModal(false)}>取消</Button>
            <Button icon={<Save />} onClick={saveGroup}>保存</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}