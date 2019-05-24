import { helper } from '@ember/component/helper';
import ucfirst from 'consul-ui/utils/ucfirst';
const convert = function(str, map) {
  const replacement = map.find(function(arr) {
    const key = arr[0];
    const val = arr[1];
    return str.startsWith(key);
  });
  const replaced = str.replace(replacement[0], '');
  return replacement[1].replace('%s', replaced).replace('%S', ucfirst(replaced));
};
export function toFilter([values], attrs) {
  return values
    .map(function(item) {
      return convert(item, attrs.map);
    }, '')
    .join(' AND ');
}

export default helper(toFilter);
