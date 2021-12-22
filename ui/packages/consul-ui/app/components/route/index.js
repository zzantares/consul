import Component from '@glimmer/component';
import { inject as service } from '@ember/service';
import { get, action } from '@ember/object';
import { tracked } from '@glimmer/tracking';
import HTTPError from 'consul-ui/utils/http/error';
import { routes } from 'consul-ui/router';

export default class RouteComponent extends Component {
  @service('routlet') routlet;
  @service('router') router;
  @service('repository/permission') permissions;

  @tracked error;
  @tracked _model;

  get params() {
    return this.routlet.paramsFor(this.args.name);
  }

  get model() {
    if(this._model) {
      return this._model;
    }
    if (this.args.name) {
      const outlet = this.routlet.outletFor(this.args.name);
      return this.routlet.modelFor(outlet.name);
    }
    return;
  }

  /**
   * Inspects a custom `abilities` array on the router for this route. Every
   * ability needs to 'pass' for the route not to throw a 403 error. Anything
   * more complex then this (say ORs) should use a single ability and perform
   * the OR logic in the test for the ability. Note, this ability check happens
   * before any calls to the backend for this model/route.
   */
  authorize(routeName, routes) {
    const abilities = get(routes, `${routeName}._options.abilities`) || [];
    if (abilities.length > 0) {
      if (!abilities.every(ability => this.permissions.can(ability))) {
        this.error = new HTTPError('403');
      }
    }
  }

  @action
  connect() {
    this.authorize(this.args.name, routes);
    this.routlet.addRoute(this.args.name, this);
  }

  @action
  disconnect() {
    this.routlet.removeRoute(this.args.name, this);
  }
}
