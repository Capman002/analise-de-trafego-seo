<script lang="ts">
  import type { Snippet } from 'svelte';
  interface Props {
    label: string;
    value: string | number;
    subValue?: string;
    color?: string;
    icon?: Snippet;
    delta?: number | null;       // Variação % vs período anterior
    deltaInverted?: boolean;     // true = queda é positiva (ex: posição média)
  }
  let { label, value, subValue, color = '#008272', icon, delta = null, deltaInverted = false }: Props = $props();

  let deltaClass = $derived(
    delta === null || delta === undefined ? '' :
    deltaInverted
      ? (delta < 0 ? 'delta--positive' : delta > 0 ? 'delta--negative' : 'delta--neutral')
      : (delta > 0 ? 'delta--positive' : delta < 0 ? 'delta--negative' : 'delta--neutral')
  );

  let deltaLabel = $derived(
    delta === null || delta === undefined ? '' :
    (delta > 0 ? '+' : '') + delta.toFixed(1) + '%'
  );
</script>

<div class="metric-card glass-card">
  <div class="metric-card__header">
    <div class="metric-card__icon" style="color: {color}">
      {#if icon}{@render icon()}{/if}
    </div>
    <span class="metric-card__label">{label}</span>
  </div>
  <div class="metric-card__value">{value}</div>
  <div class="metric-card__footer">
    {#if subValue}
      <span class="metric-card__sub">{subValue}</span>
    {/if}
    {#if delta !== null && delta !== undefined}
      <span class="metric-card__delta {deltaClass}">
        {#if delta > 0}
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M7 17V7h10"/><path d="M7 7l10 10"/></svg>
        {:else if delta < 0}
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M17 7v10H7"/><path d="M17 17 7 7"/></svg>
        {:else}
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/></svg>
        {/if}
        {deltaLabel}
      </span>
    {/if}
  </div>
</div>

<style>
  .metric-card { padding: 16px; min-width: 140px; flex: 1 1 180px; transition: transform 0.15s ease; }
  .metric-card:hover { transform: scale(1.02); }
  .metric-card__header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
  .metric-card__icon { display: flex; }
  .metric-card__label { font-size: 11px; text-transform: uppercase; letter-spacing: 0.5px; color: var(--ink-subtle); }
  .metric-card__value { font-size: 24px; font-weight: 700; color: var(--ink); line-height: 1.2; font-variant-numeric: tabular-nums; }
  .metric-card__footer { display: flex; align-items: center; gap: 8px; margin-top: 6px; flex-wrap: nowrap; overflow: hidden; }
  .metric-card__sub { font-size: 12px; color: var(--ink-subtle); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex-shrink: 1; min-width: 0; }
  .metric-card__delta {
    display: inline-flex; align-items: center; gap: 3px;
    padding: 2px 7px; border-radius: var(--r-full);
    font-size: 11px; font-weight: 600; font-variant-numeric: tabular-nums;
    transition: opacity 0.2s ease;
    flex-shrink: 0;
  }
  .delta--positive { color: var(--ok); background: var(--ok-wash); }
  .delta--negative { color: var(--danger); background: var(--danger-wash); }
  .delta--neutral { color: var(--ink-muted); background: var(--surface-2); }
</style>
