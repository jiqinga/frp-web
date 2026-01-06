import { Clock, Save } from 'lucide-react';
import { Button, NumberStepper, Card, CardHeader, CardContent } from '../../../components/ui';
import type { Setting } from '../../../api/setting';

interface IntervalSettingsProps {
  settings: Setting[];
  intervalValues: Record<string, string>;
  setIntervalValues: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  loading: boolean;
  onSubmit: () => void;
}

export function IntervalSettings({
  settings,
  intervalValues,
  setIntervalValues,
  loading,
  onSubmit,
}: IntervalSettingsProps) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Clock className="h-5 w-5 text-cyan-400" />
          <span>采集间隔设置</span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="max-w-xl space-y-4">
          {settings.map(s => (
            <div key={s.key} className="space-y-2">
              <label className="text-sm font-medium text-foreground-secondary">
                {s.description || s.key}
              </label>
              <div className="flex items-center gap-2">
                <NumberStepper
                  value={intervalValues[s.key] || ''}
                  onChange={(value) => setIntervalValues(prev => ({ ...prev, [s.key]: value }))}
                  min={5}
                  max={3600}
                  step={5}
                />
                <span className="text-sm text-foreground-muted">秒</span>
              </div>
            </div>
          ))}
          <Button onClick={onSubmit} loading={loading} icon={<Save />}>
            保存采集间隔
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}