// Simple RSVP.EventTarget wrapper to make it more like a standard EventTarget
import RSVP from 'rsvp';
const EventTarget = function() {};
EventTarget.prototype = Object.assign(
  Object.create(Object.prototype, {
    constructor: {
      value: EventTarget,
      configurable: true,
      writable: true,
    },
  }),
  {
    dispatchEvent: function(obj) {
      this.trigger(obj.type, { data: obj.data });
    },
    addEventListener: function(event, cb) {
      this.on(event, cb);
    },
    removeEventListener: function(event, cb) {
      this.off(event, cb);
    },
  }
);
RSVP.EventTarget.mixin(EventTarget.prototype);
export default EventTarget;
