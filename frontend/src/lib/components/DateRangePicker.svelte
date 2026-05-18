<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { onMount } from 'svelte';

  interface Props {
    period: number;
    startDate: string;
    endDate: string;
    onchange: (period: number, start: string, end: string) => void;
  }
  let { period, startDate, endDate, onchange }: Props = $props();

  let open = $state(false);
  let customStart = $state(startDate);
  let customEnd = $state(endDate);

  const PRESETS = [
    { days: 7, label: '7 dias' },
    { days: 28, label: '28 dias' },
    { days: 90, label: '3 meses' },
    { days: 180, label: '6 meses' },
    { days: 365, label: '12 meses' },
    { days: 480, label: '16 meses' },
  ];

  let displayLabel = $derived.by(() => {
    if (startDate && endDate) {
      // Ex: 01/Jan a 15/Jan
      const s = new Date(startDate + 'T00:00:00');
      const e = new Date(endDate + 'T00:00:00');
      return `${s.toLocaleDateString('pt-BR', {day:'2-digit', month:'short'})} a ${e.toLocaleDateString('pt-BR', {day:'2-digit', month:'short'})}`;
    }
    const preset = PRESETS.find(p => p.days === period);
    return preset ? preset.label : `${period} dias`;
  });

  function selectPreset(days: number) {
    onchange(days, '', '');
    open = false;
  }

  function applyCustom() {
    if (customStart && customEnd) {
      onchange(0, customStart, customEnd);
      open = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') open = false;
  }

  function clickOutside(e: MouseEvent) {
    const t = e.target as HTMLElement;
    if (!t.closest('.date-picker-root')) {
      open = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', clickOutside);
    return () => document.removeEventListener('click', clickOutside);
  });
</script>

<div class="date-picker-root" onkeydown={handleKeydown} role="button" tabindex="0">
  <label class="field-label">Período</label>
  <button class="select date-trigger" onclick={() => open = !open}>
    {displayLabel}
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>
  </button>

  {#if open}
    <div class="popover glass-card" transition:fly={{ y: -5, duration: 150 }}>
      <div class="popover-layout">
        <!-- Presets -->
        <div class="presets-list">
          <div class="group-title">Períodos Fixos</div>
          {#each PRESETS as p}
            <button 
              class="preset-btn" 
              class:active={period === p.days && !startDate}
              onclick={() => selectPreset(p.days)}
            >
              {p.label}
            </button>
          {/each}
        </div>

        <!-- Custom Calendar (Native Inputs Styled) -->
        <div class="custom-range">
          <div class="group-title">Personalizado</div>
          <div class="date-inputs">
            <div class="input-group">
              <label>De</label>
              <input type="date" bind:value={customStart} class="styled-date" max={customEnd || undefined} />
            </div>
            <div class="input-group">
              <label>Até</label>
              <input type="date" bind:value={customEnd} class="styled-date" min={customStart || undefined} />
            </div>
          </div>
          <button class="apply-btn" onclick={applyCustom} disabled={!customStart || !customEnd || customStart > customEnd}>
            Aplicar Período
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .date-picker-root { position: relative; display: flex; flex-direction: column; gap: 6px; }
  .field-label { font-size: 11px; font-weight: 600; color: var(--ink-subtle); text-transform: uppercase; letter-spacing: 0.5px; }
  
  .date-trigger {
    height: 40px;
    padding: 0 12px;
    border: 1px solid var(--line-soft);
    border-radius: 5px;
    background: var(--surface);
    color: var(--ink);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    min-width: 140px;
    transition: all 0.2s;
  }
  .date-trigger:hover { border-color: var(--line); }

  .popover {
    position: absolute;
    top: calc(100% + 8px);
    left: 0;
    z-index: 100;
    min-width: 380px;
    padding: 0;
    border-radius: 8px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.08);
    overflow: hidden;
  }

  .popover-layout {
    display: flex;
    background: var(--surface);
  }

  .presets-list {
    width: 130px;
    border-right: 1px solid var(--line-soft);
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    background: var(--surface-2);
  }

  .group-title {
    font-size: 11px;
    font-weight: 600;
    color: var(--ink-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
    padding: 0 8px;
  }

  .preset-btn {
    width: 100%;
    text-align: left;
    padding: 8px 10px;
    background: transparent;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    color: var(--ink);
    cursor: pointer;
    transition: all 0.15s;
  }
  .preset-btn:hover { background: rgba(0,0,0,0.04); }
  .preset-btn.active { background: var(--accent); color: white; font-weight: 500; }

  .custom-range {
    flex: 1;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .date-inputs {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .input-group label {
    font-size: 12px;
    color: var(--ink-subtle);
  }

  .styled-date {
    height: 36px;
    padding: 0 12px;
    border: 1px solid var(--line-soft);
    border-radius: 5px;
    background: var(--surface-2);
    color: var(--ink);
    font-family: inherit;
    font-size: 14px;
    outline: none;
    transition: border-color 0.2s;
  }
  .styled-date:focus { border-color: var(--accent); }
  .styled-date::-webkit-calendar-picker-indicator { cursor: pointer; opacity: 0.6; transition: 0.2s; }
  .styled-date::-webkit-calendar-picker-indicator:hover { opacity: 1; }

  .apply-btn {
    margin-top: auto;
    height: 36px;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 5px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  .apply-btn:hover:not(:disabled) { opacity: 0.9; }
  .apply-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
