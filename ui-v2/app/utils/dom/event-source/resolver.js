export default function(P = Promise) {
  return function(source, listeners) {
    return new P(function(resolve, reject) {
      // close, cleanup and reject if we get an error
      listeners.add(source, 'error', function(e) {
        listeners.remove();
        e.target.close();
        reject(e.error);
      });
      // ...or cleanup and respond with the first lot of data
      listeners.add(source, 'message', function(e) {
        listeners.remove();
        resolve(e.data);
      });
      source.open();
    });
  };
}
