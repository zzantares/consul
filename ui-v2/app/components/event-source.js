import Component from '@ember/component';
import { get, set } from '@ember/object';
import { inject as service } from '@ember/service';

import { CallableEventSource as EventSource } from 'consul-ui/utils/dom/event-source';

let uuid = 0;
export default Component.extend({
  tagName: '',
  dom: service('dom'),
  onmessage: function() {},
  onerror: function() {},
  init: function() {
    this._super(...arguments);
    set(this, 'uuid', `event-source-${uuid++}`);
    set(this, '_src', new EventSource());
  },
  didInsertElement: function() {
    this._super(...arguments);
    const source = get(this, 'src');
    if (source) {
      set(this, '_src', null);
      if (!get(this, 'ref')) {
        const $placeholder = get(this, 'dom').element(`#${get(this, 'uuid')}`);
        set(this, 'ref', $placeholder.previousElementSibling);
      }
      this._listeners = get(this, 'dom').listeners();
      this._listeners.add(source, 'message', e => {
        const type = e.data.type;
        this[`on${type}`]({ target: get(this, 'ref') });
      });
      this._listeners.add(source, 'error', e => this.onerror(e));
    }
  },
  willDestroyElement: function() {
    this._super(...arguments);
    if (this._listeners) {
      this._listeners.remove();
    }
  },
  actions: {
    dispatch: function(eventName, e) {
      get(this, '_src').dispatchEvent({ type: 'message', data: { type: eventName, data: e } });
    },
    call: function() {
      const args = [...arguments];
      const event = args.pop();
      const method = args.shift();
      event.target[method](...args);
    },
    message: function(data) {
      get(this, '_src').dispatchEvent({ type: 'message', data: data });
    },
    error: function(e) {
      get(this, '_src').dispatchEvent({ type: 'error', error: e });
    },
  },
});
