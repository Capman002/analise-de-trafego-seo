// Tipos TypeScript compartilhados entre componentes — espelham os models Go.

export interface Client {
  id: number;
  name: string;
  gscUrl: string;
  gscType: 'domain' | 'url' | '';
  ga4Id: string;
  brandRegex: string;
  syncedAt: string;
  permissionError: boolean;
}

export interface TrafficOverview {
  gsc_clicks: number;
  gsc_impressions: number;
  gsc_ctr: number;
  gsc_position: number;
  ga4_sessions: number;
  ga4_revenue: number;
  ga4_items_purchased: number;
}

export interface SyncStatus {
  [source: string]: {
    status: 'success' | 'error' | 'cached' | 'skip';
    error: string;
    last_sync_at: string;
  };
}

export interface PositionDistribution {
  top3: number;
  top10: number;
  top20: number;
  beyond20: number;
  prev_top3: number;
  prev_top10: number;
  prev_top20: number;
  prev_beyond20: number;
}

export interface TrafficResponse {
  client: Client;
  period: number;
  startDate?: string;
  endDate?: string;
  overview: TrafficOverview;
  overview_prev: TrafficOverview;
  position_distribution: PositionDistribution;
  gsc_queries: any[];
  gsc_pages: any[];
  gsc_chart: any[];
  gsc_chart_prev: any[];
  gsc_rise_queries: any[];
  gsc_rise_pages: any[];
  gsc_fall_queries: any[];
  gsc_fall_pages: any[];
  ga4: any[];
  bing: any[];
  sync_status: SyncStatus;
}
