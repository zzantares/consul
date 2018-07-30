import Service from '@ember/service';
import { Promise } from 'rsvp';
import { get } from '@ember/object';

export default Service.extend({
  // TODO: change name
  storage: window.localStorage,
  findHeaders: function() {
    // TODO: if possible this should be a promise
    const token = get(this, 'storage').getItem('token');
    // TODO: The old UI always sent ?token=
    // replicate the old functionality here
    // but remove this to be cleaner if its not necessary
    return {
      'X-Consul-Token': token === null ? '' : token,
    };
  },
  findAll: function(key) {
    const token = get(this, 'storage').getItem('token');
    const autorefresh = get(this, 'storage').getItem('autorefresh');
    return Promise.resolve({
      token: token === null ? '' : token,
      autorefresh: Boolean(
        autorefresh === null ? '' : autorefresh === 'false' ? false : autorefresh
      ),
    });
  },
  findBySlug: function(slug) {
    // TODO: Force localStorage to always be strings...
    // const value = get(this, 'storage').getItem(slug);
    const item = get(this, 'storage').getItem(slug);
    return Promise.resolve(item === 'false' ? false : item);
  },
  persist: function(obj) {
    const storage = get(this, 'storage');
    Object.keys(obj).forEach((item, i) => {
      // TODO: ...everywhere
      storage.setItem(item, obj[item]);
    });
    return Promise.resolve(obj);
  },
  delete: function(obj) {
    return Promise.resolve(get(this, 'storage').removeItem('token'));
  },
});
