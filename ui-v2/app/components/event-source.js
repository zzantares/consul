import Component from '@ember/component';
import { get, set } from '@ember/object';
import { inject as service } from '@ember/service';

import { CallableEventSource as EventSource } from 'consul-ui/utils/dom/event-source';

import WithListeners from 'consul-ui/mixins/with-listeners';

let uuid = 0;
export default Component.extend(WithListeners, {
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
    const source = get(this, 'src');
    if (source) {
      set(this, '_src', null);
      const $placeholder = get(this, 'dom').element(`#${get(this, 'uuid')}`);
      let $previous = $placeholder.previousElementSibling;
      // TODO: Find a nicer way to figure if something is a component or not
      // if($previous.getAttribute('id').startsWith('ember-')) {
      //   const component = get(this, 'dom').component($previous);
      //   // for the moment, just incase a non-component element has an
      //   // id starting with ember-
      //   if(component) {
      //     $previous = component;
      //   }
      // }
      set(this, 'target', $previous);
      $placeholder.remove();
      this.listen(source, 'message', e => {
        const type = e.data.type;
        if (get(this, 'type') == type) {
          set(this, `on${type}`, get(this, 'handler'));
        }
        get(this, 'dom').dispatch(get(this, 'target'), { type: type, target: this });
      });
      this.listen(source, 'error', e => this.onerror(e));
    }
  },
  actions: {
    dispatch: function() {
      let e, type;
      let args = [];
      [...arguments].forEach(function(item, i) {
        switch (true) {
          case i === 0 && typeof item === 'string':
            type = item;
            break;
          case item instanceof Event:
          case typeof item.type === 'string':
            e = item;
            break;
          default:
            args.push(item);
            break;
        }
      });
      get(this, '_src').dispatchEvent({ type: 'message', data: { type: type, data: e } });
    },
    call: function() {
      const args = [...arguments];
      const event = args.pop();
      const method = args.shift();
      event.target[method](...args);
      // TODO: Maybe allow paths to be specified (action target 'something.method')
      // const args = [...arguments];
      // const event = args.pop();
      // const path = args.shift();
      // const temp = path.split('.');
      // const method = temp.pop();
      // const obj = get(event.target, temp.join('.'));
      // obj[method].bind(obj)('selected')
    },
    message: function(data) {
      get(this, '_src').dispatchEvent({ type: 'message', data: data });
    },
    error: function(e) {
      get(this, '_src').dispatchEvent({ type: 'error', error: e });
    },
  },
});
