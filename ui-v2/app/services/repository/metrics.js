import RepositoryService from 'consul-ui/services/repository';
import { inject as service } from '@ember/service';

export default RepositoryService.extend({
  env: service('env'),

  get provider() {
    return window.consul.getMetricsProvider(this.env.var('CONSUL_METRICS_PROVIDER'), {
      ...this.env.var('CONSUL_METRICS_PROVIDER_OPTIONS'),
      metrics_proxy_enabled: this.env.var('CONSUL_METRICS_PROXY_ENABLED'),
    });
  },

  findDashboardURLs: function() {
    return Promise.resolve(this.env.var('CONSUL_DASHBOARD_URL_TEMPLATES'));
  },

  findServiceSummary: function(protocol, slug, dc, nspace, configuration = {}) {
    const promises = [
      // TODO: support namespaces in providers
      this.provider.serviceRecentSummarySeries(slug, protocol, {}),
      this.provider.serviceRecentSummaryStats(slug, protocol, {}),
    ];
    return Promise.all(promises).then(results => {
      return {
        meta: {
          interval: this.env.var('CONSUL_METRICS_REFRESH_INTERVAL'),
        },
        series: results[0].series,
        stats: results[1].stats,
      };
    });
  },

  findUpstreamSummary: function(slug, dc, nspace, configuration = {}) {
    return this.provider.upstreamRecentSummaryStats(slug, {}).then(result => {
      result.meta = {
        interval: this.env.var('CONSUL_METRICS_REFRESH_INTERVAL'),
      };
      return result;
    });
  },

  findDownstreamSummary: function(slug, dc, nspace, configuration = {}) {
    return this.provider.downstreamRecentSummaryStats(slug, {}).then(result => {
      result.meta = {
        interval: this.env.var('CONSUL_METRICS_REFRESH_INTERVAL'),
      };
      return result;
    });
  },
});
