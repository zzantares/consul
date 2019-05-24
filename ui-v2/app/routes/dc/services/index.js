import Route from '@ember/routing/route';
import { get } from '@ember/object';

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
    const repo = get(this, 'repo');
    let s = params.s;
    // we check for the old style `status` variable here
    // and convert it to the new style filter=status:critical
    let status = params.status;
    if (status) {
      status = `status:${status}`;
      if (s && s.indexOf(status) === -1) {
        s = s
          .split('\n')
          .concat(status)
          .join('\n')
          .trim();
      }
    }
    return {
      s: s,
      slug: '*',
      dc: this.modelFor('dc').dc.Name,
    };
  },
  setupController: function(controller, model) {
    controller.setProperties(model);
  },
});
