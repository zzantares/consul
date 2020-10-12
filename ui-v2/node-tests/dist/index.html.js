const test = require('tape');
const read = require('fs').readFileSync;
const path = require('path');


test(
  'index.html has required go template interpolations',
  function(t) {
    const index = read(path.resolve(__dirname, '../../dist/index.html')).toString();
    t.notEquals(index.indexOf('{{.ContentPath}}'), -1, 'index.html contains ContentPath');
    t.notEquals(index.indexOf('{{ jsonEncodeAndEscape .UIConfig }}'), -1, 'index.html contains UIConfig');
    t.notEquals(index.indexOf('{{ range .ExtraScripts }}'), -1, 'index.html contains ExtraScripts');
    t.end();
  }
);
