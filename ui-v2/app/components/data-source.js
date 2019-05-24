import Component from '@ember/component';
import { inject as service } from '@ember/service';
import { get, set, computed } from '@ember/object';

import WithListeners from 'consul-ui/mixins/with-listeners';

const replace = function(
  obj,
  prop,
  value,
  destroy = (prev = null, value) => (typeof prev === 'function' ? prev() : null)
) {
  const prev = get(obj, prop);
  if (prev !== value) {
    destroy(prev, value);
  }
  return set(obj, prop, value);
};

export default Component.extend(WithListeners, {
  tagName: 'span',
  // TODO: Temporary repo list here
  service: service('repository/service'),
  node: service('repository/node'),
  session: service('repository/session'),

  blocking: service('blocking'),
  settings: service('settings'),
  onmessage: function() {},
  onerror: function() {},
  onprogress: function() {},
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
  didInsertElement: function() {
    this._super(...arguments);
    const options = {
      rootMargin: '0px',
      threshold: 1.0,
    };

    // const source = new InViewportEventSource(this.element);
    const observer = new IntersectionObserver((entries, observer) => {
      entries.map(item => {
        set(this, 'isIntersecting', item.isIntersecting);
        if (!item.isIntersecting) {
          this.actions.close.bind(this)();
        } else {
          this.actions.open.bind(this)();
        }
      });
    }, options);
    observer.observe(this.element);
    this.listen(() => {
      this.actions.close.bind(this)();
      observer.disconnect();
    });
  },
  didReceiveAttrs: function() {
    this._super(...arguments);
    if (this.element && get(this, 'isIntersecting')) {
      this.actions.open.bind(this)();
    }
  },
  actions: {
    open: function() {
      // keep this argumentless
      const src = get(this, 'src');
      const filter = get(this, 'filter');

      replace(
        this,
        'source',
        get(this, 'blocking').open(this.finder(src, filter), {
          id: `${src}${filter ? `?filter=${filter}` : ``}`,
        }),
        (prev, source) => {
          get(this, 'blocking').close(prev);
          const remove = this.listen(source, {
            message: e => this.onmessage(e),
            error: e => {
              remove();
              this.onerror(e);
            },
          });
          replace(this, '_remove', remove);
          const previousEvent = source.getPreviousEvent();
          if (previousEvent) {
            source.dispatchEvent(previousEvent);
          }
        }
      );
    },
    close: function() {
      // keep this argumentless
      get(this, 'blocking').close(get(this, 'source'));
      replace(this, '_remove', null);
      set(this, 'source', undefined);
    },
  },
});
