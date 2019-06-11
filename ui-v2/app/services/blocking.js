import Service, { inject as service } from '@ember/service';
import { get } from '@ember/object';
import { BlockingEventSource } from 'consul-ui/utils/dom/event-source';
import LRUMap from 'mnemonist/lru-map';
import MultiMap from 'mnemonist/multi-map';

const restartWhenAvailable = function(client) {
  return function(e) {
    // setup the aborted connection restarting
    // this should happen here to avoid cache deletion
    const status = get(e, 'errors.firstObject.status');
    if (status === '0') {
      // Any '0' errors (abort) should possibly try again, depending upon the circumstances
      // whenAvailable returns a Promise that resolves when the client is available
      // again
      return client.whenAvailable(e);
    }
    throw e;
  };
};
const maybeCall = function(cb, what) {
  return function(res) {
    return what.then(function(bool) {
      if (bool) {
        cb();
      }
      return res;
    });
  };
};
const ifNotBlocking = function(settings) {
  return settings.findBySlug('client').then(res => !res.blocking);
};

// TODO: Expose this as a env var
const cache = new Map();
// sources are 'manually' removed when closed,
// they are only closed when the usage counter is 0
const sources = new Map();
const refs = new MultiMap(Set);

export default Service.extend({
  // TODO: Temporary repo list here
  service: service('repository/service'),
  node: service('repository/node'),
  session: service('repository/session'),

  client: service('client/http'),
  settings: service('settings'),
  // TODO: Temporary finder
  finder: function(src, filter) {
    const temp = src.split('/');
    temp.shift();
    const slug = temp.pop();
    const model = temp.pop();
    const dc = temp.shift();

    switch (slug) {
      case '*':
        switch (model) {
          default:
            return configuration => {
              return get(this, model).findAllByDatacenter(dc, {
                cursor: configuration.cursor,
                filter: filter,
              });
            };
        }
      default:
        switch (model) {
          case 'session':
            return configuration => {
              return get(this, model).findByNode(slug, dc, { cursor: configuration.cursor });
            };
          default:
            return configuration => {
              return get(this, model).findBySlug(slug, dc, { cursor: configuration.cursor });
            };
        }
    }
  },
  open: function(uri, ref) {
    if (!sources.has(uri)) {
      const cb = this.finder.apply(this, uri.split('?filter='));
      let configuration = {};
      if (cache.has(uri)) {
        configuration = cache.get(uri);
      }
      // TODO: if something is filtered we shouldn't reconcile things
      const source = new BlockingEventSource((configuration, source) => {
        const close = source.close.bind(source);
        return cb(configuration)
          .then(maybeCall(close, ifNotBlocking(get(this, 'settings'))))
          .catch(restartWhenAvailable(get(this, 'client')));
      }, configuration);
      source.addEventListener('close', function close(e) {
        const source = e.target;
        source.removeEventListener('close', close);
        cache.set(uri, {
          currentEvent: e.target.getCurrentEvent(),
          cursor: e.target.configuration.cursor,
        });
        sources.delete(uri);
      });
      sources.set(uri, source);
    }
    const source = sources.get(uri);
    refs.set(source, ref);
    source.open();
    return source;
  },
  close: function(source, ref) {
    if (source) {
      refs.remove(source, ref);
      if (!refs.has(source)) {
        source.close();
      }
    }
  },
});
