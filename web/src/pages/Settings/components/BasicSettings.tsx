import { Settings as SettingsIcon, Save } from 'lucide-react';
import { Button, Input, Card, CardHeader, CardContent } from '../../../components/ui';
import type { Setting } from '../../../api/setting';

interface BasicSettingsProps {
  settings: Setting[];
  formValues: Record<string, string>;
  setFormValues: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  loading: boolean;
  onSubmit: () => void;
}

export function BasicSettings({
  settings,
  formValues,
  setFormValues,
  loading,
  onSubmit,
}: BasicSettingsProps) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <SettingsIcon className="h-5 w-5 text-indigo-400" />
          <span>系统设置</span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="max-w-xl space-y-4">
          {settings.map(s => (
            <Input
              key={s.key}
              label={s.description || s.key}
              value={formValues[s.key] || ''}
              onChange={(e) => setFormValues(prev => ({ ...prev, [s.key]: e.target.value }))}
              placeholder={s.value}
            />
          ))}
          <Button onClick={onSubmit} loading={loading} icon={<Save />}>
            保存设置
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}