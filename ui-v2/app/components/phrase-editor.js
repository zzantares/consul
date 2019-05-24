import Component from '@ember/component';
import { get, set, computed } from '@ember/object';
import { inject as service } from '@ember/service';

export default Component.extend({
  classNames: ['phrase-editor'],
  dom: service('dom'),
  item: '',
  onchange: function(e) {},
  oninput: function(e) {},
  onkeydown: function(e) {},
  actions: {
    keydown: function(e) {
      switch (e.keyCode) {
        case 8: // backspace
          if (e.target.value == '' && get(this, 'value').length > 0) {
            this.actions.remove.bind(this)(get(this, 'value').length - 1);
          }
          break;
        case 27: // escape
          set(this, 'value', []);
          get(this, 'dom').dispatch(this.element, { type: 'change', target: this }, e);
          break;
      }
      get(this, 'dom').dispatch(this.element, { type: 'keydown', target: this }, e);
    },
    input: function(e) {
      set(this, 'item', e.target.value);
      get(this, 'dom').dispatch(this.element, { type: 'input', target: this }, e);
    },
    remove: function(index, e) {
      get(this, 'value').removeAt(index, 1);
      get(this, 'dom').dispatch(this.element, { type: 'change', target: this }, e);
    },
    add: function(e) {
      const item = get(this, 'item').trim();
      if (item !== '') {
        set(this, 'item', '');
        const currentItems = get(this, 'value') || [];
        const items = new Set(currentItems).add(item);
        if (items.size > currentItems.length) {
          set(this, 'value', [...items]);
          get(this, 'dom').dispatch(this.element, { type: 'change', target: this }, e);
        }
      }
    },
  },
});
