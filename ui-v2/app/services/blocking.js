import Service, { inject as service } from '@ember/service';
import { get, set } from '@ember/object';
import { BlockingEventSource } from 'consul-ui/utils/dom/event-source';
import LRUMap from 'npm:mnemonist/lru-map';

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
  return what
    ? function(res) {
        cb();
        return res;
      }
    : res => res;
};
const ifNotBlocking = function(settings) {
  return settings.findBySlug('client').then(res => !res.blocking);
};

// TODO: Expose this as a env var
const messages = new LRUMap(50);
// old cursors are never deleted
const cursors = new Map();
// sources are 'manually' removed when closeed,
// they are only closed when the usage counter is 0
const sources = new Map();

const makeCounter = function() {
  const count = new WeakMap();
  return {
    add: function(obj) {
      const num = (count.get(obj) || 0) + 1;
      count.set(obj, num);
      return num;
    },
    remove: function(obj) {
      let num = count.get(obj) || 0;
      if (num) {
        num = num - 1;
        count.set(obj, num);
      }
      return Math.max(num, 0);
    },
  };
};
let counter;
export default Service.extend({
  client: service('client/http'),
  settings: service('settings'),
  init: function() {
    // counter = get('datastruct').counter();
    counter = makeCounter();
  },
  open: function(cb, configuration) {
    const id = configuration.id;
    if (!sources.has(id)) {
      // TODO: if something is filtered we shouldn't reconcile things
      const source = new BlockingEventSource(
        (configuration, source) => {
          const close = source.close.bind(source);
          return cb(configuration)
            .then(maybeCall(close, ifNotBlocking(get(this, 'settings'))))
            .catch(restartWhenAvailable(get(this, 'client')));
        },
        {
          cursor: cursors.get(id),
        }
      );
      sources.set(id, source);
      // keep this order do we aren't notified of any close events here
      // source.close();
      const previousEvent = messages.get(id);
      const open = function(e) {
        if (previousEvent) {
          e.target.dispatchEvent(previousEvent);
        }
      };
      source.addEventListener('open', open);
      source.addEventListener('close', function close(e) {
        const source = e.target;
        source.removeEventListener('close', close);
        source.removeEventListener('open', open);
        messages.set(id, source.getCurrentEvent());
        cursors.set(id, source.configuration.cursor);
        sources.delete(id);
      });
      //
    }
    const source = sources.get(id);
    counter.add(source);
    source.open();
    return source;
  },
  close: function(source) {
    if (source) {
      const num = counter.remove(source);
      if (num === 0) {
        source.close();
      }
    }
  },
});
