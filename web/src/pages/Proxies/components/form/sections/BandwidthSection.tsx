import { memo } from 'react';
import { Switch } from '../../../../../components/ui/Switch';
import { Select, type SelectOption } from '../../../../../components/ui/Select';
import { NumberStepper } from '../../../../../components/ui/NumberStepper';
import { Tooltip } from '../../../../../components/ui/Tooltip';
import { HelpCircle } from 'lucide-react';
import { FIELD_LABELS, FIELD_TOOLTIPS, BANDWIDTH_UNITS } from '../../../constants';

interface BandwidthSectionProps {
  bandwidthEnabled: boolean;
  bandwidthValue: number | undefined;
  bandwidthUnit: string;
  onBandwidthEnabledChange: (enabled: boolean) => void;
  onBandwidthValueChange: (value: number | undefined) => void;
  onBandwidthUnitChange: (unit: string) => void;
}

const bandwidthUnitOptions: SelectOption[] = BANDWIDTH_UNITS.map(unit => ({
  value: unit.value,
  label: unit.label,
}));

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

export const BandwidthSection = memo(function BandwidthSection({
  bandwidthEnabled,
  bandwidthValue,
  bandwidthUnit,
  onBandwidthEnabledChange,
  onBandwidthValueChange,
  onBandwidthUnitChange,
}: BandwidthSectionProps) {
  return (
    <div className="space-y-1.5">
      {renderLabel(FIELD_LABELS.bandwidth_limit, FIELD_TOOLTIPS.bandwidth_limit)}
      <div className="flex items-center gap-3">
        <Switch
          checked={bandwidthEnabled}
          onChange={(checked) => {
            onBandwidthEnabledChange(checked);
            if (!checked) {
              onBandwidthValueChange(undefined);
            }
          }}
          size="sm"
        />
        {bandwidthEnabled && (
          <>
            <NumberStepper
              value={bandwidthValue?.toString() || ''}
              onChange={(value) => onBandwidthValueChange(value ? parseInt(value) : undefined)}
              min={1}
              max={10000}
              step={1}
            />
            <Select
              value={bandwidthUnit}
              onChange={(value) => onBandwidthUnitChange(value as string)}
              options={bandwidthUnitOptions}
              className="w-24"
            />
          </>
        )}
      </div>
    </div>
  );
});