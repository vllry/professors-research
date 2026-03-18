export type FeatureStage = 'alpha' | 'beta';

const STAGE_STYLES: Record<
  FeatureStage,
  { label: string; className: string; title: string }
> = {
  alpha: {
    label: 'Alpha',
    title: 'Alpha feature (early, may change)',
    className: 'bg-amber-100 text-amber-900 ring-1 ring-amber-300',
  },
  beta: {
    label: 'Beta',
    title: 'Beta feature (stable-ish, still iterating)',
    className: 'bg-indigo-100 text-indigo-900 ring-1 ring-indigo-300',
  },
};

export default function FeatureBadge({ stage }: { stage: FeatureStage }) {
  const style = STAGE_STYLES[stage];

  return (
    <span
      className={[
        'inline-flex items-center rounded-full px-2.5 py-1 text-xs font-extrabold uppercase tracking-wider shadow-sm',
        style.className,
      ].join(' ')}
      title={style.title}
      aria-label={style.title}
    >
      {style.label}
    </span>
  );
}

