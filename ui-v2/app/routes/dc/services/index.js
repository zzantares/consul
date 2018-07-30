import Route from '@ember/routing/route';
import { inject as service } from '@ember/service';
import { hash } from 'rsvp';
import { get } from '@ember/object';

export default Route.extend({
  repo: service('services'),
  queryParams: {
    s: {
      as: 'filter',
      replace: true,
    },
  },
  model: function(params) {
    const repo = get(this, 'repo');
    return hash({
      items: repo.findAllByDatacenter(this.modelFor('dc').dc.Name),
    });
  },
  setupController: function(controller, model) {
    this._super(...arguments);
    controller.setProperties(model);
  },
  deactivate: function() {
    const repo = get(this, 'repo');
    repo.abortAll();
  },
  // TODO: willTransition seemed to be the hook to use
  // but if you abort here then tests aren't cleaned up properly?
  // willTransition: function(transition) {
  //   const repo = get(this, 'repo');
  //   repo.abortAll();
  // },
});
