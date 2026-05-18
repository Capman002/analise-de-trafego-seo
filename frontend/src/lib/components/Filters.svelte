<script lang="ts">
  import DateRangePicker from './DateRangePicker.svelte';

  interface Props {
    period: number;
    startDate?: string;
    endDate?: string;
    location: string;
    onperiodchange: (period: number, start: string, end: string) => void;
    onlocationchange: (location: string) => void;
  }

  let { period, startDate = '', endDate = '', location, onperiodchange, onlocationchange }: Props = $props();

  const LOCATIONS = [
    { code: 'ALL', name: 'Todos os países' },
    { code: 'BRA', name: 'Brasil' },
    { code: 'USA', name: 'Estados Unidos' },
    { code: 'PRT', name: 'Portugal' },
    { code: 'ESP', name: 'Espanha' },
    { code: 'GBR', name: 'Reino Unido' },
    { code: 'FRA', name: 'França' },
    { code: 'DEU', name: 'Alemanha' },
    { code: 'ARG', name: 'Argentina' },
    { code: 'MEX', name: 'México' },
    { code: 'ITA', name: 'Itália' },
  ];
</script>

<div class="field">
  <label for="location-select" class="field-label">Região</label>
  <select
    id="location-select"
    class="select"
    value={location}
    onchange={(e) => onlocationchange((e.target as HTMLSelectElement).value)}
  >
    {#each LOCATIONS as loc (loc.code)}
      <option value={loc.code}>{loc.name}</option>
    {/each}
  </select>
</div>

<DateRangePicker {period} {startDate} {endDate} onchange={onperiodchange} />

<style>
  .field { display: flex; flex-direction: column; gap: 6px; }
  .field-label { font-size: 11px; font-weight: 600; color: var(--ink-subtle); text-transform: uppercase; letter-spacing: 0.5px; }
  .select { height: 40px; padding: 0 32px 0 12px; border: 1px solid var(--line-soft); border-radius: 5px; background: var(--surface); color: var(--ink); font-size: 14px; font-weight: 500; cursor: pointer; outline: none; appearance: none; transition: border-color 0.2s; background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e"); background-repeat: no-repeat; background-position: right 10px center; background-size: 16px; }
  .select:hover, .select:focus { border-color: var(--line); }
</style>
