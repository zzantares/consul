import Route from '@ember/routing/route';
import convertStatus from 'consul-ui/utils/routing/convert-status';

export default Route.extend({
  queryParams: {
    s: {
      as: 'filter',
      replace: true,
    },
    // temporary support of old style status
    status: {
      as: 'status',
    },
  },
  model: function(params) {
    return {
      // we check for the old style `status` variable here
      // and convert it to the new style filter=status:critical
      s: convertStatus(params.s, params.status),
      slug: '*',
      dc: this.modelFor('dc').dc.Name,
    };
  },
  setupController: function(controller, model) {
    controller.setProperties(model);
  },
});
