import { run } from '@ember/runloop';
import EventTarget from 'consul-ui/utils/RSVPEventTarget';
import CallableEventSourceFactory from 'npm:@hashicorp/callable-event-source';
import BlockingEventSourceFactory from 'npm:@hashicorp/blocking-event-source';
const CallableEventSource = CallableEventSourceFactory(EventTarget);
const BlockingEventSource = BlockingEventSourceFactory(CallableEventSource);
import URL from 'url';

const defaultGetRequest = function(request, response, configuration) {
  let index = '';
  const headers = response.headers || { 'x-consul-index': configuration.index };
  const lower = {};
  Object.keys(headers).forEach(function(key) {
    lower[key.toLowerCase()] = headers[key];
  });
  if (lower['x-consul-index'] != null) {
    const source = request.url;
    const sep = source.indexOf('?') === -1 ? '?' : '&';
    index = lower['x-consul-index'];
    request = Object.assign({}, request, {
      url: index > 0 ? `${source}${sep}index=${index}` : '',
    });
  }
  return request;
};

const DSBlockingQuery = function(source, configuration = {}) {
  configuration.getRequest =
    typeof configuration.getRequest === 'function' ? configuration.getRequest : defaultGetRequest;
  BlockingEventSource.apply(this, [source, configuration]);

  const name = configuration.name;
  const store = configuration.store;
  const adapter = store.adapterFor(name);
  const serializer = store.serializerFor(name);
  const Model = store.modelFor(name);

  let previousIds = [];
  let ids = [];
  this.onmessage = function(e) {
    const response = adapter.handleResponse(200, {}, [e.data], { url: source })[0];
    const item = serializer.normalize(Model, response);
    ids.push(item.data.id);
    run(function() {
      try {
        store.push(item);
      } catch (e) {
        // console.error(e);
      }
    });
    // this.dispatchEvent({type: 'load', data: item});
  };
  this.oncomplete = function(e) {
    previousIds.forEach(id => {
      const contains = ids.includes(id);
      if (!contains) {
        try {
          const item = store.peekRecord(name, id);
          store.unloadRecord(item);
        } catch (e) {
          // console.error(e);
        }
        // this.dispatchEvent({type: 'unload', data: item});
      }
    });
    previousIds = ids;
    ids = [];
  };
  this.addEventListener('message', this.onmessage);
  this.addEventListener('complete', this.oncomplete);
};
DSBlockingQuery.prototype = Object.assign(
  Object.create(BlockingEventSource.prototype, {
    constructor: {
      value: DSBlockingQuery,
      configurable: true,
      writable: true,
    },
  }),
  {
    close: function() {
      this.removeEventListener('message', this.onmessage);
      this.removeEventListener('complete', this.onclose);
      const request = BlockingEventSource.prototype.close.apply(this, [arguments]);
      const index = parseInt(
        new URL(request.url, `${location.protocol}//${location.host}`).searchParams.get('index')
      );
      return index + 1;
    },
  }
);
export default DSBlockingQuery;
