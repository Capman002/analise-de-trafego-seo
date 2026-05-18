<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Chart, LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, Tooltip, Legend } from 'chart.js';

  Chart.register(LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, Tooltip, Legend);

  let { data, prevData = [], comparing = false } = $props<{ data: any[], prevData?: any[], comparing?: boolean }>();

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;

  // Toggles for metrics
  let showClicks = $state(true);
  let showImpressions = $state(true);
  let showCtr = $state(false);
  let showPosition = $state(false);

  function toggleMetric(index: number, stateVar: boolean) {
    if (chart) {
      chart.setDatasetVisibility(index, stateVar);
      if (comparing) chart.setDatasetVisibility(index + 4, stateVar);
      chart.update();
    }
  }

  $effect(() => {
    if (chart) {
      chart.setDatasetVisibility(0, showClicks);
      chart.setDatasetVisibility(1, showImpressions);
      chart.setDatasetVisibility(2, showCtr);
      chart.setDatasetVisibility(3, showPosition);
      if (comparing) {
        chart.setDatasetVisibility(4, showClicks);
        chart.setDatasetVisibility(5, showImpressions);
        chart.setDatasetVisibility(6, showCtr);
        chart.setDatasetVisibility(7, showPosition);
      }
      chart.update();
    }
  });

  function initChart() {
    if (!canvas || !data) return;

    if (chart) {
      chart.destroy();
    }

    const labels = data.map((d: any) => d.date);
    const clicks = data.map((d: any) => d.clicks);
    const impressions = data.map((d: any) => d.impressions);
    const ctr = data.map((d: any) => (d.ctr || 0) * 100);
    const position = data.map((d: any) => d.position || 0);

    let datasets: any[] = [
      {
        label: 'Cliques',
        data: clicks,
        borderColor: '#4285F4',
        backgroundColor: 'rgba(66, 133, 244, 0.1)',
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 4,
        tension: 0.3,
        yAxisID: 'y',
        hidden: !showClicks
      },
      {
        label: 'Impressões',
        data: impressions,
        borderColor: '#9E9E9E',
        backgroundColor: 'rgba(158, 158, 158, 0.1)',
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 4,
        tension: 0.3,
        yAxisID: 'y1',
        hidden: !showImpressions
      },
      {
        label: 'CTR',
        data: ctr,
        borderColor: '#0f9d58',
        backgroundColor: 'rgba(15, 157, 88, 0.1)',
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 4,
        tension: 0.3,
        yAxisID: 'yCtr',
        hidden: !showCtr
      },
      {
        label: 'Posição',
        data: position,
        borderColor: '#ab47bc',
        backgroundColor: 'rgba(171, 71, 188, 0.1)',
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 4,
        tension: 0.3,
        yAxisID: 'yPos',
        hidden: !showPosition
      }
    ];

    if (comparing && prevData.length > 0) {
      const prevClicks = prevData.map((d: any) => d.clicks);
      const prevImpressions = prevData.map((d: any) => d.impressions);
      const prevCtr = prevData.map((d: any) => (d.ctr || 0) * 100);
      const prevPosition = prevData.map((d: any) => d.position || 0);

      datasets.push(
        {
          label: 'Cliques (Ant.)',
          data: prevClicks,
          borderColor: 'rgba(66, 133, 244, 0.4)',
          borderWidth: 2,
          borderDash: [5, 5],
          pointRadius: 0,
          tension: 0.3,
          yAxisID: 'y',
          hidden: !showClicks
        },
        {
          label: 'Impr. (Ant.)',
          data: prevImpressions,
          borderColor: 'rgba(158, 158, 158, 0.4)',
          borderWidth: 2,
          borderDash: [5, 5],
          pointRadius: 0,
          tension: 0.3,
          yAxisID: 'y1',
          hidden: !showImpressions
        },
        {
          label: 'CTR (Ant.)',
          data: prevCtr,
          borderColor: 'rgba(15, 157, 88, 0.4)',
          borderWidth: 2,
          borderDash: [5, 5],
          pointRadius: 0,
          tension: 0.3,
          yAxisID: 'yCtr',
          hidden: !showCtr
        },
        {
          label: 'Posição (Ant.)',
          data: prevPosition,
          borderColor: 'rgba(171, 71, 188, 0.4)',
          borderWidth: 2,
          borderDash: [5, 5],
          pointRadius: 0,
          tension: 0.3,
          yAxisID: 'yPos',
          hidden: !showPosition
        }
      );
    }

    chart = new Chart(canvas, {
      type: 'line',
      data: {
        labels,
        datasets
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: 'index',
          intersect: false,
        },
        scales: {
          x: {
            grid: { display: false },
            ticks: { maxTicksLimit: 7 }
          },
          y: {
            type: 'linear',
            display: 'auto',
            position: 'left',
            title: { display: true, text: 'Cliques' },
            beginAtZero: true
          },
          y1: {
            type: 'linear',
            display: 'auto',
            position: 'right',
            title: { display: true, text: 'Impressões' },
            beginAtZero: true,
            grid: { drawOnChartArea: false }
          },
          yCtr: {
            type: 'linear',
            display: 'auto',
            position: 'left',
            title: { display: true, text: 'CTR (%)' },
            beginAtZero: true,
            grid: { drawOnChartArea: false }
          },
          yPos: {
            type: 'linear',
            display: 'auto',
            position: 'right',
            title: { display: true, text: 'Posição' },
            reverse: true, // Google Search Console inverts the position axis (1 is at top)
            min: 1,
            grid: { drawOnChartArea: false }
          }
        },
        plugins: {
          legend: {
            display: false // We will use custom HTML legend
          },
          tooltip: {
            mode: 'index',
            intersect: false,
            callbacks: {
              label: function(context) {
                let label = context.dataset.label || '';
                if (label) {
                  label += ': ';
                }
                if (context.parsed.y !== null) {
                  if (context.datasetIndex === 2) { // CTR
                    label += context.parsed.y.toFixed(2) + '%';
                  } else if (context.datasetIndex === 3) { // Posição
                    label += context.parsed.y.toFixed(1);
                  } else {
                    label += new Intl.NumberFormat('pt-BR').format(context.parsed.y);
                  }
                }
                return label;
              }
            }
          }
        }
      }
    });
  }

  $effect(() => {
    if (data && canvas) {
      initChart();
    }
  });

  onDestroy(() => {
    if (chart) chart.destroy();
  });
</script>

<div class="gsc-chart-wrapper">
  <div class="gsc-chart-toggles">
    <button 
      class="gsc-toggle" 
      class:active={showClicks} 
      style="--toggle-color: #4285F4;"
      onclick={() => showClicks = !showClicks}>
      <span class="gsc-toggle-box"></span>
      Cliques
    </button>
    <button 
      class="gsc-toggle" 
      class:active={showImpressions} 
      style="--toggle-color: #9E9E9E;"
      onclick={() => showImpressions = !showImpressions}>
      <span class="gsc-toggle-box"></span>
      Impressões
    </button>
    <button 
      class="gsc-toggle" 
      class:active={showCtr} 
      style="--toggle-color: #0f9d58;"
      onclick={() => showCtr = !showCtr}>
      <span class="gsc-toggle-box"></span>
      CTR
    </button>
    <button 
      class="gsc-toggle" 
      class:active={showPosition} 
      style="--toggle-color: #ab47bc;"
      onclick={() => showPosition = !showPosition}>
      <span class="gsc-toggle-box"></span>
      Posição
    </button>
  </div>

  <div class="chart-container" style="position: relative; height: 300px; width: 100%; margin-bottom: var(--s-6);">
    <canvas bind:this={canvas}></canvas>
  </div>
</div>

<style>
  .gsc-chart-wrapper {
    display: flex;
    flex-direction: column;
    gap: var(--s-4);
  }

  .gsc-chart-toggles {
    display: flex;
    justify-content: center;
    gap: var(--s-4);
    flex-wrap: wrap;
    margin-bottom: var(--s-2);
  }

  .gsc-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 8px 16px;
    font-size: 13px;
    font-weight: 500;
    color: var(--fg-muted);
    cursor: pointer;
    transition: all 0.2s;
  }

  .gsc-toggle:hover {
    background: var(--bg-subtle);
  }

  .gsc-toggle-box {
    width: 14px;
    height: 14px;
    border-radius: 2px;
    background: transparent;
    border: 2px solid var(--toggle-color);
    transition: all 0.2s;
  }

  .gsc-toggle.active {
    color: var(--fg-default);
    background: var(--bg-elevated);
    box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  }

  .gsc-toggle.active .gsc-toggle-box {
    background: var(--toggle-color);
  }
</style>
