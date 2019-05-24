import { helper } from '@ember/component/helper';

export function isUndefined([value] /*, hash*/) {
  let actualValue = value;
  return typeof actualValue === 'undefined';
}

export default helper(isUndefined);
