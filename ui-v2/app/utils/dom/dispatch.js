export default function(EventClass = CustomEvent) {
  return function dispatch(eventTarget, desc, originalEvent) {
    const target = desc.target || eventTarget;
    const type = `event-source.${desc.type}`;
    const event = new EventClass(type, {
      // TODO: this should probably be detail.target
      detail: target,
    });
    const cb = target[`on${desc.type}`];
    // TODO: Should we fire the event anyway, maybe something else
    // is listening?
    if (typeof cb === 'function') {
      // if eventTarget is a component
      // maybe we can just call without firing?
      const listener = cb.bind(target);
      eventTarget.addEventListener(type, listener);
      eventTarget.dispatchEvent(event);
      eventTarget.removeEventListener(type, listener);
    }
  };
}
