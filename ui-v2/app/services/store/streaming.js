import BlockingEventSource from 'consul-ui/utils/DSBlockingEventSource';
import Service, { inject as service } from '@ember/service';

import { get } from '@ember/object';

export default Service.extend({
  store: service('store'),
  settings: service('settings'),
  init: function() {
    this._super(...arguments);
    this.sources = [];
  },
  query: function(name, query) {
    return get(this, 'settings')
      .findBySlug('autorefresh')
      .then(autorefresh => {
        const store = get(this, 'store');
        if (autorefresh) {
          const adapter = store.adapterFor(name);
          const url = adapter.urlForQuery(query);
          if (typeof this.sources[name] === 'undefined') {
            this.sources[name] = {};
          }
          const sources = this.sources[name];
          sources[url] = new BlockingEventSource(url, {
            store: store,
            name: name,
            index: sources[url],
          });
          // this will do a one off streamer, equivalent to store.query etc
          // ready for later instead of `fallback`
          // queries[url].addEventListener(
          //   'complete',
          //   function()
          //   {
          //     this.close();
          //   }
          // );
          return store.peekAll(name);
        } else {
          return store.query(name, query);
        }
      });
  },
  queryRecord: function() {
    return get(this, 'store').queryRecord(...arguments);
  },
  abortAll: function() {},
  abort: function(name) {
    const sources = this.sources[name] || {};
    Object.keys(sources).forEach(item => {
      sources[item] = sources[item].close();
    });
  },
});
