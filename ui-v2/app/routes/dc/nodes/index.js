import Route from '@ember/routing/route';

export default Route.extend({
  queryParams: {
    s: {
      as: 'filter',
      replace: true,
    },
  },
  model: function(params) {
    return {
      s: params.s,
      dc: this.modelFor('dc').dc.Name,
      slug: '*',
    };
  },
  setupController: function(controller, model) {
    controller.setProperties(model);
  },
});
