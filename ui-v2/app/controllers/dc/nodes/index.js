import Controller from '@ember/controller';
import { set } from '@ember/object';
import { get } from '@ember/object';
let s;
export default Controller.extend({
  setProperties: function(model) {
    s = model.s = typeof model.s === 'undefined' ? s : model.s;
    if (s) {
      set(this, 'terms', s.split('\n'));
    } else {
      set(this, 'terms', []);
    }
    return this._super(model);
  },
  actions: {
    query: function(args) {
      s = args.length > 0 ? args.join('\n') : null;
      set(this, 's', s);
    },
  },
});
