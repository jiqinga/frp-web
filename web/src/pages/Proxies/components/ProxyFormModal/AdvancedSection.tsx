import { HelpCircle } from 'lucide-react';
import { Input } from '../../../../components/ui/Input';
import { Select, type SelectOption } from '../../../../components/ui/Select';
import { Switch } from '../../../../components/ui/Switch';
import { Tooltip } from '../../../../components/ui/Tooltip';
import { NumberStepper } from '../../../../components/ui/NumberStepper';
import { FIELD_LABELS, FIELD_TOOLTIPS, FIELD_PLACEHOLDERS, BANDWIDTH_UNITS } from '../../constants';
import type { FormData } from './types';

interface AdvancedSectionProps {
  formData: FormData;
  bandwidthEnabled: boolean;
  bandwidthValue: number | undefined;
  bandwidthUnit: string;
  errors: Record<string, string>;
  shouldShowField: (field: string) => boolean;
  updateField: (field: keyof FormData, value: string | number | boolean | undefined) => void;
  setBandwidthEnabled: (enabled: boolean) => void;
  setBandwidthValue: (value: number | undefined) => void;
  setBandwidthUnit: (unit: string) => void;
}

const renderLabel = (label: string, tooltip?: string) => (
  <div className="flex items-center gap-1">
    <span>{label}</span>
    {tooltip && (
      <Tooltip content={tooltip}>
        <span className="cursor-help text-foreground-subtle">
          <HelpCircle className="h-3.5 w-3.5" />
        </span>
      </Tooltip>
    )}
  </div>
);

export function AdvancedSection({
  formData,
  bandwidthEnabled,
  bandwidthValue,
  bandwidthUnit,
  errors,
  shouldShowField,
  updateField,
  setBandwidthEnabled,
  setBandwidthValue,
  setBandwidthUnit,
}: AdvancedSectionProps) {
  const bandwidthUnitOptions: SelectOption[] = BANDWIDTH_UNITS.map(unit => ({
    value: unit.value,
    label: unit.label,
  }));

  return (
    <>
      {shouldShowField('remote_port') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.remote_port, FIELD_TOOLTIPS.remote_port)}
          <Input
            type="number"
            value={formData.remote_port?.toString() || ''}
            onChange={(e) => updateField('remote_port', e.target.value ? parseInt(e.target.value) : undefined)}
            placeholder={FIELD_PLACEHOLDERS.remote_port}
          />
        </div>
      )}

      {shouldShowField('locations') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.locations, FIELD_TOOLTIPS.locations)}
          <Input
            value={formData.locations}
            onChange={(e) => updateField('locations', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.locations}
          />
        </div>
      )}
      
      {shouldShowField('host_header_rewrite') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.host_header_rewrite, FIELD_TOOLTIPS.host_header_rewrite)}
          <Input
            value={formData.host_header_rewrite}
            onChange={(e) => updateField('host_header_rewrite', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.host_header_rewrite}
          />
        </div>
      )}
      
      {shouldShowField('http_user') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.http_user, FIELD_TOOLTIPS.http_user)}
          <Input
            value={formData.http_user}
            onChange={(e) => updateField('http_user', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.http_user}
          />
        </div>
      )}
      
      {shouldShowField('http_password') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.http_password, FIELD_TOOLTIPS.http_password)}
          <Input
            type="password"
            value={formData.http_password}
            onChange={(e) => updateField('http_password', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.http_password}
          />
        </div>
      )}
      
      {shouldShowField('secret_key') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.secret_key, FIELD_TOOLTIPS.secret_key)}
          <Input
            type="password"
            value={formData.secret_key}
            onChange={(e) => updateField('secret_key', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.secret_key}
            error={errors.secret_key}
          />
        </div>
      )}
      
      {shouldShowField('allow_users') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.allow_users, FIELD_TOOLTIPS.allow_users)}
          <Input
            value={formData.allow_users}
            onChange={(e) => updateField('allow_users', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.allow_users}
          />
        </div>
      )}

      <div className="space-y-1.5">
        {renderLabel(FIELD_LABELS.bandwidth_limit, FIELD_TOOLTIPS.bandwidth_limit)}
        <div className="flex items-center gap-3">
          <Switch
            checked={bandwidthEnabled}
            onChange={(checked) => {
              setBandwidthEnabled(checked);
              if (!checked) setBandwidthValue(undefined);
            }}
            size="sm"
          />
          {bandwidthEnabled && (
            <>
              <NumberStepper
                value={bandwidthValue?.toString() || ''}
                onChange={(value) => setBandwidthValue(value ? parseInt(value) : undefined)}
                min={1}
                max={10000}
                step={1}
              />
              <Select
                value={bandwidthUnit}
                onChange={(value) => setBandwidthUnit(value as string)}
                options={bandwidthUnitOptions}
                className="w-24"
              />
            </>
          )}
        </div>
      </div>
    </>
  );
}