<script lang="ts">
  import type { Snippet } from 'svelte';
  interface Props {
    title: string; subtitle: string; count: number; color: string;
    trend: 'up' | 'down'; onclick: () => void; icon?: Snippet;
  }
  let { title, subtitle, count, color, trend, onclick, icon }: Props = $props();
</script>

<button class="section-card glass-card" onclick={onclick} type="button">
  <div class="section-card__top">
    <div class="section-card__icon" style="background: {trend === 'up' ? 'rgba(34,197,94,0.12)' : 'rgba(239,68,68,0.12)'}; color: {color}">
      {#if icon}{@render icon()}{/if}
    </div>
    <div class="section-card__badge">
      {#if trend === 'up'}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 7h10v10"/><path d="M7 17 17 7"/></svg>
      {:else}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 7h10v10"/><path d="M17 7 7 17"/></svg>
      {/if}
      {count} item{count !== 1 ? 's' : ''}
    </div>
  </div>
  <div>
    <h3 class="section-card__title">{title}</h3>
    <p class="section-card__subtitle">{subtitle}</p>
  </div>
  <div class="section-card__action">
    Ver detalhes
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 7h10v10"/><path d="M7 17 17 7"/></svg>
  </div>
</button>

<style>
  .section-card {
    text-align: left; padding: 18px; cursor: pointer;
    border: 1px solid var(--line-soft); background: var(--surface);
    display: flex; flex-direction: column; gap: 12px; min-height: 130px;
    transition: transform 0.15s ease;
  }
  .section-card:hover { transform: translateY(-2px); }
  .section-card:active { transform: scale(0.99); }
  .section-card__top { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
  .section-card__icon { width: 36px; height: 36px; border-radius: 5px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .section-card__badge { font-size: 11px; font-weight: 500; color: var(--ink-subtle); text-transform: uppercase; letter-spacing: 0.5px; display: flex; align-items: center; gap: 4px; }
  .section-card__title { margin: 0; font-size: 14px; font-weight: 600; color: var(--ink); margin-bottom: 4px; }
  .section-card__subtitle { margin: 0; font-size: 12px; color: var(--ink-subtle); line-height: 1.4; }
  .section-card__action { display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 500; color: var(--accent); margin-top: auto; }
</style>
