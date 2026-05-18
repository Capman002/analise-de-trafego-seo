<script lang="ts">
  import '../app.css';
  let { children } = $props();
</script>

<div class="dashboard-main stack">
  <!-- Header -->
  <div class="pg-header">
    <div style="display: flex; align-items: center; gap: var(--s-4);">
      <div class="hdr-icon">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
      </div>
      <div>
        <h1 class="pg-title">Análise de Tráfego</h1>
        <p class="pg-subtitle">Dashboard de inteligência de tráfego orgânico</p>
      </div>
    </div>
  </div>

  {@render children()}
</div>
